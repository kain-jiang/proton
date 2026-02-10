/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/base64"
	"fmt"
	baseresource "proton-mongodb-operator/controllers/builtinresource"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/go-logr/logr"
	pkgerr "github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	mongodbv1 "proton-mongodb-operator/api/v1"
)

// MongodbOperatorReconciler reconciles a MongodbOperator object
type MongodbOperatorReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Clientcmd *Client
	lockers   lockStore
}

// newReconciler returns a new reconcile.Reconciler
func NewReconciler(mgr manager.Manager) (*MongodbOperatorReconciler, error) {
	// sv, err := version.Server()
	// if err != nil {
	// 	return nil, errors.Wrap(err, "get server version")
	// }

	// log.Info("server version", "platform", sv.Platform, "version", sv.Info)

	cli, err := NewClient()
	if err != nil {
		return nil, pkgerr.Wrap(err, "create clientcmd")
	}

	return &MongodbOperatorReconciler{
		mgr.GetClient(),
		mgr.GetScheme(),
		cli,
		newLockStore(),

		// clientcmd: cli,
	}, nil
}

type lockStore struct {
	store *sync.Map
}

func newLockStore() lockStore {
	return lockStore{
		store: new(sync.Map),
	}
}

func (l lockStore) LoadOrCreate(key string) lock {
	val, _ := l.store.LoadOrStore(key, lock{
		statusMutex: new(sync.Mutex),
		updateSync:  new(int32),
	})

	return val.(lock)
}

type lock struct {
	statusMutex *sync.Mutex
	updateSync  *int32
}

const (
	updateDone = 0
	updateWait = 1
)

//+kubebuilder:rbac:groups=mongodb.proton.aishu.cn,resources=mongodboperators,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mongodb.proton.aishu.cn,resources=mongodboperators/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mongodb.proton.aishu.cn,resources=mongodboperators/finalizers,verbs=update
//+kubebuilder:rbac:groups="";apps;batch;rbac.authorization.k8s.io,resources=configmaps;deployments;persistentvolumeclaims;persistentvolumes;pods;pods/exec;secrets;services;statefulsets;cronjobs;jobs;serviceaccounts;clusterroles;clusterrolebindings,verbs=create;delete;get;list;watch;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MongodbOperator object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *MongodbOperatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// TODO(user): your logic here
	l := log.FromContext(ctx).WithName("proton-mongodb-controller").WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	lock := r.lockers.LoadOrCreate(req.NamespacedName.String())
	lock.statusMutex.Lock()
	defer lock.statusMutex.Unlock()
	defer atomic.StoreInt32(lock.updateSync, updateDone)
	//get MongodbOerator object
	instance := &mongodbv1.MongodbOperator{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		l.Error(err, "reconcile get cr error")
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	l.Info("reconcile begin")
	// TODO(user): your logic here

	if !instance.ObjectMeta.DeletionTimestamp.IsZero() {
		instance.GetFinalizers()
	}
	myFinalizerName := "delete-proton-mongodb-pv/pvc"

	if instance.Spec.MongoDBSpec.Storage.StorageClassName != "" {
		if instance.ObjectMeta.DeletionTimestamp.IsZero() {
			// The object is not being deleted, so if it does not have our finalizer,
			// then add the finalizer and update the object. This is equivalent
			// registering our finalizer.
			if !controllerutil.ContainsFinalizer(instance, myFinalizerName) {
				controllerutil.AddFinalizer(instance, myFinalizerName)
				if err := r.Client.Update(ctx, instance); err != nil {
					return ctrl.Result{}, err
				}
			}
		} else {
			// The object is being deleted
			if controllerutil.ContainsFinalizer(instance, myFinalizerName) {
				// our finalizer is present, so lets handle any external dependency
				if err := r.deleteExternalResources(ctx, instance, l); err != nil {
					return ctrl.Result{}, err
				}
				controllerutil.RemoveFinalizer(instance, myFinalizerName)
				if err := r.Client.Update(ctx, instance); err != nil {
					return ctrl.Result{}, err
				}
			}

			// Stop reconciliation as the item is being deleted
			return ctrl.Result{}, nil
		}
	}

	err = r.NewMongodb(ctx, instance, l)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.NewMgmt(ctx, instance, l)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.NewExporter(ctx, instance, l)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.updateMongoReplsetStatus(ctx, instance, l)
	if err != nil {
		if errors.IsConflict(err) {
			return reconcile.Result{Requeue: true}, nil
		} else {
			l.Error(pkgerr.Wrap(err, "unexpected error when update mongoReplset status"), "update status fail")
		}
		return reconcile.Result{}, err
	}
	// if instance.Spec.MongoDBSpec.Resources.Limits.Memory().String() != "0" {
	// 	//
	// }
	return ctrl.Result{}, nil
}

func (r *MongodbOperatorReconciler) deleteExternalResources(ctx context.Context, instance *mongodbv1.MongodbOperator, l logr.Logger) error {
	//delete pvc
	for postdel := 0; postdel < 3; postdel++ {
		delpvc := &corev1.PersistentVolumeClaim{}
		key := types.NamespacedName{
			Name:      fmt.Sprintf("%s-%s-"+strconv.Itoa(int(postdel)), "mongodb-datadir", fmt.Sprintf("%s-%s", instance.GetName(), "mongodb")),
			Namespace: instance.Namespace,
		}
		if err := r.Get(ctx, key, delpvc); err != nil {
			if errors.IsNotFound(err) {
				l.Error(pkgerr.Wrap(err, "not found, pvc is post-deleted"), key.Name+" in "+key.Namespace+"not found")
			} else {
				l.Error(pkgerr.Wrap(err, "get delpvc error"), key.Name+" in "+key.Namespace+"get delpvc error")
				return pkgerr.Wrap(err, "get delpvc error")
			}
		} else {
			if err := r.Delete(ctx, delpvc); err != nil {
				l.Error(pkgerr.Wrap(err, "delete postpvc failed"), key.Name+" in "+key.Namespace+"delete postpvc failed")
				return pkgerr.Wrap(err, "delete postpvc failed")
			}
		}
	}
	//delete pv
	for postdel := 0; postdel < 3; postdel++ {
		delpv := &corev1.PersistentVolume{}
		key := types.NamespacedName{
			Name:      fmt.Sprintf("%s-%s-%s", instance.GetName(), instance.GetNamespace(), strconv.Itoa(int(postdel))),
			Namespace: instance.Namespace,
		}
		if err := r.Get(ctx, key, delpv); err != nil {
			if errors.IsNotFound(err) {
				l.Error(pkgerr.Wrap(err, "not found, pv is post-deleted"), key.Name+"in"+key.Namespace+"not found")
			} else {
				l.Error(pkgerr.Wrap(err, "get delpv error"), key.Name+"in"+key.Namespace+"get delpv error")
				return pkgerr.Wrap(err, "get delpv error")
			}
		} else {
			if err := r.Delete(ctx, delpv); err != nil {
				l.Error(pkgerr.Wrap(err, "delete postpv failed"), key.Name+"in"+key.Namespace+"delete postpv failed")
				return pkgerr.Wrap(err, "delete postpv failed")
			}
		}
	}
	return nil
}

// new mongodb
func (r *MongodbOperatorReconciler) NewMongodb(ctx context.Context, instance *mongodbv1.MongodbOperator, l logr.Logger) error {
	//first to check if secret exist
	if instance.Spec.MongoDBSpec.Mongodconf.TLS.Enabled {
		customSecret := false
		tlsSecret := &corev1.Secret{}
		tlskey := types.NamespacedName{
			Name:      "mongo-tls-secret",
			Namespace: instance.Namespace,
		}
		if instance.Spec.MongoDBSpec.Mongodconf.TLS.TLSSecretName != "" {
			tlskey = types.NamespacedName{
				Name:      instance.Spec.MongoDBSpec.Mongodconf.TLS.TLSSecretName,
				Namespace: instance.Namespace,
			}
			customSecret = true
		}
		if !customSecret {
			if err := r.Get(ctx, tlskey, tlsSecret); err != nil {
				if errors.IsNotFound(err) {
					l.Info("mongo-tls-secret not found,use default")
					mongoSecret, gerr := baseresource.NewMongoTLSSecret(instance)
					if gerr != nil {
						return pkgerr.Wrap(gerr, "mongoTLSSecret generate fail")
					}
					err := r.createOrUpdate(ctx, mongoSecret, instance, false)
					if err != nil {
						return pkgerr.Wrap(err, "mongoTLSSecret create fail")
					}
				} else {
					return pkgerr.Wrap(err, "get tls-secret error")
				}
			} else {
				l.Info("mongo-tls-secret already exists")
			}
		}
	}

	existSecret := &corev1.Secret{}
	key := types.NamespacedName{
		Name:      instance.Spec.SecretName,
		Namespace: instance.Namespace,
	}
	if err := r.Get(ctx, key, existSecret); err != nil {
		if errors.IsNotFound(err) {
			l.Error(err, "mongo-secret not found")
			return pkgerr.Wrap(err, "secret not created before reconcile")
		} else {
			return pkgerr.Wrap(err, "get secret error")
		}
	} else {
		l.Info("mongo-secret already exists")
	}
	//make mongosvc
	mongoSvcs := baseresource.NewMongoService(instance)
	for _, svc := range mongoSvcs {
		err := r.createOrUpdate(ctx, svc, instance, true)
		if err != nil {
			l.Info("mongosvcs create fail", err)
			return err
		}
	}
	//make mongocm
	mongocm := baseresource.NewMongoConfigmap(instance)
	err := r.createOrUpdate(ctx, mongocm, instance, true)
	if err != nil {
		l.Info("mongocm create fail", err)
		return err
	}
	//make mongopvc
	mongoPvcs := baseresource.NewMongoPersistentVolumeClaim(instance)
	for _, pvc := range mongoPvcs {
		err = r.createOrUpdate(ctx, pvc, instance, false)
		if err != nil {
			l.Info("mongopvcs create fail", err)
			return err
		}
	}
	//make mongopv
	mongoPvs := baseresource.NewMongoPersistentVolume(instance)
	for _, pv := range mongoPvs {
		err = r.createOrUpdate(ctx, pv, instance, false)
		if err != nil {
			l.Info("mongopvs create fail", err)
			return err
		}
	}
	//make mongosts
	mongoSts := baseresource.NewMongoStatefulSet(instance)
	err = r.createOrUpdate(ctx, mongoSts, instance, true)
	if err != nil {
		l.Info("mongosts create fail", err)
		return err
	}
	// 组建mongodb集群
	// 先获取secret里的admin username和password
	// annotation is for mongo 1.x migrate
	upgrade := false
	var adminuser string
	var adminpwd []byte
	annotations := instance.GetAnnotations()
	if annotations != nil {
		if _, ok := annotations["proton-mongodb-1.x-migrate"]; ok {
			upgrade = true
		}
	}
	err = r.Client.Get(context.TODO(), key, existSecret)
	if err != nil {
		return pkgerr.Wrap(err, "get secret error")
	} else {
		adminuser = string(existSecret.Data["username"])
		adminpwd, err = base64.StdEncoding.DecodeString(string(existSecret.Data["password"]))
		if err != nil {
			return pkgerr.Wrap(err, "decode failed")
		}
	}

	replset := instance.Spec.MongoDBSpec.Replset
	err = reconcileMongoReplicaSet(ctx, r, instance, mongoSts, replset, adminuser, string(adminpwd), upgrade, l)
	if err != nil {
		return pkgerr.Wrap(err, "set replicaset fail")
	}
	//make mongocjb
	//mongo集群组建完成后创建日志切割cronjob
	//cronjob需要根据副本数扩缩容
	// storageclass不提供日志切割功能
	if instance.Spec.MongoDBSpec.Storage.StorageClassName == "" {
		for predel := instance.Spec.MongoDBSpec.Replicas; predel < 3; predel++ {
			delcjb := &batchv1.CronJob{}
			key := types.NamespacedName{
				Name:      fmt.Sprintf("logrotate-cron-%s", strconv.Itoa(int(predel))),
				Namespace: instance.Namespace,
			}
			if err := r.Get(ctx, key, delcjb); err != nil {
				if errors.IsNotFound(err) {
					// l.Info("not found, cronjob is pre-deleted")
				} else {
					l.Error(err, "get delcjb error")
					return err
				}
			} else {
				if err := r.Delete(ctx, delcjb); err != nil {
					l.Error(err, "delete precjb failed")
				}
			}
		}

		mongoCjbs := baseresource.NewMongoCronjob(instance)
		for _, cjb := range mongoCjbs {
			err = r.createOrUpdate(ctx, cjb, instance, true)
			if err != nil {
				l.Info("mongocjbs create fail", err)
				return err
			}
		}
	}
	return nil
}

// new mgmt
func (r *MongodbOperatorReconciler) NewMgmt(ctx context.Context, instance *mongodbv1.MongodbOperator, l logr.Logger) error {
	//for backup,create serviceaccount,clusterrole,clusterrolebinding
	mgmtSa := baseresource.NewMgmtServiceAccount(instance)
	err := r.createOrUpdate(ctx, mgmtSa, instance, false)
	if err != nil {
		l.Info("mgmtSA create fail", err)
		return err
	}
	mgmtClusterRole := baseresource.NewMgmtClusterRole(instance)
	err = r.createOrUpdate(ctx, mgmtClusterRole, instance, false)
	if err != nil {
		l.Info("mgmtClusterRole create fail", err)
		return err
	}
	mgmtClusterRoleBinding := baseresource.NewMgmtClusterRoleBinding(instance)
	err = r.createOrUpdate(ctx, mgmtClusterRoleBinding, instance, false)
	if err != nil {
		l.Info("mgmtClusterRoleBinding create fail", err)
		return err
	}
	//make mgmtsvc
	mgmtSvcs := baseresource.NewMgmtService(instance)
	for _, svc := range mgmtSvcs {
		err := r.createOrUpdate(ctx, svc, instance, true)
		if err != nil {
			l.Info("mgmtsvcs create fail", err)
			return err
		}
	}
	//make mgmtsts
	mgmtSts := baseresource.NewMgmtStatefulSet(instance)
	err = r.createOrUpdate(ctx, mgmtSts, instance, true)
	if err != nil {
		l.Info("mgmtsts create fail", err)
		return err
	}
	return nil
}

// new exporter
func (r *MongodbOperatorReconciler) NewExporter(ctx context.Context, instance *mongodbv1.MongodbOperator, l logr.Logger) error {
	//make mongosvc
	exporterSvc := baseresource.NewExporterService(instance)
	err := r.createOrUpdate(ctx, exporterSvc, instance, true)
	if err != nil {
		l.Info("exportersvc create fail", err)
		return err
	}
	//make mongosts
	exporterSts := baseresource.NewExporterStatefulSet(instance)
	err = r.createOrUpdate(ctx, exporterSts, instance, true)
	if err != nil {
		l.Info("exportersts create fail", err)
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MongodbOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mongodbv1.MongodbOperator{}).
		WithEventFilter(ignorePredicate()).
		Complete(r)
}

func (r *MongodbOperatorReconciler) createOrUpdate(ctx context.Context, obj client.Object, instance *mongodbv1.MongodbOperator, controller bool) error {
	metaAccessor, ok := obj.(metav1.ObjectMetaAccessor)
	if !ok {
		return fmt.Errorf("can't convert object to ObjectMetaAccessor")
	}

	objectMeta := metaAccessor.GetObjectMeta()
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = reflect.Indirect(val)
	}
	oldObject, ok := reflect.New(val.Type()).Interface().(client.Object)
	if !ok {
		fmt.Println("new oldobj failed")
	}
	err := r.Client.Get(ctx, types.NamespacedName{
		Name:      objectMeta.GetName(),
		Namespace: objectMeta.GetNamespace(),
	}, oldObject)

	if err != nil && !errors.IsNotFound(err) {
		return pkgerr.Wrap(err, "get oldobject")
	}

	if errors.IsNotFound(err) {
		if controller {
			err = controllerutil.SetControllerReference(instance, obj, r.Scheme)
			if err != nil {
				return err
			}
		}
		return r.Client.Create(ctx, obj)
	}

	oldObjectMeta := oldObject.(metav1.ObjectMetaAccessor).GetObjectMeta()
	objectMeta.SetResourceVersion(oldObjectMeta.GetResourceVersion())
	switch object := obj.(type) {
	case *corev1.Service:
		object.Spec.ClusterIP = oldObject.(*corev1.Service).Spec.ClusterIP
	case *corev1.PersistentVolume:
		return nil
	case *corev1.PersistentVolumeClaim:
		return nil
	}

	if controller {
		err = controllerutil.SetControllerReference(instance, obj, r.Scheme)
		if err != nil {
			return err
		}
	}
	return r.Client.Update(context.TODO(), obj)
}

func (r *MongodbOperatorReconciler) updateMongoReplsetStatus(ctx context.Context, instance *mongodbv1.MongodbOperator, l logr.Logger) error {
	currentstatus := instance.Status
	size := instance.Spec.MongoDBSpec.Replicas

	if currentstatus.Replsets == nil {
		instance.Status = mongodbv1.MongodbOperatorStatus{Replsets: map[string]*mongodbv1.ReplsetStatus{instance.Spec.MongoDBSpec.Replset.Name: {Size: size}}}
	} else {
		if size == currentstatus.Replsets[instance.Spec.MongoDBSpec.Replset.Name].Size {
			l.Info("replset size no change,not update status")
			return nil
		}
		instance.Status.Replsets[instance.Spec.MongoDBSpec.Replset.Name].Size = size
	}

	moclone := instance.DeepCopy()
	err := r.Status().Update(ctx, moclone)
	if err != nil {
		return err
	}
	return nil
}

func ignorePredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
		},
	}
}

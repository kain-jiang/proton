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
	"fmt"
	mongodbv1 "proton-mongodb-operator/api/v1"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ConfigMember struct {
	ID   int    `bson:"_id" json:"_id"`
	Host string `bson:"host" json:"host"`
}

type ConfigMembers []ConfigMember

type RSConfig struct {
	ID      string        `bson:"_id" json:"_id"`
	Version int           `bson:"version" json:"version"`
	Members ConfigMembers `bson:"members" json:"members"`
	// Configsvr                          bool          `bson:"configsvr,omitempty" json:"configsvr,omitempty"`
	// Settings                           Settings      `bson:"settings,omitempty" json:"settings,omitempty"`
	WriteConcernMajorityJournalDefault bool `bson:"writeConcernMajorityJournalDefault,omitempty" json:"writeConcernMajorityJournalDefault,omitempty"`
}

func reconcileMongoReplicaSet(ctx context.Context, r *MongodbOperatorReconciler, instance *mongodbv1.MongodbOperator, mongosts *appsv1.StatefulSet, replset *mongodbv1.ReplsetSpec, adminuser string, adminpwd string, upgrade bool, logger logr.Logger) error {
	logger.Info("============ReconcileMongoReplicaSet===============")
	rsName := replset.Name
	replicas := int(*mongosts.Spec.Replicas)
	members := bson.A{}
	for v := 0; v < replicas; v++ {
		host := mongosts.Name + "-" + strconv.Itoa(v) + "." + mongosts.Name + "." + mongosts.Namespace + ".svc.cluster.local:" + strconv.Itoa(int(mongosts.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort))
		members = append(members, bson.D{
			{Key: "_id", Value: v},
			{Key: "host", Value: host},
		})
	}
	cfg := bson.D{
		{Key: "_id", Value: rsName},
		{Key: "members", Value: members},
		{Key: "settings", Value: bson.D{
			{Key: "getLastErrorDefaults", Value: bson.D{
				{Key: "w", Value: "majority"},
				{Key: "wtimeout", Value: 5000},
			},
			},
		},
		},
	}
	pods, err := GetRSPods(ctx, r.Client, instance, rsName)
	if err != nil {
		return errors.Wrapf(err, "get pods list for replset %s", rsName)
	}

	// compare status and spec to reconfig
	if instance.Status.Replsets == nil {
		logger.Info("install, or upgrade from mongodb-1.x, config rs")
		err = r.handleReplsetInit(ctx, instance, replset, pods.Items, cfg, adminuser, adminpwd, upgrade, logger)
		if err != nil {
			return errors.Wrap(err, "handleReplsetInit")
		}
	} else if instance.Status.Replsets[instance.Spec.MongoDBSpec.Replset.Name].Size != instance.Spec.MongoDBSpec.Replicas {
		// 副本数不一致,如果不是缩容为零,重新组建集群
		if instance.Spec.MongoDBSpec.Replicas != 0 {
			logger.Info("scaling")
			mongoc, err := r.handleReplsetReconfig(ctx, instance, replset, pods.Items, adminuser, adminpwd, logger)
			if err != nil {
				logger.Error(err, "can't get new mongo client")
				return errors.Wrap(err, "reconfig replset fail")
			}
			defer func() {
				if err := mongoc.Disconnect(ctx); err != nil {
					logger.Error(err, "failed to close connection")
				}
			}()
			cnf, err := ReadConfig(ctx, mongoc)
			if err != nil {
				return errors.Wrap(err, "get mongo replset config failed")
			}
			version := cnf.Version
			version++
			cfg = append(cfg, bson.E{Key: "version", Value: version})
			err = WriteConfig(ctx, mongoc, cfg)
			if err != nil {
				return errors.Wrap(err, "scale replset,reconfig failed")
			}
			logger.Info("success scale replicas from " + strconv.Itoa(int(instance.Status.Replsets[instance.Spec.MongoDBSpec.Replset.Name].Size)) + " to " + strconv.Itoa(replicas))
		} else {
			logger.Info("scale to 0 replicas")
		}
	} else {
		logger.Info("not upgrade mongo spec")
	}

	return nil
}

// OKResponse is a standard MongoDB response
type OKResponse struct {
	Errmsg string `bson:"errmsg,omitempty" json:"errmsg,omitempty"`
	OK     int    `bson:"ok" json:"ok"`
	Code   int    `bson:"code" json:"code"`
}

// Response document from 'replSetGetConfig': https://docs.mongodb.com/manual/reference/command/replSetGetConfig/#dbcmd.replSetGetConfig
type ReplSetGetConfig struct {
	Config     *RSConfig `bson:"config" json:"config"`
	OKResponse `bson:",inline"`
}

func ReadConfig(ctx context.Context, client *mongo.Client) (RSConfig, error) {
	resp := ReplSetGetConfig{}
	res := client.Database("admin").RunCommand(ctx, bson.D{{Key: "replSetGetConfig", Value: 1}})
	if res.Err() != nil {
		return RSConfig{}, errors.Wrap(res.Err(), "replSetGetConfig")
	}
	if err := res.Decode(&resp); err != nil {
		return RSConfig{}, errors.Wrap(err, "failed to decode to replSetGetConfig")
	}

	if resp.Config == nil {
		return RSConfig{}, errors.Errorf("mongo says: %s", resp.Errmsg)
	}

	return *resp.Config, nil
}

// isContainerAndPodRunning returns a boolean reflecting if
// a container and pod are in a running state
func isContainerAndPodRunning(pod corev1.Pod, containerName string) bool {
	if pod.Status.Phase != corev1.PodRunning {
		return false
	}
	for _, container := range pod.Status.ContainerStatuses {
		if container.Name == containerName && container.State.Running != nil {
			return true
		}
	}
	return false
}

type Config struct {
	Hosts       []string
	ReplSetName string
	Username    string
	Password    string
	// TLSConf     *tls.Config
	Direct bool
}

func WriteConfig(ctx context.Context, client *mongo.Client, cfg bson.D) error {
	resp := OKResponse{}

	// Using force flag since mongo 4.4 forbids to add multiple members at a time.
	res := client.Database("admin").
		RunCommand(ctx, bson.D{
			{Key: "replSetReconfig", Value: cfg},
			{Key: "force", Value: true},
		})
	if res.Err() != nil {
		return errors.Wrap(res.Err(), "replSetReconfig")
	}

	if err := res.Decode(&resp); err != nil {
		return errors.Wrap(err, "failed to decode to replSetReconfigResponse")
	}

	if resp.OK != 1 {
		return errors.Errorf("mongo says: %s", resp.Errmsg)
	}

	return nil
}

type Member struct {
	Id     int    `bson:"_id" json:"_id"`
	Name   string `bson:"name" json:"name"`
	Health int    `bson:"health" json:"health"`
	State  int    `bson:"state" json:"state"`
	// Uptime            int64               `bson:"uptime" json:"uptime"`
	// Optime            *Optime             `bson:"optime" json:"optime"`
	// OptimeDate        time.Time           `bson:"optimeDate" json:"optimeDate"`
	// ConfigVersion     int                 `bson:"configVersion" json:"configVersion"`
	// ElectionTime      primitive.Timestamp `bson:"electionTime,omitempty" json:"electionTime,omitempty"`
	// ElectionDate      time.Time           `bson:"electionDate,omitempty" json:"electionDate,omitempty"`
	// InfoMessage       string              `bson:"infoMessage,omitempty" json:"infoMessage,omitempty"`
	// OptimeDurable     *Optime             `bson:"optimeDurable,omitempty" json:"optimeDurable,omitempty"`
	// OptimeDurableDate time.Time           `bson:"optimeDurableDate,omitempty" json:"optimeDurableDate,omitempty"`
	Self bool `bson:"self,omitempty" json:"self,omitempty"`
}
type Status struct {
	// Set                     string      `bson:"set" json:"set"`
	// Date                    time.Time   `bson:"date" json:"date"`
	MyState                 int       `bson:"myState" json:"myState"`
	Members                 []*Member `bson:"members" json:"members"`
	Term                    int64     `bson:"term,omitempty" json:"term,omitempty"`
	HeartbeatIntervalMillis int64     `bson:"heartbeatIntervalMillis,omitempty" json:"heartbeatIntervalMillis,omitempty"`
	// Optimes                 *StatusOptimes `bson:"optimes,omitempty" json:"optimes,omitempty"`
	OKResponse `bson:",inline"`
}

func RSStatus(ctx context.Context, client *mongo.Client) (Status, error) {
	status := Status{}

	resp := client.Database("admin").RunCommand(ctx, bson.D{{Key: "replSetGetStatus", Value: 1}})
	if resp.Err() != nil {
		return status, errors.Wrap(resp.Err(), "replSetGetStatus")
	}

	if err := resp.Decode(&status); err != nil {
		return status, errors.Wrap(err, "failed to decode rs status")
	}

	if status.OK != 1 {
		return status, errors.Errorf("mongo says: %s", status.Errmsg)
	}

	return status, nil
}

type Hello struct {
	IsWritablePrimary bool `bson:"isWritablePrimary" json:"isWritablePrimary"`
	OKResponse        `bson:",inline"`
}

type Userinfo struct {
	User string `bson:"user" json:"user"`
	Db   string `bson:"db" json:"db"`
}

type UsersInfo struct {
	Users      []*Userinfo `bson:"users" json:"users"`
	OKResponse `bson:",inline"`
}

// handleReplsetInit runs the k8s-mongodb-initiator from within the first running pod's mongod container.
// This must be ran from within the running container to utilize the MongoDB Localhost Exception.
// See: https://docs.mongodb.com/manual/core/security-users/#localhost-exception
var errNoRunningMongodContainers = errors.New("no mongod containers in running state")
var errUpgradefail = errors.New("proton-mongodb-1.x-migrate fail")

func (r *MongodbOperatorReconciler) handleReplsetInit(ctx context.Context, instance *mongodbv1.MongodbOperator, replset *mongodbv1.ReplsetSpec, pods []corev1.Pod, cfg bson.D, adminuser string, adminpwd string, upgrade bool, logger logr.Logger) error {
	for _, pod := range pods {
		// if !isMongodPod(pod) || !isContainerAndPodRunning(pod, "mongodb") || !isPodReady(pod) {
		if !isMongodPod(pod) || !isContainerAndPodRunning(pod, "mongodb") {
			continue
		}
		logger.Info("initiating replset", "replset", replset.Name, "pod", pod.Name)
		clientOptions := options.Client().
			ApplyURI(fmt.Sprintf("mongodb://%s/admin", pod.Name+"."+instance.Name+"-mongodb."+pod.Namespace+".svc.cluster.local:"+strconv.Itoa(int(pod.Spec.Containers[0].Ports[0].ContainerPort)))).
			SetAuth(options.Credential{
				Username: adminuser,
				Password: adminpwd,
			}).
			SetDirect(true)
		mongoc, err := mongo.Connect(context.TODO(), clientOptions)
		if err != nil {
			return errors.Wrap(err, "Failed to create mongodb connect, using operator secret")
		}

		defer func() {
			if err != nil {
				derr := mongoc.Disconnect(ctx)
				if derr != nil {
					logger.Error(err, "failed to disconnect")
				}
			}
		}()

		err = mongoc.Database("admin").RunCommand(context.TODO(), bson.D{
			{Key: "replSetInitiate", Value: bson.D{}},
		}).Err()
		if err != nil {
			if strings.Contains(err.Error(), "already initialized") {
				logger.Info("Mongo RS AlreadyInitialized,Ignore")
			} else {
				return errors.Wrap(err, "replset init failed")
			}
		}

		cnf, err := ReadConfig(ctx, mongoc)
		if err != nil {
			return errors.Wrap(err, "get init mongo config fail")
		} else {
			version := cnf.Version
			version++
			cfg = append(cfg, bson.E{Key: "version", Value: version})
		}
		err = mongoc.Database("admin").RunCommand(context.TODO(), bson.D{
			{Key: "replSetReconfig", Value: cfg},
			{Key: "force", Value: true},
		}).Err()
		if err != nil {
			return errors.Wrap(err, "replset reconfig failed")
		}

		time.Sleep(time.Second * 5)

		logger.Info("replset was initialized", "replset", replset.Name, "pod", pod.Name)
		return nil
	}

	return errNoRunningMongodContainers
}

// isMongodPod returns a boolean reflecting if a pod
// is running a mongod container
func isMongodPod(pod corev1.Pod) bool {
	return getPodContainer(&pod, "mongodb") != nil
}

func getPodContainer(pod *corev1.Pod, containerName string) *corev1.Container {
	for _, cont := range pod.Spec.Containers {
		if cont.Name == containerName {
			return &cont
		}
	}
	return nil
}

// isPodReady returns a boolean reflecting if a pod is in a "ready" state
// func isPodReady(pod corev1.Pod) bool {
// 	for _, condition := range pod.Status.Conditions {
// 		if condition.Status != corev1.ConditionTrue {
// 			continue
// 		}
// 		if condition.Type == corev1.PodReady {
// 			return true
// 		}
// 	}
// 	return false
// }

func GetRSPods(ctx context.Context, k8sclient client.Client, cr *mongodbv1.MongodbOperator, rsName string) (corev1.PodList, error) {
	pods := corev1.PodList{}
	err := k8sclient.List(ctx,
		&pods,
		&client.ListOptions{
			Namespace:     cr.Namespace,
			LabelSelector: labels.SelectorFromSet(rsLabels(cr, rsName)),
		},
	)

	return pods, err
}

func rsLabels(cr *mongodbv1.MongodbOperator, rsName string) map[string]string {
	lbls := make(map[string]string, 0)
	lbls["app"] = fmt.Sprintf("%s-%s", cr.Name, "mongodb")
	return lbls
}

func (r *MongodbOperatorReconciler) handleReplsetReconfig(ctx context.Context, instance *mongodbv1.MongodbOperator, replset *mongodbv1.ReplsetSpec, pods []corev1.Pod, adminuser string, adminpwd string, logger logr.Logger) (*mongo.Client, error) {
	// var errNoRunningMongodContainers = errors.("no mongod containers in running state")
	for _, pod := range pods {
		// if !isMongodPod(pod) || !isContainerAndPodRunning(pod, "mongodb") || !isPodReady(pod) {
		if !isMongodPod(pod) || !isContainerAndPodRunning(pod, "mongodb") {
			continue
		}
		clientOptions := options.Client().
			ApplyURI(fmt.Sprintf("mongodb://%s/admin", pod.Name+"."+instance.Name+"-mongodb."+pod.Namespace+".svc.cluster.local:"+strconv.Itoa(int(pod.Spec.Containers[0].Ports[0].ContainerPort)))).
			SetAuth(options.Credential{
				Username: adminuser,
				Password: adminpwd,
			}).
			SetDirect(true)
		mongoc, err := mongo.Connect(context.TODO(), clientOptions)
		if err != nil {
			logger.Error(err, "Failed to create mongodb connect")
		}

		ctx, pingcancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer pingcancel()

		err = mongoc.Ping(ctx, nil)
		if err != nil {
			logger.Error(err, "mongo ping error")
			return nil, errors.Wrap(err, "ping mongo")
		}
		return mongoc, nil
	}
	return nil, errNoRunningMongodContainers
}

// func mongoCreateAdminUser(user, pwd string) string {
// 	return fmt.Sprintf("'db.getSiblingDB(\"admin\").auth(\"%s\",\"%s\")&&"+
// 		"db.getSiblingDB(\"admin\").createUser("+
// 		"{"+
// 		"user: \"%s\","+
// 		"pwd: \"%s\","+
// 		"roles: [ {role: \"root\", db: \"admin\"} ]"+
// 		"})'", user, pwd, user, pwd)
// }

func mongoGetAdminUser(user, pwd string) string {
	return fmt.Sprintf("'db.getSiblingDB(\"admin\").auth(\"%s\",\"%s\")&&"+
		"db.getSiblingDB(\"admin\").getUser(\"%s\")'", user, pwd, user)
}

func mongoUpgradeChangeAdminUserPassword(user, pwd string) string {
	return fmt.Sprintf("'db.getSiblingDB(\"admin\").auth(\"%s\",\"%s\")&&"+
		"db.getSiblingDB(\"admin\").updateUser(\"%s\","+
		"{"+
		"authenticationRestrictions:[]"+
		"})'", user, pwd, user)
}

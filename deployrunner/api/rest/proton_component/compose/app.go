package compose

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	"taskrunner/api/rest"
	"taskrunner/pkg/utils"
	"taskrunner/trait"

	"github.com/gin-gonic/gin"
	// v1 "k8s.io/api/core/v1"
	// kerrors "k8s.io/apimachinery/pkg/api/errors"
	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

type applicationOperator struct {
	*rest.ExecutorEngine
	kcli kubernetes.Interface
}

func newApplicationOperator(r *rest.ExecutorEngine, kcli kubernetes.Interface) *applicationOperator {
	return &applicationOperator{
		ExecutorEngine: r,
		kcli:           kcli,
	}
}

func (op *applicationOperator) CreateJob(ctx context.Context, a *trait.ApplicationInstance) (int, *trait.Error) {
	e := op.ExecutorEngine
	ajid := -1

	err := utils.RetryN(ctx, func() (bool, *trait.Error) {
		aid, err := e.GetAPPID(ctx, a.AName, a.Version)
		if trait.IsInternalError(err, trait.ErrNotFound) {
			err.Detail = fmt.Sprintf("aname: %s, version: %s", a.AName, a.Version)
			return false, err
		}
		if err != nil {
			return true, err
		}
		a.AID = aid
		jid, err := e.CreateAndSetJobWithConfig(ctx, a)
		if trait.IsInternalError(err, trait.ErrNotFound) {
			return false, err
		}
		if trait.IsInternalError(err, trait.ErrParam) {
			return false, err
		}
		if err != nil {
			return true, err
		}
		ajid = jid
		return false, nil
	}, 10, 3*time.Second)

	return ajid, err
}

func (op *applicationOperator) StartJob(ctx context.Context, jid int) *trait.Error {
	r := op.ExecutorEngine.Engine
	return utils.RetryN(ctx, func() (bool, *trait.Error) {
		req, rerr := http.NewRequestWithContext(context.Background(), http.MethodPut, fmt.Sprintf("/job/executor/%d", jid), nil)
		if rerr != nil {
			err := &trait.Error{
				Err:      rerr,
				Internal: trait.ECNetUnknow,
				Detail: fmt.Sprintf(
					"start job %d request init fail",
					jid,
				),
			}
			return false, err
		}
		resp := httptest.NewRecorder()
		gctx := gin.CreateTestContextOnly(resp, r)
		gctx.Params = append(gctx.Params, gin.Param{Key: "jid", Value: strconv.Itoa(jid)})
		gctx.Request = req
		op.ExecutorEngine.StartJob(gctx)
		// r.ServeHTTP(resp, req)
		bs, rerr := io.ReadAll(resp.Body)
		if rerr != nil {
			err := &trait.Error{
				Err:      rerr,
				Internal: trait.ECNULL,
				Detail: fmt.Sprintf(
					"read start job %d request reponse fail",
					jid,
				),
			}
			return false, err
		}
		if resp.Code != 200 {
			switch resp.Code {
			case 400:
				err := &trait.Error{
					Err:      errors.New(string(bs)),
					Internal: trait.ErrParam,
					Detail: fmt.Sprintf(
						"start job %d error", jid,
					),
				}
				return true, err
			case 404:
				err := &trait.Error{
					Err:      errors.New(string(bs)),
					Internal: trait.ErrNotFound,
					Detail: fmt.Sprintf(
						"start job %d error", jid,
					),
				}
				return true, err
			default:
				err := &trait.Error{
					Err:      errors.New(string(bs)),
					Internal: trait.ECNULL,
					Detail: fmt.Sprintf(
						"start job %d error", jid,
					),
				}
				return true, err
			}
		}
		return false, nil
	}, 10, 3*time.Second)
}

func (op *applicationOperator) StopJob(ctx context.Context, jid int) *trait.Error {
	r := op.ExecutorEngine.Engine
	return utils.RetryN(ctx, func() (bool, *trait.Error) {
		req, rerr := http.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("/job/executor/%d", jid), nil)
		if rerr != nil {
			err := &trait.Error{
				Err:      rerr,
				Internal: trait.ECNetUnknow,
				Detail: fmt.Sprintf(
					"stop job %d request init fail",
					jid,
				),
			}
			return false, err
		}
		resp := httptest.NewRecorder()
		gctx := gin.CreateTestContextOnly(resp, r)
		gctx.Params = append(gctx.Params, gin.Param{Key: "jid", Value: strconv.Itoa(jid)})
		gctx.Request = req
		op.ExecutorEngine.StartJob(gctx)
		// r.ServeHTTP(resp, req)
		bs, rerr := io.ReadAll(resp.Body)
		if rerr != nil {
			err := &trait.Error{
				Err:      rerr,
				Internal: trait.ECNULL,
				Detail: fmt.Sprintf(
					"read stop job %d request reponse fail",
					jid,
				),
			}
			return false, err
		}
		if resp.Code != 200 {
			switch resp.Code {
			case 404:
				return false, nil
			default:
				err := &trait.Error{
					Err:      errors.New(string(bs)),
					Internal: trait.ECNULL,
					Detail: fmt.Sprintf(
						"stop job %d error", jid,
					),
				}
				return true, err
			}
		}
		return false, nil
	}, 10, 3*time.Second)
}

func (op *applicationOperator) StopAndWaitJob(ctx context.Context, jid int) *trait.Error {
	if err := op.StopJob(ctx, jid); err != nil {
		op.Log.Errorf("try cancel application job %d error: %s", jid, err.Error())
		return err
	}
	_, err := op.WaitJob(ctx, jid)
	if err != nil {
		if trait.IsInternalError(err, trait.ErrNotFound) {
			return nil
		}
		op.Log.Errorf("wait cancel job %d finish error: %s", jid, err.Error())
	}
	return err
}

func (op *applicationOperator) WaitJob(ctx context.Context, jid int) (int, *trait.Error) {
	s := op.ExecutorEngine.Store
	for {
		j, err := s.GetJobRecord(ctx, jid)
		if trait.IsInternalError(err, trait.ErrNotFound) {
			op.Log.Warnf("wait job operate can't find job %d, see as fail ", jid)
			return trait.AppFailStatus, err
		}
		if trait.IsInternalError(err, trait.ECJobCancel) {
			return trait.AppFailStatus, err
		}
		if trait.IsInternalError(err, trait.ECExit) {
			// just return if main exit
			return trait.AppDoingStatus, err
		}
		if err != nil {
			op.Log.Warnf("wait job operate receive error: %s, retry later", err.Error())
			time.Sleep(10 * time.Second)
			continue
		}
		switch j.Target.Status {
		case trait.AppinitStatus, trait.AppConfirmedStatus,
			trait.AppSucessStatus, trait.AppFailStatus,
			trait.AppStopedStatus, trait.AppFailMissStatus,
			trait.AppFailUninstallStatus, trait.AppUpgradeParentComponentFailStatus:
			return j.Target.Status, nil
		default:
			time.Sleep(10 * time.Second)
		}
	}
}

func (op *applicationOperator) CreateSystem(ctx context.Context, s trait.System) (int, *trait.Error) {
	if op.ExecutorEngine.SID >= 0 {
		// auto system is single instance mode, don't create system
		return op.ExecutorEngine.SID, nil
	}
	ss := op.Executor
	sid := -1
	return sid, utils.RetryN(
		ctx,
		func() (bool, *trait.Error) {
			if s.SName == "" || s.NameSpace == "" {
				cur, err := ss.GetSystemInfo(ctx, s.SID)
				if trait.IsInternalError(err, trait.ErrNotFound) {
					return false, err
				} else if err != nil {
					return true, err
				} else {
					sid = cur.SID
					return false, nil
				}
			}

			id, err := ss.InsertSystemInfo(ctx, s)
			if trait.IsInternalError(err, trait.ErrUniqueKey) {
				sc, gerr := ss.GetSystemInfoByName(ctx, s.SName)
				err = gerr
				if err == nil {
					id = sc.SID
					if s.NameSpace != sc.NameSpace {
						return false, &trait.Error{
							Internal: trait.ErrParam,
							Detail:   s,
							Err:      fmt.Errorf("system %s can't change namespace from %s into %s", s.SName, sc.NameSpace, s.NameSpace),
						}
					}
				}
			}
			if err != nil {
				op.Log.Errorf("create compose job system error: %s, retry later", err.Error())
				return true, err
			}
			sid = id
			return false, nil
		},
		10,
		3*time.Second,
	)
}

// func (op *applicationOperator) CreateAccessInfo(ctx context.Context, info trait.AccessInfo, namespace string) *trait.Error {
// 	// auto system is single instance mode, don't create access info
// 	if op.ExecutorEngine.Sid >= 0 {
// 		return nil
// 	}
// 	kcli := op.kcli
// 	_, kerr := kcli.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
// 	if kerrors.IsNotFound(kerr) {
// 		_, kerr = kcli.CoreV1().Namespaces().Create(ctx, &v1.Namespace{
// 			ObjectMeta: metav1.ObjectMeta{
// 				Name: namespace,
// 			},
// 		}, metav1.CreateOptions{})
// 	}
// 	if kerr != nil {
// 		return &trait.Error{
// 			Internal: trait.ECK8sUnknow,
// 			Err:      kerr,
// 			Detail:   "init namespace error",
// 		}
// 	}
// 	_, kerr = op.kcli.CoreV1().Secrets(namespace).Get(ctx, "cms-release-config-anyshare", metav1.GetOptions{})
// 	if kerrors.IsNotFound(kerr) {
// 		// only create once, won't update
// 		se := &v1.Secret{
// 			ObjectMeta: metav1.ObjectMeta{
// 				Name: "cms-release-config-anyshare",
// 			},
// 			Data: map[string][]byte{
// 				"default.yaml":  info.ToBytes(),
// 				"name":          []byte("anyshare"),
// 				"use":           []byte("default.yaml"),
// 				"encrypt_field": []byte("[]"),
// 			},
// 			Type: v1.SecretTypeOpaque,
// 		}
// 		_, kerr = op.kcli.CoreV1().Secrets(namespace).Create(ctx, se, metav1.CreateOptions{})

// 	}
// 	if kerr != nil {
// 		return &trait.Error{
// 			Internal: trait.ECK8sUnknow,
// 			Err:      kerr,
// 			Detail:   "init access info error",
// 		}
// 	}
// 	return nil
// }

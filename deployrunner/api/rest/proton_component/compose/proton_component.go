package compose

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"

	"taskrunner/api/rest/proton_component"
	"taskrunner/pkg/utils"
	"taskrunner/trait"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type protonComponentOperator struct {
	// s   *system.Server
	r   *gin.Engine
	log *logrus.Logger
}

type ConnectObj struct {
	Type string          `json:"type"`
	Bs   json.RawMessage `json:"obj"`
}

func newProtonComponentOPerator(s proton_component.GinServer, log *logrus.Logger) protonComponentOperator {
	e := gin.New()
	e.ContextWithFallback = true
	s.RegistryHandler(&e.RouterGroup)
	return protonComponentOperator{
		// s: s,
		r:   e,
		log: log,
	}
}

func (op *protonComponentOperator) getBody(bs []byte, job trait.ComposeJobMeata) ([]byte, string, *trait.Error) {
	obj := ConnectObj{}
	rerr := json.Unmarshal(bs, &obj)
	var body []byte
	if rerr != nil {
		err := &trait.Error{
			Err:      rerr,
			Internal: trait.ErrParam,
			Detail: fmt.Sprintf(
				"compose job name: [%s], jid [%d], sid [%d] decode proton component fail",
				job.Jname, job.Jid, job.SID,
			),
		}
		return nil, "", err
	}
	// switch obj.Type {
	// case "mq":
	o := make(map[string]any)
	rerr = json.Unmarshal(obj.Bs, &o)
	if rerr != nil {
		err := &trait.Error{
			Err:      rerr,
			Internal: trait.ErrParam,
			Detail: fmt.Sprintf(
				"compose job name: [%s], jid [%d], sid [%d] decode %s fail",
				job.Jname, job.Jid, job.SID, obj.Type,
			),
		}
		return nil, "", err
	}
	o["sid"] = job.SID
	nbs, rerr := json.Marshal(o)
	if rerr != nil {
		err := &trait.Error{
			Err:      rerr,
			Internal: trait.ErrParam,
			Detail: fmt.Sprintf(
				"compose job name: [%s], jid [%d], sid [%d] decode %s fail",
				job.Jname, job.Jid, job.SID, obj.Type,
			),
		}
		return nil, "", err
	}
	body = nbs
	// }

	return body, obj.Type, nil
}

func (op *protonComponentOperator) InstallProtonComponent(ctx context.Context, bs []byte, job trait.ComposeJobMeata) *trait.Error {
	select {
	case <-ctx.Done():
		// operate cacel
		return ctx.Err().(*trait.Error)
	default:
		// continue
	}
	body, ctype, err := op.getBody(bs, job)
	if err != nil {
		return err
	}

	return utils.RetryN(ctx, func() (bool, *trait.Error) {
		var err *trait.Error
		defer func() {
			if err != nil {
				op.log.Error(err.Error())
			}
		}()
		op.log.Debugf("start install compose job name: [%s], jid [%d], sid [%d] proton component",
			job.Jname, job.Jid, job.SID)
		r, rerr := http.NewRequestWithContext(context.Background(), http.MethodPut, fmt.Sprintf("/components/info/%s", ctype), bytes.NewReader(body))
		if rerr != nil {
			err = &trait.Error{
				Err:      rerr,
				Internal: trait.ECNetUnknow,
				Detail: fmt.Sprintf(
					"compose job name: [%s], jid [%d], sid [%d] proton component request init fail",
					job.Jname, job.Jid, job.SID,
				),
			}
			return false, err
		}
		resp := httptest.NewRecorder()
		op.r.ServeHTTP(resp, r)
		bs, rerr = io.ReadAll(resp.Body)
		if rerr != nil {
			err = &trait.Error{
				Err:      rerr,
				Internal: trait.ECNULL,
				Detail: fmt.Sprintf(
					"compose job name: [%s], jid [%d], sid [%d] read proton component request response error",
					job.Jname, job.Jid, job.SID,
				),
			}
			return false, err
		}
		if resp.Code == 400 {
			err = &trait.Error{
				Err:      errors.New(string(bs)),
				Internal: trait.ErrParam,
				Detail: fmt.Sprintf(
					"compose job name: [%s], jid [%d], sid [%d] proton component [%s] request param error",
					job.Jname, job.Jid, job.SID, ctype,
				),
			}
			return false, err
		} else if resp.Code != 200 {
			err = &trait.Error{
				Err:      errors.New(string(bs)),
				Internal: trait.ECNULL,
				Detail: fmt.Sprintf(
					"compose job name: [%s], jid [%d], sid [%d] proton component request error, status code: [%d]",
					job.Jname, job.Jid, job.SID, resp.Code,
				),
			}
			return true, err
		}
		return false, nil
	}, 10, 10*time.Second)
}

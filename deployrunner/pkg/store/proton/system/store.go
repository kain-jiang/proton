package system

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"taskrunner/pkg/component/resources"
	"taskrunner/trait"
)

type CursorConn = DBConn

const (
	InfoKey        = "info"
	BindKey        = "bind"
	_TypeLabel     = "proton.connect.type"
	_ConnectPrefix = "pbsc"
)

func NewDBConn(w trait.ProtonComponentWriter) (*DBConn, *trait.Error) {
	return &DBConn{
		ProtonComponentWriter: w,
	}, nil
}

type ConnectObj struct {
	Info    any
	Options any
}

type DBConn struct {
	trait.ProtonComponentWriter
}

func (s *DBConn) GetComponentInfo(ctx context.Context, name, ctype string, sid int, receiver ConnectObj) *trait.Error {
	w := s

	obj, err := w.GetProtonComponent(ctx, name, ctype, sid)
	if err != nil {
		return err
	}

	if receiver.Info != nil {
		if rerr := json.Unmarshal(obj.Attribute, receiver.Info); rerr != nil {
			return &trait.Error{
				Err:      rerr,
				Internal: trait.ErrComponentDecodeError,
				Detail:   fmt.Sprintf("decode error: %s", string(obj.Attribute)),
			}
		}
	}

	if receiver.Options != nil {
		if obj.Options != nil {
			if rerr := json.Unmarshal(obj.Options, receiver.Options); rerr != nil {
				return &trait.Error{
					Err:      rerr,
					Internal: trait.ErrComponentDecodeError,
					Detail:   fmt.Sprintf("decode error: %s", string(obj.Options)),
				}
			}
		}
	}

	return nil
}

func (s *DBConn) StoreConnect(ctx context.Context, name, ctype string, sid int, data ConnectObj) *trait.Error {
	w := s
	obj := trait.ProtonCompoent{
		ProtonComponentMeta: trait.ProtonComponentMeta{
			Name: name,
			Type: ctype,
			System: trait.System{
				SID: sid,
			},
		},
	}

	bs, rerr := json.Marshal(data.Info)
	if rerr != nil {
		return &trait.Error{
			Err:      rerr,
			Internal: trait.ErrParam,
			Detail:   fmt.Sprintf("type: %s, name: %s, sid: %d, %#v", ctype, name, sid, data),
		}
	}
	obj.Attribute = bs
	if data.Options != nil {
		bs, rerr := json.Marshal(data.Options)
		if rerr != nil {
			return &trait.Error{
				Err:      rerr,
				Internal: trait.ErrParam,
				Detail:   fmt.Sprintf("type: %s, name: %s, sid: %d, %#v", ctype, name, sid, data),
			}
		}
		obj.Options = bs
	}

	err := w.InsertProtonComponent(ctx, obj)
	if trait.IsInternalError(err, trait.ErrUniqueKey) {
		err = w.UpdateProtonComponent(ctx, obj)
	}

	return err
}

func numToLetter(n int) string {
	res := bytes.NewBuffer(nil)
	for {
		reamin := n % 26
		res.WriteRune(rune('a' + reamin))
		n = n / 26
		if n == 0 {
			break
		}
	}
	return res.String()
}

func ComponentToValues(ctype string, bs []byte, sid int, core *CoreConfig) (map[string]interface{}, map[string]interface{}, *trait.Error) {
	// TODO reflect to avoid code
	var attr, cfg map[string]interface{}
	var err *trait.Error
	switch ctype {
	case resources.RDSType:
		rds := &resources.RDS{}
		if rerr := json.Unmarshal(bs, rds); rerr != nil {
			return nil, cfg, &trait.Error{
				Err:      rerr,
				Internal: trait.ErrComponentDecodeError,
				Detail:   fmt.Sprintf("decode error: %s", string(bs)),
			}
		}
		attr, cfg, err = rds.ToMap()
	case resources.REDISType:
		obj := &resources.Redis{}
		if rerr := json.Unmarshal(bs, obj); rerr != nil {
			return nil, cfg, &trait.Error{
				Err:      rerr,
				Internal: trait.ErrComponentDecodeError,
				Detail:   fmt.Sprintf("decode error: %s", string(bs)),
			}
		}
		attr, err = obj.ToMap()

	case resources.MQType:
		obj := &resources.MQ{}
		if rerr := json.Unmarshal(bs, obj); rerr != nil {
			return nil, cfg, &trait.Error{
				Err:      rerr,
				Internal: trait.ErrComponentDecodeError,
				Detail:   fmt.Sprintf("decode error: %s", string(bs)),
			}
		}
		attr, err = obj.ToMap()
	case resources.OpensearchType:
		obj := &resources.Opensearch{}
		if rerr := json.Unmarshal(bs, obj); rerr != nil {
			return nil, cfg, &trait.Error{
				Err:      rerr,
				Internal: trait.ErrComponentDecodeError,
				Detail:   fmt.Sprintf("decode error: %s", string(bs)),
			}
		}
		attr, err = obj.ToDepMap()
	case resources.MongodbType:
		obj := &resources.MongoDB{}
		if rerr := json.Unmarshal(bs, obj); rerr != nil {
			return nil, cfg, &trait.Error{
				Err:      rerr,
				Internal: trait.ErrComponentDecodeError,
				Detail:   fmt.Sprintf("decode error: %s", string(bs)),
			}
		}
		attr, err = obj.ToDepMap()

	case resources.EtcdType:
		// TODO COPY SECRET
		obj := &resources.Etcd{}
		if rerr := json.Unmarshal(bs, obj); rerr != nil {
			return nil, cfg, &trait.Error{
				Err:      rerr,
				Internal: trait.ErrComponentDecodeError,
				Detail:   fmt.Sprintf("decode error: %s", string(bs)),
			}
		}
		attr, err = obj.ToDepMap("")
	case resources.POAType:
		obj := map[string]interface{}{}
		if rerr := json.Unmarshal(bs, &obj); rerr != nil {
			return nil, cfg, &trait.Error{
				Err:      rerr,
				Internal: trait.ErrComponentDecodeError,
				Detail:   fmt.Sprintf("decode error: %s", string(bs)),
			}
		}
		if host, ok := obj["hosts"]; !ok {
			err = &trait.Error{
				Err:      fmt.Errorf("%s attribute error, the hosts attribute no set", ctype),
				Internal: trait.ErrComponentDecodeError,
			}
		} else {
			obj["host"] = host
		}
		attr = obj
	case resources.GraphType:
		return nil, cfg, &trait.Error{
			Err:      fmt.Errorf("%s not support nebula in multil instance mode", ctype),
			Internal: trait.ErrComponentTypeNotDefined,
		}
	case resources.DeployCoreType:
		attr = core.ToMapValues()
	default:
		return nil, cfg, &trait.Error{
			Err:      fmt.Errorf("%s type not defined", ctype),
			Internal: trait.ErrComponentTypeNotDefined,
		}
	}

	if err != nil {
		return attr, cfg, err
	}
	return attr, cfg, err
}

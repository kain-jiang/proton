package system

import (
	"context"
	"encoding/json"
	"fmt"

	"taskrunner/pkg/component"
	"taskrunner/pkg/component/resources"
	"taskrunner/pkg/store/proton/deploy"
	"taskrunner/pkg/utils"
	"taskrunner/trait"
)

type CoreConfig = deploy.CoreConfig

type ProtonTX struct {
	trait.Transaction
	DBConn
	*CoreConfig
}

type ProtonWrapperStore struct {
	trait.Store
	DBConn
	*CoreConfig
}

func NewStore(s trait.Store, cfg *CoreConfig) (*ProtonWrapperStore, *trait.Error) {
	conn, err := NewDBConn(s)
	if err != nil {
		return nil, err
	}
	return &ProtonWrapperStore{
		Store:      s,
		DBConn:     *conn,
		CoreConfig: cfg,
	}, nil
}

func (s *ProtonWrapperStore) Begin(ctx context.Context) (trait.Transaction, *trait.Error) {
	tx, err := s.Store.Begin(ctx)
	return &ProtonTX{
		Transaction: tx,
		DBConn:      s.DBConn,
		CoreConfig:  s.CoreConfig,
	}, err
}

func (s *ProtonWrapperStore) GetJobRecord(ctx context.Context, jid int) (trait.JobRecord, *trait.Error) {
	return getJobRecord(ctx, s.Store, s.DBConn, jid, s.CoreConfig)
}

func (s *ProtonTX) GetJobRecord(ctx context.Context, jid int) (trait.JobRecord, *trait.Error) {
	return getJobRecord(ctx, s.Transaction, s.DBConn, jid, s.CoreConfig)
}

func getJobRecord(
	ctx context.Context, reader trait.JobRecordReader,
	conn CursorConn, jid int,
	core *CoreConfig,
) (trait.JobRecord, *trait.Error) {
	jb, err := reader.GetJobRecord(ctx, jid)
	if err != nil {
		return jb, err
	}
	return jb, replaceProtonComponentMarco(ctx, conn, jb.Target, core)
}

func (s *ProtonWrapperStore) GetAPPIns(ctx context.Context, id int) (*trait.ApplicationInstance, *trait.Error) {
	return getAPPIns(ctx, s.Store, s.DBConn, id, s.CoreConfig)
}

func (s *ProtonTX) GetAPPIns(ctx context.Context, id int) (*trait.ApplicationInstance, *trait.Error) {
	return getAPPIns(ctx, s.Transaction, s.DBConn, id, s.CoreConfig)
}

func getAPPIns(
	ctx context.Context, reader trait.ApplicationInsReader,
	conn CursorConn, id int,
	core *CoreConfig,
) (*trait.ApplicationInstance, *trait.Error) {
	ains, err := reader.GetAPPIns(ctx, id)
	if err != nil {
		return nil, err
	}
	return ains, replaceProtonComponentMarco(ctx, conn, ains, core)
}

// GetWorkAPPIns replace componentProtonResourceTyp component with proton conf object
func (s *ProtonWrapperStore) GetWorkAPPIns(c context.Context, name string, sid int) (*trait.ApplicationInstance, *trait.Error) {
	return getWorkAPPIns(c, s.Store, s.DBConn, name, sid, s.CoreConfig)
}

// GetWorkAPPIns replace componentProtonResourceTyp component with proton conf object
func (s *ProtonTX) GetWorkAPPIns(c context.Context, name string, sid int) (*trait.ApplicationInstance, *trait.Error) {
	return getWorkAPPIns(c, s.Transaction, s.DBConn, name, sid, s.CoreConfig)
}

func getWorkAPPIns(
	c context.Context, reader trait.ApplicationInsReader,
	cli CursorConn, name string,
	sid int, core *CoreConfig,
) (*trait.ApplicationInstance, *trait.Error) {
	ains, err := reader.GetWorkAPPIns(c, name, sid)
	if err != nil {
		return nil, err
	}
	return ains, replaceProtonComponentMarco(c, cli, ains, core)
}

// GetWorkComponentIns replace componentProtonResourceTyp component with proton conf object
func (s *ProtonWrapperStore) GetWorkComponentIns(c context.Context, sid int, com trait.ComponentNode) (*trait.ComponentInstance, *trait.Error) {
	return getWorkComponentIns(c, s.Store, s.DBConn, sid, com, s.CoreConfig)
}

// GetWorkComponentIns replace componentProtonResourceTyp component with proton conf object
func (s *ProtonTX) GetWorkComponentIns(c context.Context, sid int, com trait.ComponentNode) (*trait.ComponentInstance, *trait.Error) {
	return getWorkComponentIns(c, s.Transaction, s.DBConn, sid, com, s.CoreConfig)
}

func getWorkComponentIns(
	ctx context.Context, reader trait.ComponentInsReader,
	conn CursorConn, sid int, com trait.ComponentNode,
	core *CoreConfig,
) (*trait.ComponentInstance, *trait.Error) {
	cins, err := reader.GetWorkComponentIns(ctx, sid, com)
	if err != nil {
		return nil, err
	}
	return cins, getProtonComponent(ctx, conn, cins, core)
}

func (s *ProtonWrapperStore) GetComponentIns(c context.Context, cid int) (*trait.ComponentInstance, *trait.Error) {
	return getComponentIns(c, s.Store, s.DBConn, cid, s.CoreConfig)
}

func (s *ProtonTX) GetComponentIns(c context.Context, cid int) (*trait.ComponentInstance, *trait.Error) {
	return getComponentIns(c, s.Transaction, s.DBConn, cid, s.CoreConfig)
}

func getComponentIns(ctx context.Context, reader trait.ComponentInsReader, cli CursorConn, cid int, core *CoreConfig) (*trait.ComponentInstance, *trait.Error) {
	cins, err := reader.GetComponentIns(ctx, cid)
	if err != nil {
		return nil, err
	}
	return cins, getProtonComponent(ctx, cli, cins, core)
}

// GetAPP replace component schema document
func (s *ProtonWrapperStore) GetAPP(ctx context.Context, aid int) (*trait.Application, *trait.Error) {
	return getAPP(ctx, s.Store, aid)
}

// GetAPP replace component schema document
func (s *ProtonTX) GetAPP(ctx context.Context, aid int) (*trait.Application, *trait.Error) {
	return getAPP(ctx, s.Transaction, aid)
}

func getAPP(ctx context.Context, reader trait.ApplicationReader, aid int) (*trait.Application, *trait.Error) {
	a, err := reader.GetAPP(ctx, aid)
	if err != nil {
		return nil, err
	}
	for _, c := range a.Component {
		if err := resources.ReplaceApplicationComponentSchame(c); err != nil {
			return nil, err
		}
	}
	return a, err
}

// GetAPPComponent replace component schema document
func (s *ProtonWrapperStore) GetAPPComponent(ctx context.Context, acid int) (*trait.ComponentMeta, *trait.Error) {
	return getAPPComponent(ctx, s.Store, acid)
}

// GetAPPComponent replace component schema document
func (s *ProtonTX) GetAPPComponent(ctx context.Context, acid int) (*trait.ComponentMeta, *trait.Error) {
	return getAPPComponent(ctx, s.Transaction, acid)
}

func getAPPComponent(ctx context.Context, reader trait.ApplicationReader, acid int) (*trait.ComponentMeta, *trait.Error) {
	c, err := reader.GetAPPComponent(ctx, acid)
	if err != nil {
		return nil, err
	}
	err = resources.ReplaceApplicationComponentSchame(c)
	return c, err
}

func replaceProtonComponentMarco(
	ctx context.Context,
	conn CursorConn,
	ains *trait.ApplicationInstance,
	core *CoreConfig,
) *trait.Error {
	if ains == nil {
		return nil
	}

	for _, cins := range ains.Components {
		if err := getProtonComponent(ctx, conn, cins, core); err != nil {
			return err
		}
	}

	return nil
}

func getProtonComponent(ctx context.Context, conn CursorConn, cins *trait.ComponentInstance, core *CoreConfig) *trait.Error {
	com := cins.Component
	if com.ComponentDefineType != component.ComponentProtonResourceType {
		return nil
	}

	bs := json.RawMessage([]byte{})
	if com.Type != resources.DeployCoreType {
		// deployCore类型为传入的固定值，不会存储于资源组件信息内，不需要获取信息
		err := conn.GetComponentInfo(ctx, com.Name, com.Type, cins.System.SID, ConnectObj{
			Info: &bs,
		})
		if trait.IsInternalError(err, trait.ErrNotFound) {
			err.Detail = fmt.Sprintf("component name: %s, type: %s system: %d not found", com.Name, com.Type, cins.System.SID)
		}
		if err != nil {
			return err
		}
	}

	attr, cfg, err := ComponentToValues(com.Type, bs, cins.System.SID, core)
	if err != nil {
		return err
	}
	cins.Attribute = utils.MergeMaps(cins.Attribute, attr)
	cins.Config = utils.MergeMaps(cins.Config, cfg)
	return nil
}

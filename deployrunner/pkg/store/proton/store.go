package store

import (
	"context"

	"taskrunner/pkg/component"
	"taskrunner/pkg/component/resources"
	"taskrunner/trait"
)

// Store wrrapper inner store, read ComponentProtonResourceType component from proton conf
type Store struct {
	trait.Store
	*deployConf
}

// InsertJobRecord check job config before insert jobrecord  into bankend.
func (s *Store) InsertJobRecord(ctx context.Context, j *trait.JobRecord) (int, *trait.Error) {
	return insertJobRecord(ctx, s.Store, s.deployConf, j)
}

// Begin wrrapper the inner store transaction
func (s *Store) Begin(c context.Context) (trait.Transaction, *trait.Error) {
	tx, err := s.Store.Begin(c)
	return &TX{Transaction: tx, deployConf: s.deployConf}, err
}

// GetJobRecord impl store interface and replaca target attribute
func (s *Store) GetJobRecord(ctx context.Context, jid int) (trait.JobRecord, *trait.Error) {
	return getJobRecord(ctx, s.Store, s.deployConf, jid)
}

// GetWorkComponentIns replace componentProtonResourceTyp component with proton conf object
func (s *Store) GetWorkComponentIns(c context.Context, sid int, com trait.ComponentNode) (*trait.ComponentInstance, *trait.Error) {
	return getWorkComponentIns(c, s.Store, s.deployConf, sid, com)
}

// GetComponentIns replace componentProtonResourceTyp component with proton conf object
func (s *Store) GetComponentIns(c context.Context, cid int) (*trait.ComponentInstance, *trait.Error) {
	return getComponentIns(c, s.Store, s.deployConf, cid)
}

// GetAPPIns replace componentProtonResourceTyp component with proton conf object
func (s *Store) GetAPPIns(c context.Context, id int) (*trait.ApplicationInstance, *trait.Error) {
	return getAPPIns(c, s.Store, s.deployConf, id)
}

// GetWorkAPPIns replace componentProtonResourceTyp component with proton conf object
func (s *Store) GetWorkAPPIns(c context.Context, name string, sid int) (*trait.ApplicationInstance, *trait.Error) {
	return getWorkAPPIns(c, s.Store, s.deployConf, name, sid)
}

// GetAPP replace component schema document
func (s *Store) GetAPP(ctx context.Context, aid int) (*trait.Application, *trait.Error) {
	return getAPP(ctx, s.Store, aid)
}

// GetAPPComponent replace component schema document
func (s *Store) GetAPPComponent(ctx context.Context, acid int) (*trait.ComponentMeta, *trait.Error) {
	return getAPPComponent(ctx, s.Store, acid)
}

// TX wrraper inner tx, read componentProtonResourceTyp component from proton conf
type TX struct {
	trait.Transaction
	*deployConf
}

// InsertJobRecord check job config before insert jobrecord  into bankend.
func (tx *TX) InsertJobRecord(ctx context.Context, j *trait.JobRecord) (int, *trait.Error) {
	return insertJobRecord(ctx, tx.Transaction, tx.deployConf, j)
}

// GetJobRecord impl store interface and replaca target attribute
func (tx *TX) GetJobRecord(ctx context.Context, jid int) (trait.JobRecord, *trait.Error) {
	return getJobRecord(ctx, tx.Transaction, tx.deployConf, jid)
}

// GetWorkComponentIns replace componentProtonResourceTyp component with proton conf object
func (tx *TX) GetWorkComponentIns(c context.Context, sid int, com trait.ComponentNode) (*trait.ComponentInstance, *trait.Error) {
	return getWorkComponentIns(c, tx.Transaction, tx.deployConf, sid, com)
}

// GetComponentIns replace componentProtonResourceTyp component with proton conf object
func (tx *TX) GetComponentIns(c context.Context, cid int) (*trait.ComponentInstance, *trait.Error) {
	return getComponentIns(c, tx.Transaction, tx.deployConf, cid)
}

// GetAPPIns replace componentProtonResourceTyp component with proton conf object
func (tx *TX) GetAPPIns(c context.Context, id int) (*trait.ApplicationInstance, *trait.Error) {
	return getAPPIns(c, tx.Transaction, tx.deployConf, id)
}

// GetWorkAPPIns replace componentProtonResourceTyp component with proton conf object
func (tx *TX) GetWorkAPPIns(c context.Context, name string, sid int) (*trait.ApplicationInstance, *trait.Error) {
	return getWorkAPPIns(c, tx.Transaction, tx.deployConf, name, sid)
}

// GetAPP replace component schema document
func (tx *TX) GetAPP(ctx context.Context, aid int) (*trait.Application, *trait.Error) {
	return getAPP(ctx, tx.Transaction, aid)
}

// GetAPPComponent replace component schema document
func (tx *TX) GetAPPComponent(ctx context.Context, acid int) (*trait.ComponentMeta, *trait.Error) {
	return getAPPComponent(ctx, tx.Transaction, acid)
}

func insertJobRecord(ctx context.Context, writer trait.JobRecordWriter, cli *deployConf, j *trait.JobRecord) (int, *trait.Error) {
	if err := replaceAttributes(ctx, cli, j.Target.Components); err != nil {
		return -1, err
	}
	return writer.InsertJobRecord(ctx, j)
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

func getAPPComponent(ctx context.Context, reader trait.ApplicationReader, acid int) (*trait.ComponentMeta, *trait.Error) {
	c, err := reader.GetAPPComponent(ctx, acid)
	if err != nil {
		return nil, err
	}
	err = resources.ReplaceApplicationComponentSchame(c)
	return c, err
}

func getJobRecord(ctx context.Context, reader trait.JobRecordReader, cli *deployConf, jid int) (trait.JobRecord, *trait.Error) {
	jb, err := reader.GetJobRecord(ctx, jid)
	if err != nil {
		return jb, err
	}
	jb.Target, err = replaceAPPInsComponentAttributes(ctx, cli, jb.Target)
	return jb, err
}

func getAPPIns(c context.Context, reader trait.ApplicationInsReader, cli *deployConf, id int) (*trait.ApplicationInstance, *trait.Error) {
	ains, err := reader.GetAPPIns(c, id)
	if err != nil {
		return nil, err
	}
	return replaceAPPInsComponentAttributes(c, cli, ains)
}

func getWorkAPPIns(c context.Context, reader trait.ApplicationInsReader, cli *deployConf, name string, sid int) (*trait.ApplicationInstance, *trait.Error) {
	ains, err := reader.GetWorkAPPIns(c, name, sid)
	if err != nil {
		return nil, err
	}
	return replaceAPPInsComponentAttributes(c, cli, ains)
}

func replaceAPPInsComponentAttributes(ctx context.Context, cli *deployConf, ains *trait.ApplicationInstance) (*trait.ApplicationInstance, *trait.Error) {
	replaces := []*trait.ComponentInstance{}
	for _, cins := range ains.Components {
		if cins.Component.ComponentDefineType == component.ComponentProtonResourceType {
			replaces = append(replaces, cins)
		}
	}
	return ains, replaceAttributesIgnoreNoInstalled(ctx, cli, replaces)
}

func getWorkComponentIns(ctx context.Context, reader trait.ComponentInsReader, cli *deployConf, sid int, com trait.ComponentNode) (*trait.ComponentInstance, *trait.Error) {
	cins, err := reader.GetWorkComponentIns(ctx, sid, com)
	if err != nil {
		return nil, err
	}
	return replaceAttribute(ctx, cli, cins, nil)
}

func getComponentIns(ctx context.Context, reader trait.ComponentInsReader, cli *deployConf, cid int) (*trait.ComponentInstance, *trait.Error) {
	cins, err := reader.GetComponentIns(ctx, cid)
	if err != nil {
		return nil, err
	}
	return replaceAttribute(ctx, cli, cins, nil)
}

package store

import (
	"context"
	"fmt"

	"taskrunner/pkg/component/resources"
	driver "taskrunner/pkg/sql-driver"
	"taskrunner/trait"

	kdb9 "taskrunner/pkg/store/mysql/driver/kdb9"
	mysql "taskrunner/pkg/store/mysql/driver/mysql"
)

// DBOP 业务无关,重构迁移到driver模块
type DBOP = driver.DBAdminOP

type Structor struct {
	DBType string
	Store  func(context.Context, resources.RDS) (trait.Store, *trait.Error)
}

var Structors = make(map[string]Structor)

// RegistryDriver registry without lock, this should use in func init
func RegistryDriver(s Structor) {
	dbType := driver.ConvertDBType(s.DBType)
	s.DBType = dbType
	Structors[dbType] = s
}

// NewDBOP create a db operator for init
func NewDBOP(ctx context.Context, rds resources.RDS) (DBOP, *trait.Error) {
	return driver.Factory.NewDBOP(ctx, rds)
}

func NewStore(ctx context.Context, rds resources.RDS) (trait.Store, *trait.Error) {
	// dbType := driver.ConvertDBType(rds.Type)
	// 差异仅在驱动层,实现内部初始化连接时会根据配置自行区分
	//  因此暂时固定操作对象为mariadb实现

	dbType := driver.ConvertDBType(rds.Type)
	s, ok := Structors[dbType]
	if !ok {
		// 懒得设置别名,未识别的统一是为mariadb模式
		dbType := driver.ConvertDBType(resources.MariadbDBType)
		s, ok := Structors[dbType]
		if !ok {
			return nil, &trait.Error{
				Internal: trait.ErrNotFound,
				Err:      fmt.Errorf("the input rds type %s not support", dbType),
			}
		}
		return s.Store(ctx, rds)
	}

	return s.Store(ctx, rds)
}

func init() {
	s := Structor{
		DBType: resources.MariadbDBType,
		Store: func(ctx context.Context, r resources.RDS) (trait.Store, *trait.Error) {
			s, err := mysql.NewStore(ctx, r)
			return s, err
		},
	}
	RegistryDriver(s)

	dm := Structor{
		DBType: resources.DM8DBType,
		Store: func(ctx context.Context, r resources.RDS) (trait.Store, *trait.Error) {
			s, err := mysql.NewStore(ctx, r)
			return s, err
		},
	}
	RegistryDriver(dm)

	kdb := Structor{
		DBType: resources.KDB9DBType,
		Store: func(ctx context.Context, r resources.RDS) (trait.Store, *trait.Error) {
			s, err := kdb9.NewStore(ctx, r)
			return s, err
		},
	}
	RegistryDriver(kdb)
}

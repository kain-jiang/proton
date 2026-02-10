package store

import (
	"context"

	"taskrunner/pkg/component/resources"
	mysql "taskrunner/pkg/store/mysql/driver/mysql"
	"taskrunner/trait"
)

type Store struct {
	*mysql.Store
}

// TX impl trait.Transaction
type TX struct {
	// trait.Transaction
	*mysql.TX
}

func NewStore(ctx context.Context, rds resources.RDS) (*Store, *trait.Error) {
	s, err := mysql.NewStore(ctx, rds)
	return &Store{Store: s}, err
}

// Begin imply trait.store return a transaction
func (s *Store) Begin(ctx context.Context) (trait.Transaction, *trait.Error) {
	tx, err := s.Store.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &TX{
		TX: tx.(*mysql.TX),
	}, err
}

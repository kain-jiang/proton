package v1alpha1

import "context"

type Interface interface {
	DatabaseInterface
	UserInterface
}

type DatabaseInterface interface {
	CreateDatabase(ctx context.Context, db *Database) error
	DeleteDatabase(ctx context.Context, name string) error
	ListDatabases(ctx context.Context) ([]Database, error)
}

type UserInterface interface {
	CreateUser(ctx context.Context, username, password string) error
	ListUsers(ctx context.Context) ([]User, error)
	PatchUserPrivileges(ctx context.Context, username string, privileges []Privilege) error
}

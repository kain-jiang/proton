package store

import (
	"context"
	"fmt"

	"taskrunner/trait"
)

// CreateUser create the sql instance user
// if the user has been creates, ensure the user has the password
func (db *DBOP) CreateUser(ctx context.Context, user, passwd string) *trait.Error {
	// fmt.Printf(db.Stmt.CreateUser+"\n", user, passwd)
	_, err := db.ExecContext(ctx, fmt.Sprintf(db.Stmt.CreateUser, user, passwd))
	if err != nil {
		return err
	}

	// fmt.Printf(db.Stmt.UpdateUser+"\n", user, passwd)
	_, err = db.ExecContext(ctx, fmt.Sprintf(db.Stmt.UpdateUser, user, passwd))
	return err
}

func (db *DBOP) DeleteUser(ctx context.Context, user string) *trait.Error {
	_, err := db.ExecContext(ctx, fmt.Sprintf(db.Stmt.RevokeUserDB, user))
	// fmt.Printf(db.Stmt.RevokeUserDB+"\n", user)
	if err != nil {
		return err
	}
	// fmt.Printf(db.Stmt.DeleteUser+"\n", user)
	_, err = db.ExecContext(ctx, fmt.Sprintf(db.Stmt.DeleteUser, user))
	return err
}

// GrantUserDB grant the sql instance user with database
func (db *DBOP) GrantUserDB(ctx context.Context, user, dbName string) *trait.Error {
	// fmt.Printf(db.Stmt.GrantUserDB+"\n", db.Stmt.RwPrivileges, dbName, user)
	_, err := db.ExecContext(ctx, fmt.Sprintf(db.Stmt.GrantUserDB, db.Stmt.RwPrivileges, dbName, user))
	return err
}

func (db *DBOP) deleteDatabase(ctx context.Context, dbName string) *trait.Error {
	_, err := db.ExecContext(ctx, fmt.Sprintf(db.Stmt.DeleteDataBaseStmt, dbName))
	return err
}

// CreateDatabase create database if the database is not exist
func (db *DBOP) CreateDatabase(ctx context.Context, dbName string) *trait.Error {
	_, err := db.ExecContext(ctx, fmt.Sprintf(db.Stmt.CreateDataBaseStmt, dbName))
	return err
}

func (db *DBOP) DeleteDatabase(ctx context.Context, dbName string) *trait.Error {
	return db.deleteDatabase(ctx, dbName)
}

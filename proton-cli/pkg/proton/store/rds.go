package store

import (
	"context"

	"github.com/sirupsen/logrus"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/rds/mgmt/v1alpha1"
)

// reconcileDatabase creates database if it not exists.
func (m *Manager) reconcileDatabase(c v1alpha1.DatabaseInterface) error {
	databases, err := c.ListDatabases(context.TODO())
	if err != nil {
		return err
	}

	for _, db := range databases {
		if db.DBName == DatabaseName {
			m.Logger.WithField("name", DatabaseName).Debug("database already exists")
			return nil
		}
	}

	database := &v1alpha1.Database{DBName: DatabaseName, Charset: DatabaseCharset, Collation: DatabaseCollation}
	m.Logger.WithFields(logrus.Fields{
		"name":    database.DBName,
		"charset": database.Charset,
		"collate": database.Collation,
		"object":  database,
	}).Info("create database")
	return c.CreateDatabase(context.TODO(), database)
}

// reconcileDatabaseUserPrivileges
func (m *Manager) reconcileDatabaseUserPrivileges(c v1alpha1.UserInterface, username string) error {
	users, err := c.ListUsers(context.TODO())
	if err != nil {
		return err
	}

	var privileges []v1alpha1.Privilege

	for _, u := range users {
		if u.Username != username {
			continue
		}
		privileges = u.Privileges
		break
	}

	for _, p := range privileges {
		if p.DBName != DatabaseName {
			continue
		}
		if p.PrivilegeType != RDSUserPrivilege {
			break
		}
		m.Logger.WithFields(logrus.Fields{
			"username":  username,
			"database":  DatabaseName,
			"privilege": RDSUserPrivilege,
		}).Debug("user's privileges are already satisfied")
		return nil
	}

	privileges = []v1alpha1.Privilege{{DBName: DatabaseName, PrivilegeType: RDSUserPrivilege}}

	m.Logger.WithFields(logrus.Fields{
		"username":   username,
		"privileges": privileges,
	}).Info("patch database users privileges")
	return c.PatchUserPrivileges(context.TODO(), username, privileges)
}

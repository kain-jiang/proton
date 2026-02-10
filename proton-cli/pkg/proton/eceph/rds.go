package eceph

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"

	rds_mgmt_v1alpha1 "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/rds/mgmt/v1alpha1"
)

func (m *Manager) reconcileDatabase() error {
	if !m.InitDatabase {
		m.Logger.Info("This cluster is using external RDS database for ECeph, please create required databases manually.")
		return nil
	}
	m.Logger.Info("reconcile database")
	c, err := m.RDS_MGMTClient()
	if err != nil {
		return err
	}

	var database = &rds_mgmt_v1alpha1.Database{
		DBName:    ECephDatabaseName,
		Charset:   rds_mgmt_v1alpha1.CharsetUTF8MB4,
		Collation: rds_mgmt_v1alpha1.CollationUTF8MB4GeneralCI,
	}

	databases, err := c.ListDatabases(context.TODO())
	if err != nil {
		return err
	}

	for _, d := range databases {
		if d.DBName != database.DBName {
			continue
		}
		if d.Charset != database.Charset {
			return fmt.Errorf("the charset of database %q should be %q instead of %q", database.DBName, database.Charset, d.Charset)
		}
		if d.Collation != database.Collation {
			// ignore a specific error screnaio
			if d.DBName == ECephDatabaseName && d.Collation == "utf8mb4_unicode_ci" {
				errStr := fmt.Sprintf("the collation of database %q should be %q instead of %q", database.DBName, database.Collation, d.Collation)
				m.Logger.WithField("name", d.DBName).Warningf("ignoring the error of %s", errStr)
			} else {
				return fmt.Errorf("the collation of database %q should be %q instead of %q", database.DBName, database.Collation, d.Collation)
			}
		}

		m.Logger.WithField("name", d.DBName).Debug("skip creating proton rds database")
		return nil
	}

	m.Logger.WithFields(logrus.Fields{
		"name":      database.DBName,
		"charset":   database.Charset,
		"collation": database.Collation,
	}).Info("create proton rds database")
	return c.CreateDatabase(context.TODO(), database)
}

func (m *Manager) reconcileDatabaseUser() error {
	if !m.InitDatabase {
		m.Logger.Info("This cluster is using external RDS database for ECeph, please create required users manually.")
		return nil
	}
	m.Logger.Info("reconcile database user")

	c, err := m.RDS_MGMTClient()
	if err != nil {
		return err
	}

	users, err := c.ListUsers(context.TODO())
	if err != nil {
		return err
	}

	var user *rds_mgmt_v1alpha1.User
	for _, u := range users {
		if u.Username == m.RDS.Username {
			user = &u
			break
		}
	}

	// create user if not exist
	if user != nil {
		m.Logger.WithField("name", m.RDS.Username).Debug("skip creating database user")
	} else {
		m.Logger.WithField("name", m.RDS.Username).Info("create database user")
		user = &rds_mgmt_v1alpha1.User{Username: m.RDS.Username}
		if err := c.CreateUser(context.TODO(), m.RDS.Username, m.RDS.Password); err != nil {
			return err
		}
	}

	// expected privilege
	var privilege = rds_mgmt_v1alpha1.Privilege{DBName: ECephDatabaseName, PrivilegeType: rds_mgmt_v1alpha1.PrivilegeReadWrite}

	if slices.Contains(user.Privileges, privilege) {
		m.Logger.WithFields(logrus.Fields{
			"user":      user.Username,
			"database":  privilege.DBName,
			"privilege": privilege.PrivilegeType,
		}).Debug("skip patching database user privileges")
		return nil
	}

	m.Logger.WithFields(logrus.Fields{
		"user":      user.Username,
		"database":  privilege.DBName,
		"privilege": privilege.PrivilegeType,
	}).Info("patch database user privileges")
	return c.PatchUserPrivileges(context.TODO(), user.Username, []rds_mgmt_v1alpha1.Privilege{privilege})
}

func (m *Manager) RDS_MGMTClient() (rds_mgmt_v1alpha1.Interface, error) {
	if m.rdsMGMTClient != nil {
		return m.rdsMGMTClient, nil
	}
	m.Logger.Debug("create rds mgmt client")
	c, err := m.RDS_MGMTClientCreateFunc()
	if err != nil {
		return nil, err
	}
	m.rdsMGMTClient = c
	return c, nil
}

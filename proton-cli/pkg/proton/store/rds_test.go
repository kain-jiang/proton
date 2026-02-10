package store

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/rds/mgmt/v1alpha1"
	v1alpha_testing "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/rds/mgmt/v1alpha1/testing"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/logger"
)

func TestManager_reconcileDatabase(t *testing.T) {
	var (
		errList   = errors.New("list: something wrong")
		errCreate = errors.New("create: something wrong")
	)
	tests := []struct {
		name    string
		c       v1alpha1.DatabaseInterface
		wantErr error
	}{
		{
			name: "create",
			c:    v1alpha_testing.NewClient(),
		},
		{
			name: "already exists",
			c:    v1alpha_testing.NewClient(v1alpha1.Database{DBName: DatabaseName}),
		},
		{
			name:    "list fail",
			c:       v1alpha_testing.NewClient().WithMethodError("ListDatabases", errList),
			wantErr: errList,
		},
		{
			name:    "create fail",
			c:       v1alpha_testing.NewClient().WithMethodError("CreateDatabase", errCreate),
			wantErr: errCreate,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manager{Logger: logger.NewLogger()}
			err := m.reconcileDatabase(tt.c)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestManager_reconcileDatabaseUserPrivileges(t *testing.T) {
	var (
		userWithCorrectPrivilege   = &v1alpha1.User{Username: "hello", Privileges: []v1alpha1.Privilege{{DBName: DatabaseName, PrivilegeType: RDSUserPrivilege}}}
		userWithIncorrectPrivilege = &v1alpha1.User{Username: "hello", Privileges: []v1alpha1.Privilege{{DBName: DatabaseName, PrivilegeType: "WrongPrivilege"}}}
		userWithoutPrivilege       = &v1alpha1.User{Username: "hello", Privileges: []v1alpha1.Privilege{{DBName: "other", PrivilegeType: RDSUserPrivilege}}}
	)
	tests := []struct {
		name             string
		c                v1alpha1.UserInterface
		username         string
		wantErrSubString string
	}{
		{
			name:     "add privilege",
			c:        v1alpha_testing.NewClient(userWithoutPrivilege),
			username: "hello",
		},
		{
			name:     "update privilege",
			c:        v1alpha_testing.NewClient(userWithIncorrectPrivilege),
			username: "hello",
		},
		{
			name:     "already satisfied",
			c:        v1alpha_testing.NewClient(userWithCorrectPrivilege),
			username: "hello",
		},
		{
			name:             "list users fail",
			c:                v1alpha_testing.NewClient().WithMethodError("ListUsers", errors.New("list: something wrong")),
			wantErrSubString: "list: something wrong",
		},
		{
			name:             "user not found",
			c:                v1alpha_testing.NewClient(v1alpha1.User{Username: "other"}),
			wantErrSubString: "404020000: 资源不存在",
		},
		{
			name:             "patch user privileges fail",
			c:                v1alpha_testing.NewClient(userWithIncorrectPrivilege).WithMethodError("PatchUserPrivileges", errors.New("patch: something wrong")),
			username:         "hello",
			wantErrSubString: "patch: something wrong",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manager{Logger: logger.NewLogger()}
			err := m.reconcileDatabaseUserPrivileges(tt.c, tt.username)
			if tt.wantErrSubString != "" {
				assert.ErrorContains(t, err, tt.wantErrSubString)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

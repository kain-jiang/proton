package testing

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"golang.org/x/exp/slices"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/rds/mgmt/v1alpha1"
)

// client is a fake client in memory.
type client struct {
	databases    map[string]v1alpha1.Database
	databaseLock sync.RWMutex

	users    map[string]v1alpha1.User
	userLock sync.RWMutex

	// method while return these errors. Key is method name.
	errors map[string]error
}

func NewClient(objects ...interface{}) *client {
	var (
		databases = make(map[string]v1alpha1.Database)
		users     = make(map[string]v1alpha1.User)
	)
	for _, o := range objects {
		switch v := o.(type) {
		case v1alpha1.Database:
			databases[v.DBName] = v1alpha1.Database{DBName: v.DBName, Charset: v.Charset, Collation: v.Collation}
		case *v1alpha1.Database:
			databases[v.DBName] = v1alpha1.Database{DBName: v.DBName, Charset: v.Charset, Collation: v.Collation}
		case v1alpha1.User:
			user := v1alpha1.User{Username: v.Username, Privileges: make([]v1alpha1.Privilege, len(v.Privileges)), SSLType: v.SSLType}
			copy(user.Privileges, v.Privileges)
			users[v.Username] = user
		case *v1alpha1.User:
			user := v1alpha1.User{Username: v.Username, Privileges: make([]v1alpha1.Privilege, len(v.Privileges)), SSLType: v.SSLType}
			copy(user.Privileges, v.Privileges)
			users[v.Username] = user
		default:
			panic(fmt.Sprintf("unsupported type %T of %q", v, v))
		}
	}
	return &client{databases: databases, users: users, errors: make(map[string]error)}
}

// CreateDatabase implements v1alpha1.Interface.
func (c *client) CreateDatabase(ctx context.Context, db *v1alpha1.Database) error {
	if err := c.CurrentMethodError(); err != nil {
		return err
	}

	c.databaseLock.Lock()
	defer c.databaseLock.Unlock()

	if _, ok := c.databases[db.DBName]; ok {
		return &v1alpha1.Error{Code: 403020000, Message: "禁止执行", Cause: "数据库已存在无法重复创建"}
	}

	c.databases[db.DBName] = v1alpha1.Database{
		DBName:    db.DBName,
		Charset:   db.Charset,
		Collation: db.Collation,
	}
	return nil
}

// ListDatabases implements v1alpha1.Interface.
func (c *client) ListDatabases(ctx context.Context) ([]v1alpha1.Database, error) {
	if err := c.CurrentMethodError(); err != nil {
		return nil, err
	}

	c.databaseLock.RLock()
	defer c.databaseLock.RUnlock()

	var databases []v1alpha1.Database
	for _, d := range c.databases {
		databases = append(databases, d)
	}

	return databases, nil
}

// CreateUser implements v1alpha1.Interface.
func (c *client) CreateUser(ctx context.Context, username string, password string) error {
	if err := c.CurrentMethodError(); err != nil {
		return err
	}

	c.userLock.Lock()
	defer c.userLock.Unlock()

	if _, ok := c.users[username]; ok {
		return &v1alpha1.Error{Code: 403020000, Message: "禁止执行", Cause: "待创建的账户已存在"}
	}

	c.users[username] = v1alpha1.User{Username: username}
	return nil
}

// DeleteDatabase implements v1alpha1.Interface.
func (*client) DeleteDatabase(ctx context.Context, name string) error {
	panic("unimplemented")
}

// ListUsers implements v1alpha1.Interface.
func (c *client) ListUsers(ctx context.Context) ([]v1alpha1.User, error) {
	if err := c.CurrentMethodError(); err != nil {
		return nil, err
	}

	c.userLock.RLock()
	defer c.userLock.RUnlock()

	for _, u := range c.users {
		fmt.Printf("DEBUG:ListUsers: %#v\n", u)
	}

	var users []v1alpha1.User
	for _, u := range c.users {
		users = append(users, u)
	}

	return users, nil
}

// PatchUserPrivileges implements v1alpha1.Interface.
func (c *client) PatchUserPrivileges(ctx context.Context, username string, privileges []v1alpha1.Privilege) error {
	if err := c.CurrentMethodError(); err != nil {
		return err
	}

	c.userLock.Lock()
	defer c.userLock.Unlock()

	for _, u := range c.users {
		fmt.Printf("DEBUG:PatchUserPrivileges: %#v\n", u)
	}

	user, ok := c.users[username]
	if !ok {
		return &v1alpha1.Error{Code: 404020000, Message: "资源不存在", Cause: "用户不存在"}
	}

	for _, p := range privileges {
		i := slices.IndexFunc(user.Privileges, func(up v1alpha1.Privilege) bool { return up.DBName == p.DBName })
		if i < 0 {
			user.Privileges = append(user.Privileges, v1alpha1.Privilege{DBName: p.DBName, PrivilegeType: p.PrivilegeType})
			continue
		}
		user.Privileges[i].PrivilegeType = p.PrivilegeType
	}

	return nil
}

var _ v1alpha1.Interface = (*client)(nil)

func (c *client) WithMethodError(method string, err error) *client {
	c.errors[method] = err
	return c
}

func (c *client) CurrentMethodError() error {
	pc, _, _, ok := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)

	if !ok || fn == nil {
		return nil
	}
	name := fn.Name()
	index := strings.LastIndex(name, ".")
	if index < 0 {
		return nil
	}

	return c.errors[name[index+1:]]
}

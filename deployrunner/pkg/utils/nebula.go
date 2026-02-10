package utils

import (
	"context"
	"fmt"

	nebula "github.com/vesoft-inc/nebula-go/v3"
)

// SimpleNebulaClient simple client for create user
type SimpleNebulaClient struct {
	Logger nebula.DefaultLogger
	Host   string
	Port   int
	User   string
	Passwd string
}

// CreateUserWithPasswd create  the user w and set with the passwd
func (c *SimpleNebulaClient) CreateUserWithPasswd(ctx context.Context, user, passwd string) error {
	hostAddress := nebula.HostAddress{Host: c.User, Port: c.Port}
	hostList := []nebula.HostAddress{hostAddress}
	// Create configs for connection pool using default values
	testPoolConfig := nebula.GetDefaultConf()

	// Initialize connection pool
	pool, err := nebula.NewConnectionPool(hostList, testPoolConfig, c.Logger)
	if err != nil {
		return err
	}
	// Close all connections in the pool
	defer pool.Close()

	// Create session
	session, err := pool.GetSession(c.User, c.Passwd)
	if err != nil {
		return err
	}
	// Release session and return connection back to connection pool
	defer session.Release()

	params := map[string]interface{}{
		"user":   user,
		"passwd": passwd,
	}
	res, err := session.ExecuteWithParameter("CREATE USER [IF NOT EXISTS] $user;", params)
	if err != nil {
		return err
	}
	if !res.IsSucceed() {
		return fmt.Errorf("create nebule user error: %d: %s", res.GetErrorCode(), res.GetErrorMsg())
	}
	res, err = session.ExecuteWithParameter("ALTER USER $user WITH PASSWORD $passwd", params)
	if err != nil {
		return err
	}
	if !res.IsSucceed() {
		return fmt.Errorf("set user password error: %d: %s", res.GetErrorCode(), res.GetErrorMsg())
	}
	return nil
}

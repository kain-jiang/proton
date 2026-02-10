package task

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"taskrunner/pkg/component/resources"
	"taskrunner/pkg/log"
	"taskrunner/test"
	"taskrunner/trait"

	"bou.ke/monkey"
	"github.com/sirupsen/logrus"
)

func TestResource(t *testing.T) {
	tt := test.TestingT{T: t}
	ins := &trait.ComponentInstance{
		ComponentInstanceMeta: trait.ComponentInstanceMeta{
			Component: trait.ComponentNode{
				Type: resources.RDSType,
			},
		},
		Attribute: map[string]interface{}{
			"host":        "123456",
			"port":        float64(123),
			"source_type": "internal",
			"mgmt_host":   "123456",
			"mgmt_port":   123,
		},
	}
	ts := newProtonResourceTask(ins, nil)
	ts.WithLog(log.NewTaskLogger(logrus.New(), logrus.FatalLevel, 10))
	count := 1
	total := 2
	monkey.Patch(net.DialTimeout, func(_, _ string, _ time.Duration) (net.Conn, error) {
		if count < total {
			count++
			return nil, fmt.Errorf("test")
		}
		return &net.TCPConn{}, nil
	})

	{
		ctx, cancel := context.WithCancel(context.Background())
		err := ts.Install(ctx)
		tt.AssertNil(err)
		cancel()
		count = 1
		if err := ts.Install(ctx); err == nil {
			t.Fatal("shuold error")
			t.FailNow()
		}
	}

	{
		ins.Component.Type = resources.MongodbType
		err := ts.Install(context.Background())
		tt.AssertNil(err)
	}
}

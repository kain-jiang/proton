package grafana

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/node/v1alpha1/fake"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
)

func TestManager_checkEnvironment(t *testing.T) {
	const (
		host             = "host-example"
		dataPath         = "/var/lib/grafana"
		storageClassName = "standard"
	)
	var (
		logger = &logrus.Logger{Out: os.Stdout, Hooks: make(logrus.LevelHooks), Formatter: new(logrus.TextFormatter), Level: logrus.DebugLevel}
	)
	tests := []struct {
		name    string
		m       *Manager
		wantErr bool
	}{
		{
			name: "success",
			m: &Manager{
				Spec:   &configuration.Grafana{Hosts: []string{host}, DataPath: dataPath},
				Node:   fake.NewForTesting(t, host, []string{dataPath}, nil),
				Logger: logger,
			},
		},
		{
			name: "failure",
			m: &Manager{
				Spec:   &configuration.Grafana{Hosts: []string{host}, DataPath: dataPath},
				Node:   fake.NewForTesting(t, host, nil, []string{dataPath}),
				Logger: logger,
			},
			wantErr: true,
		},
		{
			name: "without host and data path",
			m: &Manager{
				Spec:   &configuration.Grafana{StorageClassName: storageClassName},
				Logger: logger,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Skip("unimplemented")
		})
	}
}

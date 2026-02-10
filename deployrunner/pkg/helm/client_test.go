package helm

import (
	"context"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/cli"
)

func TestClisntChart(t *testing.T) {
	t.SkipNow()
	fin, _ := os.Open("/tmp/work/installer-sync-3.0.0-feature-657500.tgz")
	c, _ := ParseChartFromTGZ(fin, "v2")
	cfg := cli.New()
	log := logrus.New()
	cli := NewHelm3Client(log, &EnvSettings{
		EnvSettings: cfg,
		Force:       true,
	})
	err := cli.Install(context.Background(), "installer-sync", "anyshare", c, nil, 10, log.Infof)
	if err != nil {
		t.Fatal(err.Error())
	}
}

package helm3

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/registry"
)

type OCIRegistryConfig struct {
	PlainHTTP bool
	Registry  string
	Username  string
	Password  string
}

func (r OCIRegistryConfig) OCIUrl() string {
	return fmt.Sprintf("oci://%s", r.Registry)
}

func (r OCIRegistryConfig) OCIChartUrl(name string) string {
	return fmt.Sprintf("%s/%s", strings.TrimRight(r.OCIUrl(), "/"), name)
}

func (c *helmv3) loginRegistry(reg *OCIRegistryConfig) error {
	if reg.Username != "" {
		// login
		regClient, err := registry.NewClient(registry.ClientOptCredentialsFile(c.settings.RegistryConfig))
		if err != nil {
			c.log.WithError(err).Errorln("cannot create oci registry client")
			return fmt.Errorf("cannot create oci registry client: %w", err)
		}
		err = regClient.Login(
			reg.Registry,
			registry.LoginOptInsecure(false),
			registry.LoginOptBasicAuth(reg.Username, reg.Password),
		)
		if err != nil {
			c.log.WithError(err).Errorln("cannot login oci registry")
			return fmt.Errorf("cannot login oci registry: %w", err)
		}
	}
	return nil
}

func (c *helmv3) PushChart(f string, reg *OCIRegistryConfig) error {
	err := c.loginRegistry(reg)
	if err != nil {
		return fmt.Errorf("login registry failed: %w", err)
	}

	push := action.NewPushWithOpts(action.WithPlainHTTP(reg.PlainHTTP), action.WithPushConfig(c.actionConfig))
	_, err = push.Run(f, reg.OCIUrl())
	if err != nil {
		return fmt.Errorf("push chart failed: %w", err)
	}
	return nil
}

func (c *helmv3) PullChart(name, version string, reg *OCIRegistryConfig) (string, func(), error) {
	err := c.loginRegistry(reg)
	if err != nil {
		return "", func() {}, fmt.Errorf("login registry failed: %w", err)
	}
	pull := action.NewPullWithOpts(action.WithConfig(c.actionConfig))
	dir, err := os.MkdirTemp("", fmt.Sprintf("%s-*", name))
	if err != nil {
		return "", func() {}, fmt.Errorf("create temp dir failed: %w", err)
	}
	cleaner := func() {
		_ = os.RemoveAll(dir)
	}
	pull.PlainHTTP = true
	pull.Version = version
	pull.Settings = c.settings
	pull.DestDir = dir
	_, err = pull.Run(reg.OCIChartUrl(name))
	if err != nil {
		cleaner()
		return "", func() {}, fmt.Errorf("pull chart failed: %w", err)
	}
	return filepath.Join(dir, fmt.Sprintf("%s-%s.tgz", name, version)), cleaner, nil
}

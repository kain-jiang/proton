package helm

import (
	"context"
	"fmt"
	"time"

	"taskrunner/trait"

	"github.com/pkg/errors"

	"github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"
)

// Client helm operation
type Client interface {
	Install(ctx context.Context, name, ns string, chart *Chart, cfg map[string]interface{}, timeout int, log action.DebugLog) *trait.Error
	Uninstall(ctx context.Context, name, ns string, timeout int, log action.DebugLog) *trait.Error
	Values(ctx context.Context, name, ns string) (map[string]interface{}, *trait.Error)
}

type helmConfig struct {
	Force           bool
	CreateNamespace bool
	Wait            bool
}

// Helm3Client helm3 client
// TODO: unit test
type Helm3Client struct {
	setting *cli.EnvSettings
	log     *logrus.Logger
	cfg     helmConfig
}

type EnvSettings struct {
	Force           bool
	CreateNamespace bool
	*cli.EnvSettings
}

// NewHelm3Client return a helm3 client
func NewHelm3Client(log *logrus.Logger, cfg *EnvSettings) *Helm3Client {
	// copy from helm/v3/pkg/cli.New
	hcfg := helmConfig{
		Force:           cfg.Force,
		CreateNamespace: cfg.CreateNamespace,
		Wait:            true,
	}
	return &Helm3Client{setting: cfg.EnvSettings, log: log, cfg: hcfg}
}

// func (c *Helm3Client) WithForce(force bool) {
// 	c.cfg.Force = force
// }

func (c *Helm3Client) WithCreateNamespace(create bool) {
	c.cfg.CreateNamespace = create
}

func (c *Helm3Client) WithWait(wait bool) {
	c.cfg.Wait = wait
}

// Values get the release currnet values
func (c *Helm3Client) Values(ctx context.Context, name, ns string) (map[string]interface{}, *trait.Error) {
	actCfg, err0 := c.actionConfig(ns, nil)
	if err0 != nil {
		return nil, err0
	}
	getter := action.NewGet(actCfg)

	rls, err := getter.Run(name)
	if err == driver.ErrReleaseNotFound {
		return nil, &trait.Error{
			Internal: trait.ErrNotFound,
			Err:      err,
			Detail:   err,
		}
	}
	if err != nil {
		return nil, &trait.Error{
			Internal: trait.ECHelmK8s,
			Err:      err,
			Detail:   fmt.Sprintf("get helm release %s values", name),
		}
	}
	return rls.Config, nil
}

func (c *Helm3Client) actionConfig(namespace string, log action.DebugLog) (*action.Configuration, *trait.Error) {
	actionConfig := &action.Configuration{}
	if log == nil {
		log = c.log.Tracef
	}
	if err := actionConfig.Init(c.setting.RESTClientGetter(), namespace, "", log); err != nil {
		return nil, &trait.Error{
			Internal: trait.ECHelmK8s,
			Detail:   "unknow error when create helm action config",
			Err:      err,
		}
	}

	hkcli := newkubeClient(actionConfig.KubeClient.(*kube.Client))
	hkcli.Namespace = namespace
	actionConfig.KubeClient = hkcli
	return actionConfig, nil
}

// Install imply Client interface
func (c *Helm3Client) Install(ctx context.Context, name, ns string, chart *Chart, cfg map[string]interface{}, timeout int, log action.DebugLog) *trait.Error {
	actionConfig, err0 := c.actionConfig(ns, log)
	if err0 != nil {
		return err0
	}
	runner := action.NewInstall(actionConfig)
	runner.ReleaseName = name
	runner.Timeout = time.Second * time.Duration(timeout)
	runner.CreateNamespace = c.cfg.CreateNamespace
	runner.Wait = c.cfg.Wait
	// runner.WaitForJobs = true
	runner.Namespace = ns
	// runner.Atomic = true
	runner.DisableOpenAPIValidation = true
	runner.Force = c.cfg.Force

	rel, err := runner.RunWithContext(ctx, chart.v2, cfg)
	if err != nil {
		if err == driver.ErrReleaseExists || err.Error() == "cannot re-use a name that is still in use" {
			return c.upgrade(ctx, actionConfig, name, ns, chart, cfg, timeout)
		}
		if rel == nil {
			return &trait.Error{
				Internal: trait.ECHelmK8s,
				Err:      err,
				Detail:   fmt.Sprintf("install helm chart %s:%s with k8s operate error", name, chart.v2.Metadata.Version),
			}
		} else {
			if rel.Manifest == "" && len(rel.Hooks) == 0 {
				return &trait.Error{
					Internal: trait.ECTemplate,
					Err:      err,
					Detail:   fmt.Sprintf("install helm chart %s:%s error when render resource", name, chart.v2.Metadata.Version),
				}
			} else {
				return &trait.Error{
					Internal: trait.ECHelmRun,
					Err:      err,
					Detail:   fmt.Sprintf("install helm chart %s:%s in status: %s, description: %s", name, chart.v2.Metadata.Version, rel.Info.Status, rel.Info.Description),
				}
			}
		}
	}

	return nil
}

func (c *Helm3Client) getLastRelease(_ context.Context, actionConfig *action.Configuration, name string) (*release.Release, error) {
	getter := action.NewGet(actionConfig)
	return getter.Run(name)
}

func (c *Helm3Client) upgrade(ctx context.Context, actionConfig *action.Configuration, name, ns string, chart *Chart, cfg map[string]interface{}, timeout int) *trait.Error {
	rls, err := c.getLastRelease(ctx, actionConfig, name)
	if err == driver.ErrReleaseNotFound {
		// don't do anything
		err = nil
	} else if isReleaseNotfound(err) {
		err = nil
	} else if err != nil {
		// abort
		return &trait.Error{
			Internal: trait.ECHelmReleaseNotFound,
			Err:      err,
			Detail:   fmt.Sprintf("upgrade chart release %s:%s when get last release error", name, chart.v2.Metadata.Version),
		}
	} else {
		// cfg = utils.MergeMaps(rls.Config, cfg)
		if rls.Info.Status.IsPending() {
			c.log.Warnf("the helm %s task may be operting by other user or the relase is disaster interruption, current release status is %s, tr overwrite and load by this task", name, rls.Info.Status.String())
			rls.SetStatus(release.StatusDeployed, "set deploy ill be re run and over write by deploy-installer")
			if err = actionConfig.Releases.Update(rls); err != nil {
				c.log.Errorf("the helm %s task try to overwrite status error: %s", name, err.Error())
				return &trait.Error{
					Internal: trait.ECHelmReleaseForceUpdate,
					Err:      err,
					Detail:   fmt.Sprintf("upgrade chart release %s:%s when force change last release upgrading status", name, chart.v2.Metadata.Version),
				}
			}
		}
	}

	runner := action.NewUpgrade(actionConfig)
	runner.Timeout = time.Second * time.Duration(timeout)
	runner.Wait = c.cfg.Wait
	// runner.WaitForJobs = true
	runner.Namespace = ns
	// runner.Atomic = true
	runner.DisableOpenAPIValidation = true
	runner.Force = c.cfg.Force

	rel, err := runner.RunWithContext(ctx, name, chart.v2, cfg)
	if err != nil {
		if rel == nil {
			return &trait.Error{
				Internal: trait.ECHelmK8SORRender,
				Err:      err,
				Detail:   fmt.Sprintf("install helm chart %s:%s with k8s operate error", name, chart.v2.Metadata.Version),
			}
		} else if err0 := trait.UnwrapError(err); err0 != nil {
			return err0
		} else {
			return &trait.Error{
				Internal: trait.ECHelmRun,
				Err:      err,
				Detail:   fmt.Sprintf("install helm chart %s:%s in status: %s, description: %s", name, chart.v2.Metadata.Version, rel.Info.Status, rel.Info.Description),
			}
		}
	}
	return nil
}

// Uninstall imply Client interface
func (c *Helm3Client) Uninstall(ctx context.Context, name, ns string, timeout int, log action.DebugLog) *trait.Error {
	actionConfig, err0 := c.actionConfig(ns, log)
	if err0 != nil {
		return err0
	}
	runner := action.NewUninstall(actionConfig)
	runner.Timeout = time.Second * time.Duration(timeout)
	_, err := runner.Run(name)
	if err != nil {
		if isReleaseNotfound(err) {
			return nil
		} else if err.Error() == "no release provided" {
			err = nil
		} else if err0 := trait.UnwrapError(err); err0 != nil {
			return err0
		} else {
			return &trait.Error{
				Internal: trait.ECHelmRun,
				Err:      err,
				Detail:   fmt.Sprintf("uninstall helm release %s error", name),
			}
		}
	}
	return nil
}

func isReleaseNotfound(err error) bool {
	err0 := errors.Cause(err)
	return err0 == driver.ErrNoDeployedReleases || err0 == driver.ErrReleaseNotFound
}

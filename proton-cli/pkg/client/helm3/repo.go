package helm3

import (
	"fmt"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/repo"
)

func FetchChartFromChartmuseum(entry *repo.Entry, name, version string) (*chart.Chart, error) {
	/**
	  action.ChartPathOptions{RepoURL: c.repo.URL(), Version: version},
	*/
	cpo := action.ChartPathOptions{
		RepoURL:  entry.URL,
		Username: entry.Username,
		Password: entry.Password,
		Version:  version,
	}
	fPath, err := cpo.LocateChart(name, cli.New())
	if err != nil {
		return nil, err
	}
	c, err := loader.Load(fPath)
	if err != nil {
		return nil, err
	}
	if version != "" && c.Metadata.Version != version {
		return c, fmt.Errorf(
			`chart "%s" version "%s" not found in %s repository, but found version "%s"`,
			name,
			version,
			entry.URL,
			c.Metadata.Version,
		)
	}
	return c, nil
}

func FetchChart(cli Client, entry *repo.Entry, reg *OCIRegistryConfig, name, version string) (*chart.Chart, error) {
	if entry != nil {
		return FetchChartFromChartmuseum(entry, name, version)
	}
	if reg != nil {
		fPath, cleaner, err := cli.PullChart(name, version, reg)
		defer cleaner()
		if err != nil {
			return nil, err
		}
		return loader.Load(fPath)
	}
	return nil, fmt.Errorf("cannot find chartmuseum or oci registry")
}

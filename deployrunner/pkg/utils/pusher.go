package utils

import (
	"fmt"
	"net/http"
	"regexp"

	"helm.sh/helm/v3/pkg/pusher"
	"helm.sh/helm/v3/pkg/registry"

	cm "github.com/chartmuseum/helm-push/pkg/chartmuseum"
	ch "github.com/chartmuseum/helm-push/pkg/helm"
)

func PuserPushChart(chartPath string, repo, username, password string) error {
	if regexp.MustCompile(`^https?://`).MatchString(repo) {
		// Chartmuseum 仓库
		repo, err := ch.TempRepoFromURL(repo)
		if err != nil {
			return err
		}
		client, err := cm.NewClient(
			cm.URL(repo.Config.URL),
			cm.InsecureSkipVerify(true),
			cm.Username(username),
			cm.Password(password),
		)
		if err != nil {
			return err
		}

		resp, err := client.UploadChartPackage(chartPath, true)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusCreated {
			return fmt.Errorf("upload chart package failed: %s", resp.Status)
		}
		return nil
	}
	if regexp.MustCompile(`^oci://`).MatchString(repo) {
		// Oci 仓库
		reg, err := registry.NewClient(
			registry.ClientOptBasicAuth(username, password),
			registry.ClientOptPlainHTTP(),
		)
		if err != nil {
			return err
		}
		pr, err := pusher.NewOCIPusher(pusher.WithRegistryClient(reg))
		if err != nil {
			return err
		}
		return pr.Push(chartPath, repo)
	}
	return fmt.Errorf("repo: [%s] must be http/https or oci", repo)
}

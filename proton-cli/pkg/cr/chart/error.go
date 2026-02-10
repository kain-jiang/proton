package chart

import (
	"errors"
	"strings"

	"helm.sh/helm/v3/pkg/repo"
)

func IsNotFound(err error) bool {
	if errors.Is(err, repo.ErrNoChartName) || errors.Is(err, repo.ErrNoChartVersion) {
		return true
	}

	return strings.Contains(err.Error(), "no chart version found for")
}

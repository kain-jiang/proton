package keepalived

import (
	"encoding/json"
	"errors"

	"github.com/go-test/deep"
	"github.com/hashicorp/go-version"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage/driver"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3"
)

type HelmRelease struct {
	Name    string
	Version *version.Version
	Values  *HelmValues
}

func NewHelmRelease(rls *release.Release) (*HelmRelease, error) {
	if rls == nil {
		return nil, nil
	}
	v, err := version.NewVersion(rls.Chart.Metadata.Version)
	if err != nil {
		return nil, err
	}
	valuesObj := new(HelmValues)
	valuesJson, err := json.Marshal(rls.Config)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(valuesJson, &valuesObj)
	if err != nil {
		return nil, err
	}

	return &HelmRelease{
		Name:    rls.Name,
		Version: v,
		Values:  valuesObj,
	}, nil
}

// GetHelmRelease 返回 rds keepalived 的 helm release，兼容 anysharectl 部署的 helm release。
func GetHelmRelease(helm3 helm3.Client) (rls *release.Release, err error) {
	for _, n := range []string{
		ReleaseName,
		ReleaseNameAnyShareCTLCreated,
	} {
		rls, err = helm3.GetRelease(n)
		if !errors.Is(err, driver.ErrReleaseNotFound) {
			break
		}
	}
	return
}

// NeedUpgrade 返回 helm release 是否需要升级
func (r *HelmRelease) NeedUpgrade(expect *HelmValues) bool {
	return r.Version.LessThan(MinimumVersion) || deep.Equal(r.Values, expect) != nil
}

package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"taskrunner/pkg/component"
	"taskrunner/pkg/helm"
	"taskrunner/pkg/utils"
	"taskrunner/trait"

	"helm.sh/helm/v3/pkg/chart/loader"
)

func (e *Executor) PushImages(ctx context.Context, ociFile string) ([]map[string]string, *trait.Error) {
	images, err := utils.OrasListTags(ociFile)
	if err != nil {
		return nil, &trait.Error{
			Internal: trait.ECNULL,
			Err:      fmt.Errorf("list tags failed by oras: %w", err),
			Detail:   "get images from oci file error",
		}
	}

	pcfg, perr := e.pcli.GetConf(ctx)
	if perr != nil {
		return nil, perr
	}

	registeris := pcfg.GetRegistries()
	if len(registeris) == 0 {
		return nil, &trait.Error{
			Internal: trait.ECNULL,
			Err:      fmt.Errorf("no registries found in proton conf"),
			Detail:   "get registries from proton conf error",
		}
	}

	result := make([]map[string]string, 0)

	for _, rh := range registeris {
		for _, image := range images {
			srcImage := image
			segments := strings.Split(image, "/")
			segments[0] = rh.Registry
			dstImage := strings.Join(segments, "/")
			err := utils.OrasPushImage(
				ociFile,
				srcImage,
				dstImage,
				rh.Username,
				rh.Password,
			)
			if err != nil {
				return nil, &trait.Error{
					Internal: trait.ECNULL,
					Err:      err,
					Detail:   fmt.Sprintf("push image %s to registry to %s falied", srcImage, dstImage),
				}
			}
			result = append(result, map[string]string{
				"from": srcImage,
				"to":   dstImage,
			})
		}
	}

	return result, nil
}

func (e *Executor) PushChart(ctx context.Context, chartData []byte) (map[string]any, map[string]any, *trait.Error) {
	c, err := loader.LoadArchive(bytes.NewReader(chartData))
	if err != nil {
		return nil, nil, &trait.Error{
			Internal: trait.ECNULL,
			Err:      fmt.Errorf("load chart failed: %w", err),
			Detail:   "load chart by chart loader failed",
		}
	}
	cp := &component.HelmComponent{}
	cp.Name = c.Metadata.Name
	cp.Version = c.Metadata.Version
	terr := e.HelmRepo.Store(ctx, cp, chartData)
	if terr != nil {
		return nil, nil, terr
	}

	cmb, err := json.Marshal(c.Metadata)
	if err != nil {
		return nil, nil, &trait.Error{
			Internal: trait.ECNULL,
			Err:      fmt.Errorf("marshal chart metadata failed: %w", err),
			Detail:   "marshal chart metadata failed",
		}
	}

	var cht map[string]any
	err = json.Unmarshal(cmb, &cht)
	if err != nil {
		return nil, nil, &trait.Error{
			Internal: trait.ECNULL,
			Err:      fmt.Errorf("unmarshal chart metadata failed: %w", err),
			Detail:   "unmarshal chart metadata failed",
		}
	}

	return cht, c.Values, nil
}

func (e *Executor) InstallRelease(
	ctx context.Context,
	name, cname, cversion string,
	ns string, setRegistry bool,
	values map[string]any,
) (map[string]any, *trait.Error) {
	cp := &component.HelmComponent{}
	cp.Name = cname
	cp.Version = cversion
	bs, terr := e.HelmRepo.Fetch(ctx, cp)
	if terr != nil {
		return nil, terr
	}

	c, terr := helm.ParseChartFromTGZ(bytes.NewReader(bs), "v2") // 强制使用 v2
	if terr != nil {
		return nil, terr
	}

	if setRegistry {
		values = utils.MergeMaps(values, map[string]any{
			"image": e.imageRepo.ToMap(),
		})
	}

	terr = e.helmCli.Install(ctx, name, ns, c, values, 600, e.Log.Debugf)
	if terr != nil {
		return nil, terr
	}

	rv, terr := e.helmCli.Values(ctx, name, ns)
	if terr != nil {
		return nil, terr
	}

	return rv, nil
}

func (e *Executor) UninstallRelease(ctx context.Context, name, ns string) (map[string]any, *trait.Error) {
	rv, terr := e.helmCli.Values(ctx, name, ns)
	if terr != nil {
		return nil, terr
	}
	terr = e.helmCli.Uninstall(ctx, name, ns, 600, e.Log.Debugf)
	if terr != nil {
		return nil, terr
	}
	return rv, nil
}

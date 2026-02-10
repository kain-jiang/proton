package task

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"taskrunner/pkg/cluster"
	"taskrunner/pkg/component"
	"taskrunner/pkg/helm"
	"taskrunner/trait"
)

// HelmTask a helm task
type HelmTask struct {
	System *cluster.SystemContext
	// Job          trait.JobRecord
	HelmComponent *component.HelmComponent
	Base
}

// Install impl task interface
func (h *HelmTask) Install(ctx context.Context) *trait.Error {
	cfg, err := h.config()
	if err != nil {
		return err
	}

	bs, err := h.System.HelmRepo.Fetch(ctx, h.HelmComponent)
	if err != nil {
		return err
	}
	c, err := helm.ParseChartFromTGZ(bytes.NewReader(bs), h.HelmComponent.HelmChartAPIVersion)
	if err != nil {
		return err
	}

	timeout := 600
	if h.ComponentInsData.Timeout > 0 {
		timeout = h.ComponentInsData.Timeout
	}

	ctx0, cancel := trait.WithTimeoutCauseContext(ctx, time.Second*time.Duration(timeout+1), &trait.Error{
		Internal: trait.ECHelmTimeout,
		Err:      context.DeadlineExceeded,
		Detail:   fmt.Sprintf("install helm chart %s timeout", h.HelmComponent.Name),
	})
	defer cancel()

	// TODO upgrade and install
	return h.System.HelmClient.Install(ctx0, h.HelmComponent.Name, h.System.NameSpace, c, cfg, timeout, h.Log.Debugf)
}

// Uninstall impl task interface
func (h *HelmTask) Uninstall(ctx context.Context) *trait.Error {
	timeout := 600
	if h.ComponentInsData.Timeout > 0 {
		timeout = h.ComponentInsData.Timeout
	}

	ctx0, cancel := trait.WithTimeoutCauseContext(ctx, time.Second*time.Duration(timeout+1), &trait.Error{
		Internal: trait.ECHelmTimeout,
		Err:      context.DeadlineExceeded,
		Detail:   fmt.Sprintf("uninstall helm release %s timeout", h.HelmComponent.Name),
	})
	defer cancel()
	return h.System.HelmClient.Uninstall(ctx0, h.HelmComponent.Name, h.System.NameSpace, timeout, h.Log.Debugf)
}

func (h *HelmTask) config() (map[string]interface{}, *trait.Error) {
	// v := chartValues{
	// 	AppConfig:          h.Base.appConfig,
	// 	ComponentConfig:    h.Base.ComponentInsData.Config,
	// 	Topology:           make(map[string]topology),
	// 	ComponentAttribute: h.Base.ComponentInsData.Attribute,
	// }

	attributes := make(map[string]interface{}, len(h.Topology)+1)
	deployTraits := make(map[string]interface{}, len(h.Topology)+1)
	cins := h.Base.ComponentInsData
	if cins.Attribute == nil {
		attributes[cins.Component.Name] = map[string]interface{}{}
	} else {
		attributes[cins.Component.Name] = attributes
	}
	attributes[cins.Component.Name] = cins.Attribute
	dTrait := map[string]interface{}{
		"deployTrait": cins.GetMiniTrait(),
	}
	deployTraits[cins.Component.Name] = dTrait

	for _, c := range h.Topology {
		if c.Attribute == nil {
			attributes[c.Component.Name] = map[string]interface{}{}
		} else {
			attributes[c.Component.Name] = c.Attribute
		}
		deployTraits[c.Component.Name] = map[string]interface{}{
			"deployTrait": c.GetMiniTrait(),
		}
	}

	deps := mergeMaps(deployTraits, attributes)

	return mergeMaps(cins.GetDeployTrait(), h.System.ToMap(), h.Base.ComponentInsData.AppConfig,
		h.Base.ComponentInsData.Config, map[string]interface{}{"depServices": deps}), nil
}

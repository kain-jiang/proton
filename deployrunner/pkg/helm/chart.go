package helm

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"

	"taskrunner/trait"

	yaml3 "gopkg.in/yaml.v3"
	"helm.sh/helm/v3/pkg/action"
	chartv2 "helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// ChartMeta helm chart.yaml  meta
type ChartMeta struct {
	APIVersion string `json:"apiVersion"`
	Version    string `json:"version"`
	Name       string `json:"name"`
}

// Chart deal helm chart info
type Chart struct {
	// v1 *chartv1.Chart
	v2 *chartv2.Chart
}

// ParseChartFromTGZ load the chart info from tgz bytes reader
func ParseChartFromTGZ(reader io.Reader, version string) (*Chart, *trait.Error) {
	c := &Chart{}
	switch version {
	// case "v1":
	// 	chart, err := chartutil2.LoadArchive(reader)
	// 	c.v1 = chart
	// 	return c, err
	case "v2", "v1":
		chart, err := loader.LoadArchive(reader)
		c.v2 = chart
		if err != nil {
			return nil, &trait.Error{
				Internal: trait.ECParseChart,
				Err:      err,
				Detail:   "helm load chart from tgz",
			}
		}
		return c, nil
	default:
		return nil, &trait.Error{
			Internal: trait.ErrHelmChartAPIVersion,
			Err:      fmt.Errorf("chart apiVersion '%s' 不支持", version),
			Detail:   "chart version not support",
		}

	}
}

// Render get manifest
func (c *Chart) Render() (string, *trait.Error) {
	// if c.v1 != nil {
	// 	return c.renderV1()
	// }
	return c.reanderV2()
}

func (c *Chart) reanderV2() (string, *trait.Error) {
	client := action.NewInstall(&action.Configuration{})
	client.DryRun = true
	client.ReleaseName = "release-name"
	client.Replace = true // Skip the name check
	client.ClientOnly = true
	rel, err := client.Run(c.v2, nil)
	if err != nil {
		return "", &trait.Error{
			Internal: trait.ECTemplate,
			Err:      err,
			Detail:   fmt.Sprintf("try render helm chart %s:%s", c.v2.Name(), c.v2.Metadata.Version),
		}
	}
	// manifest := strings.TrimSpace(rel.Manifest)
	buf := bytes.NewBuffer(nil)
	fmt.Fprintln(buf, strings.TrimSpace(rel.Manifest))
	for _, m := range rel.Hooks {
		fmt.Fprintf(buf, "---\n# Source: %s\n%s\n", m.Path, m.Manifest)
	}
	// fmt.Printf("%#v:\n%s", c.v2.Metadata, manifest)
	return buf.String(), nil
}

// func (c *Chart) renderV1() (string, error) {
// 	opt := renderutil.Options{
// 		APIVersions: []string{},
// 		KubeVersion: "v1.23.4",
// 		ReleaseOptions: chartutil2.ReleaseOptions{
// 			Name:      "builder",
// 			IsInstall: true,
// 			IsUpgrade: false,
// 			Namespace: "buidler",
// 		},
// 	}
// 	templates, err := renderutil.Render(c.v1, nil, opt)
// 	if err != nil {
// 		return "", err
// 	}

// 	sources := manifest.SplitManifests(templates)
// 	buf := bytes.NewBuffer(nil)

// 	for _, m := range sources {
// 		b := filepath.Base(m.Name)
// 		if b == "NOTES.txt" {
// 			continue
// 		}
// 		if strings.HasPrefix(b, "_") {
// 			continue
// 		}
// 		if _, err := fmt.Fprintf(buf, "---\n# Source: %s\n", m.Name); err != nil {
// 			return "", err
// 		}
// 		if _, err := fmt.Fprintln(buf, m.Content); err != nil {
// 			return "", err
// 		}
// 	}

// 	return buf.String(), nil
// }

func (c *Chart) Name() string {
	return c.v2.Metadata.Name
}

// Images return images list
func (c *Chart) Images() ([]string, *trait.Error) {
	imgIndex := make(map[string]bool)
	m, err := c.Render()
	if err != nil {
		return nil, err
	}
	yamlDec := yaml.NewYAMLReader(bufio.NewReader(strings.NewReader(m)))
	for {
		objs := map[string]interface{}{}
		bs, err := yamlDec.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, &trait.Error{
				Internal: trait.ECTemplate,
				Err:      err,
				Detail:   fmt.Sprintf("get helm chart %s:%s images, yaml decode manifest error", c.v2.Name(), c.v2.Metadata.Version),
			}
		}

		if err = yaml3.Unmarshal(bs, &objs); err != nil {
			return nil, &trait.Error{
				Internal: trait.ECTemplate,
				Err:      err,
				Detail:   fmt.Sprintf("get helm chart %s:%s images, yaml decode sub obj error", c.v2.Name(), c.v2.Metadata.Version),
			}
		}

		c.parseImage(objs, imgIndex)
	}

	res := make([]string, 0, len(imgIndex))
	for img := range imgIndex {
		res = append(res, img)
	}
	return res, nil
}

// TODO Confirm supported versions and objects
// obj struct instead of map[string]intreface
func (c *Chart) parseImage(obj map[string]interface{}, result map[string]bool) {
	containerSpecPath := []string{"spec", "template", "spec"}
	cur := obj
	// walk to the container parent path
	for _, key := range containerSpecPath {
		o, ok := cur[key]
		if !ok {
			return
		}

		nextObj, ok := o.(map[string]interface{})
		if !ok {
			return
		}
		cur = nextObj
	}

	gotImages := func(key string) {
		if containerList, ok := cur[key]; ok {
			if cs, ok := containerList.([]interface{}); ok {
				for _, c := range cs {
					container, ok := c.(map[string]interface{})
					if !ok {
						return
					}
					if img, ok := container["image"]; ok {
						if imgStr, ok := img.(string); ok {
							result[imgStr] = true
						}
					}
				}
			}
		}
	}

	gotImages("containers")
	gotImages("initContainers")
}

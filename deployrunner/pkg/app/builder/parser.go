package builder

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"taskrunner/pkg/component"
	"taskrunner/pkg/helm"
	"taskrunner/pkg/utils"
	"taskrunner/trait"

	"github.com/ghodss/yaml"
)

// ParseApplication parse tgz into application
func ParseApplication(r io.Reader) (trait.Application, map[*tar.Header][]byte, *trait.Error) {
	var a trait.Application
	br, err := utils.NewTGZReader(r)
	if err != nil {
		return a, nil, &trait.Error{
			Internal: trait.ErrApplicationFile,
			Err:      err,
			Detail:   "load application from tgz stream when parse application file",
		}
	}
	loadMeta := false
	// file bytes
	fb := make(map[*tar.Header][]byte)
	for {
		h, err := br.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return a, nil, &trait.Error{
				Internal: trait.ErrApplicationFile,
				Err:      err,
				Detail:   "read application bytes from tgz stream when parse application file",
			}
		}

		switch h.Name {
		case trait.AppMetaFile:
			loadMeta = true
			d := json.NewDecoder(br)
			err := d.Decode(&a)
			if err != nil {
				return a, nil, &trait.Error{
					Internal: trait.ErrApplicationFile,
					Err:      err,
					Detail:   "decode application metadata file error when parse application file",
				}
			}
		default:
			// cache := make([]byte, h.Size)
			// _, err = br.Read(cache)
			cache, err := io.ReadAll(br)
			if err != nil {
				return a, nil, &trait.Error{
					Internal: trait.ErrApplicationFile,
					Err:      err,
					Detail:   "load nomal chart file when parse application",
				}
			}

			fb[h] = cache
		}

	}

	if !loadMeta {
		return a, fb, &trait.Error{
			Internal: trait.ErrApplicationFile,
			Err:      fmt.Errorf("%s file not found in pacakge", trait.AppMetaFile),
			Detail:   "parse application from tgz file",
		}
	}

	return a, fb, nil
}

// ParseHelmChartMeta parse helm info from chart name
func ParseHelmChartMeta(chartName string) *component.HelmComponent {
	// chart path = helm_charts/<repo>/<chartName>-<chartVersion>.tgz
	if !strings.HasPrefix(chartName, trait.HelmChartDir) {
		return nil
	}

	if !strings.HasSuffix(chartName, ".tgz") {
		return nil
	}
	c := &component.HelmComponent{}
	items := strings.Split(strings.Trim(chartName, "/"), "/")
	l := len(items)
	if l < 3 {
		return nil
	}

	fName := strings.TrimSuffix(items[l-1], ".tgz")
	// helm_charts/<repo>
	items = items[:l-1]
	// repo
	c.Repository = strings.Join(items[1:], "/")
	// <chartName>-<chartVersion>.tgz
	rex, _ := regexp.Compile(`\d+\.\d+\.\d+`)
	loc := rex.FindStringIndex(fName)
	if len(loc) < 1 {
		return nil
	}
	c.Name = fName[:loc[0]-1]
	c.Version = fName[loc[0]:]

	return c
}

// ParseHelmChart parse helm component info from  helm chart
func ParseHelmChart(hc *component.HelmComponent, bs []byte, out *utils.TGZWriter, graph []trait.Edge) ([]trait.Edge, string, *trait.Error) {
	apiVersion := ""
	var err0 *trait.Error

	tr, err := utils.NewTGZReader(bytes.NewReader(bs))
	if err != nil {
		return graph, apiVersion, &trait.Error{
			Internal: trait.ECNULL,
			Err:      err,
			Detail:   fmt.Sprintf("解压chart'%s:%s',格式错误,联系chart开发人员确认chart格式", hc.Name, hc.Version),
		}
	}

	rootDir := filepath.SplitList(hc.Name)[0]
	metaPath := fmt.Sprintf("%s/%s", rootDir, trait.HelmDefinedPath)

	chartYamlPath := fmt.Sprintf("%s/%s", rootDir, trait.HelmChartPath)

	w := out

	for {
		h, err := tr.Next()
		if err == io.EOF {
			// if err := w.Flush(); err != nil {
			// 	err = fmt.Errorf("写内存失败,发生错误: %s,确认环境磁盘与内存无异常后重试", err.Error())
			// 	return apiVersion, err
			// }
			break
		}
		if err != nil {
			return graph, apiVersion, &trait.Error{
				Internal: trait.ECNULL,
				Err:      err,
				Detail:   fmt.Sprintf("解压chart'%s:%s'错误,联系chart开发人员确认chart格式", hc.Name, hc.Version),
			}
		}

		bs, err := io.ReadAll(tr)
		if err != nil {
			return graph, apiVersion, &trait.Error{
				Internal: trait.ECNULL,
				Err:      err,
				Detail:   fmt.Errorf("解压chart'%s:%s'错误, 联系chart开发人员确认chart格式", hc.Name, hc.Version),
			}
		}

		switch h.Name {
		case metaPath:
			helmComponent := &component.HelmComponent{}
			if err := json.Unmarshal(bs, &helmComponent); err != nil {
				return graph, apiVersion, &trait.Error{
					Internal: trait.ECNULL,
					Err:      err,
					Detail:   fmt.Sprintf("'%s'的chart中%s文件不符合格式要求,联系chart开发人员确认", hc.Name, metaPath),
				}
			}
			helmComponent.ComponentNode = hc.ComponentNode
			hc.ComponentMeta = helmComponent.ComponentMeta
			hc.Images = append(hc.Images, helmComponent.Images...)

			graph, err0 = hc.ComponentMeta.AddEdgeInto(graph)
			if err0 != nil {
				return graph, apiVersion, &trait.Error{
					Internal: trait.ECNULL,
					Err:      err0,
					Detail:   fmt.Sprintf("%s中依赖添加失败,联系组件开发人员修复", h.Name),
				}
			}

			// // skip this file
			// continue
		case chartYamlPath:
			cMeta := &helm.ChartMeta{}
			if err := yaml.Unmarshal(bs, cMeta); err != nil {
				return graph, apiVersion, &trait.Error{
					Internal: trait.ECNULL,
					Err:      err,
					Detail:   fmt.Sprintf("'%s'的chart中%s文件不符合格式要求, 联系chart开发人员确认", hc.Name, chartYamlPath),
				}
			}

			if cMeta.Name == "" || cMeta.APIVersion == "" || cMeta.Version == "" {
				return graph, apiVersion, &trait.Error{
					Internal: trait.ECNULL,
					Err:      err,
					Detail:   fmt.Sprintf("'%s'的chart中%s文件不符合格式要求,遇到错误: 其内必填值为空, 联系chart开发人员确认", hc.Name, chartYamlPath),
				}
			}
			hc.ComponentMeta.Version = cMeta.Version
			hc.ComponentMeta.Name = cMeta.Name
			apiVersion = cMeta.APIVersion
			hc.HelmChartAPIVersion = cMeta.APIVersion
		}

		if w == nil {
			continue
		}
		if err = w.WriteHeader(h); err != nil {
			return graph, apiVersion, &trait.Error{
				Internal: trait.ECNULL,
				Err:      err,
				Detail:   fmt.Sprintf("compress helm tgz header error:__--%s--__", err.Error()),
			}
		}

		if _, err = w.Write(bs); err != nil {
			return graph, apiVersion, &trait.Error{
				Internal: trait.ECNULL,
				Err:      err,
				Detail:   fmt.Sprintf("compress helm error:__--%s--__", err.Error()),
			}
		}

	}
	return graph, apiVersion, nil
}

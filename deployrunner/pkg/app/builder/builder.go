package builder

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"taskrunner/pkg/component"
	"taskrunner/pkg/graph"
	"taskrunner/pkg/graph/task"
	"taskrunner/pkg/helm"
	helmrepo "taskrunner/pkg/helm/repos"
	"taskrunner/pkg/utils"
	"taskrunner/trait"

	"github.com/sirupsen/logrus"
)

// Components components in builder meta file
type Components struct {
	HelmComponents     map[string][]*component.HelmComponent `json:"helm_charts,omitempty"`
	ResourceComponents []*component.HoleComponent            `json:"resources,omitempty"`
}

// Configuration containe app package info and builder env
type Configuration struct {
	Components        `json:"components"`
	trait.Application `json:",inline"`
	HelmRepos         []*helmrepo.HTTPHelmRepo `json:"helmRepos,omitempty"`
	LocalRepo         *helmrepo.Local          `json:"localRepo,omitempty"`
}

// LoadConfiguration load config from reader
func LoadConfiguration(r io.Reader) (Configuration, *trait.Error) {
	d := json.NewDecoder(r)
	cfg := Configuration{}
	err := d.Decode(&cfg)
	if err != nil {
		logrus.Error(err)
		return cfg, &trait.Error{
			Internal: trait.ErrApplicationFile,
			Err:      err,
			Detail:   "json decode application developer config",
		}
	}
	return cfg, nil
}

// NewApplicationBuilder create a app builder
func NewApplicationBuilder(cfg *Configuration, w io.Writer, imgout io.Writer, repos ...helm.Repo) (*Builder, *trait.Error) {
	log := logrus.New()
	log.SetReportCaller(true)

	b := &Builder{
		Log:       log,
		w:         w,
		imagesOut: imgout,
	}
	b.cfg = cfg
	b.helmRepos = helm.NewHelmIndexRepo(repos...)
	a := cfg.Application
	switch a.Type {
	case trait.AppBetav1Type:
	default:
		err := fmt.Errorf("appDefinedType: %s, is not suppport", a.Type)
		b.Log.Error(err)
		return b, &trait.Error{
			Internal: trait.ErrAppTypeNoDefined,
			Err:      err,
		}
	}

	return b, nil
}

// Builder read application defiend and build a package
type Builder struct {
	Log *logrus.Logger
	// a         *trait.Application
	cfg                *Configuration
	tw                 *utils.TGZWriter
	w                  io.Writer
	helmRepos          *helm.Repos
	imagesOut          io.Writer
	ConfigTemplatePath string
}

func (b *Builder) writeMeta() *trait.Error {
	bw := b.tw
	metaBs, err := json.Marshal(b.cfg.Application)
	if err != nil {
		return &trait.Error{
			Internal: trait.ECNULL,
			Err:      err,
			Detail:   "json marshal application",
		}
	}
	err = bw.WriteHeader(&tar.Header{
		Name: trait.AppMetaFile,
		Size: int64(len(metaBs)),
	})
	if err != nil {
		return &trait.Error{
			Internal: trait.ECNULL,
			Err:      err,
			Detail:   "writeMeta header into tgz",
		}
	}

	_, err = bw.Write(metaBs)
	if err != nil {
		return &trait.Error{
			Internal: trait.ECNULL,
			Err:      err,
			Detail:   "writeMeta body into tgz",
		}
	}

	return nil
}

func (b *Builder) packageHelmComponent(ctx context.Context, hc *component.HelmComponent) *trait.Error {
	bs, err := b.downloadHelmComponent(ctx, hc)
	if err != nil {
		b.Log.Errorf("下载chart %s:%s失败,error: %#v", hc.Name, hc.Version, err)
		return err
	}

	err = b.parseAndWriteHelmChart(hc, bs)
	if err != nil {
		b.Log.Errorf("解析chart %s:%s失败, error: %#v", hc.Name, hc.Version, err)
		return err
	}
	return nil
}

func (b *Builder) dumpHelmCompoentImages(buf io.Reader, c *component.HelmComponent, apiVersion string) *trait.Error {
	chart, err := helm.ParseChartFromTGZ(buf, apiVersion)
	if err != nil {
		return err
	}
	imgs, err := chart.Images()
	if err != nil {
		b.Log.Error(err)
		return err
	}

	for _, img := range c.Images {
		if _, err := fmt.Fprintf(b.imagesOut, "%s\n", img); err != nil {
			err = fmt.Errorf("镜像输出失败, 请检查环境，错误信息:%s", err.Error())
			b.Log.Error(err)
			return &trait.Error{
				Internal: trait.ECNULL,
				Err:      err,
				Detail:   "",
			}
		}
	}

	for _, img := range imgs {
		if _, err := fmt.Fprintf(b.imagesOut, "%s\n", img); err != nil {
			err = fmt.Errorf("镜像输出失败, 请检查环境，错误信息:%s", err.Error())
			b.Log.Error(err)
			return &trait.Error{
				Internal: trait.ECNULL,
				Err:      err,
				Detail:   "",
			}
		}
	}

	return nil
}

func (b *Builder) parseAndWriteHelmChart(c *component.HelmComponent, bs []byte) *trait.Error {
	buf := bytes.NewBuffer(nil)
	w := utils.NewTGzWriter(buf)
	defer w.Close()

	graph, apiVersion, err := ParseHelmChart(c, bs, w, b.cfg.Application.Graph)
	if err != nil {
		return err
	}
	if c.ComponentDefineType == component.ComponentHelmServiceType || c.ComponentDefineType == component.ComponentHelmTaskType {
		b.cfg.Application.Graph = graph
		spec, err := c.HelmComponentSpec.Encode()
		if err != nil {
			err = fmt.Errorf("%s 组件定义错误，序列化失败: %s", c.Name, err.Error())
			return &trait.Error{
				Internal: trait.ECComponentDefined,
				Err:      err,
				Detail:   "",
			}
		}
		c.ComponentMeta.Spec = spec
		b.cfg.Application.Component = append(b.cfg.Application.Component, &c.ComponentMeta)
	}
	if c.ComponentDefineType == component.ComponentHoleType ||
		c.ComponentDefineType == component.ComponentHelmAddtionalType {
		b.cfg.Application.Component = append(b.cfg.Application.Component, &c.ComponentMeta)
	}

	if err := w.Close(); err != nil {
		err = fmt.Errorf("compress helm error:__--%s--__", err.Error())
		return &trait.Error{
			Internal: trait.ECNULL,
			Err:      err,
			Detail:   "",
		}
	}

	bs = buf.Bytes()

	fpath := fmt.Sprintf("%s%s/%s-%s.tgz", trait.HelmChartDir, c.Repository, c.Name, c.Version)
	if err := b.tw.WriteHeader(
		&tar.Header{
			Name: fpath,
			Size: int64(len(bs)),
		},
	); err != nil {
		err = fmt.Errorf("写磁盘失败,发生错误：%s, 确认环境与磁盘无异常后重试", err.Error())
		return &trait.Error{
			Internal: trait.ECNULL,
			Err:      err,
			Detail:   "",
		}
	}

	if _, err := b.tw.Write(bs); err != nil {
		err = fmt.Errorf("写磁盘失败,发生错误：%s, 确认环境与磁盘无异常后重试", err.Error())
		return &trait.Error{
			Internal: trait.ECNULL,
			Err:      err,
			Detail:   "",
		}
	}

	return b.dumpHelmCompoentImages(bytes.NewReader(bs), c, apiVersion)
}

func (b *Builder) downloadHelmComponent(ctx context.Context, c *component.HelmComponent) ([]byte, *trait.Error) {
	bs, err := b.helmRepos.Fetch(ctx, c)
	if trait.IsInternalError(err, trait.ErrHelmRepoNoFound) {
		return nil, err
	} else if err != nil {
		return nil, err
	}

	return bs, err
}

func (b *Builder) buildHelmConponent(ctx context.Context) *trait.Error {
	count := 0
	for repo, hcs := range b.cfg.Components.HelmComponents {
		for _, hc := range hcs {
			hc.Repository = repo
			if err := b.packageHelmComponent(ctx, hc); err != nil {
				return err
			}
			count++
		}
	}

	b.Log.Tracef("total build %d helm compoent", count)
	return nil
}

func (b *Builder) buildHoleConponent(_ context.Context) *trait.Error {
	a := b.cfg.Application
	for _, c := range b.cfg.Components.ResourceComponents {
		a.Component = append(a.Component, &c.ComponentMeta)
		graph, err := c.AddEdgeInto(a.Graph)
		if err != nil {
			b.Log.Errorf("%s组件依赖声明错误,联系组件负责人修复: %s", c.Name, err.Error())
			return err
		}
		a.Graph = graph
	}
	b.cfg.Application = a
	return nil
}

func (b *Builder) WriteConfigTemplate() *trait.Error {
	if b.ConfigTemplatePath == "" {
		return nil
	}
	err := filepath.Walk(b.ConfigTemplatePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return &trait.Error{
				Err:      err,
				Detail:   path,
				Internal: trait.ECNULL,
			}
		}

		if !info.IsDir() {
			// If it's a file, read its content
			bs, err := os.ReadFile(path)
			if err != nil {
				return &trait.Error{
					Err:      err,
					Detail:   path,
					Internal: trait.ECNULL,
				}
			}
			cfg := &trait.AppliacationConfigTemplate{}
			if err := json.Unmarshal(bs, cfg); err != nil {
				return &trait.Error{
					Err:      err,
					Detail:   string(bs),
					Internal: trait.ErrParam,
				}
			}
			if err := cfg.Validate(); err != nil {
				return err
			}

			fpath := filepath.Join(trait.ConfigTemplateDir, path)
			if err := b.tw.WriteHeader(
				&tar.Header{
					Name: fpath,
					Size: int64(len(bs)),
				},
			); err != nil {
				err = fmt.Errorf("写磁盘失败,发生错误：%s, 确认环境与磁盘无异常后重试", err.Error())
				return &trait.Error{
					Internal: trait.ECNULL,
					Err:      err,
					Detail:   "",
				}
			}

			if _, err := b.tw.Write(bs); err != nil {
				err = fmt.Errorf("写磁盘失败,发生错误：%s, 确认环境与磁盘无异常后重试", err.Error())
				return &trait.Error{
					Internal: trait.ECNULL,
					Err:      err,
					Detail:   "",
				}
			}
		}

		return nil
	})
	if err != nil {
		if err0 := trait.UnwrapError(err); err0 != nil {
			return err0
		}
		return &trait.Error{
			Internal: trait.ECNULL,
			Err:      err,
		}
	}
	return nil
}

func (b *Builder) buildBetaV1(ctx context.Context) *trait.Error {
	buf := bytes.NewBuffer(nil)
	b.tw = utils.NewTGzWriter(buf)
	defer func() {
		b.tw.Flush()
		b.tw.Close()
		b.tw = nil
	}()

	if err := b.buildHelmConponent(ctx); err != nil {
		return err
	}

	if err := b.buildHoleConponent(ctx); err != nil {
		return err
	}
	tasks, err := task.NewTasks(&b.cfg.Application, nil)
	if err != nil {
		b.Log.Errorf("create tasks error: %s", err.Error())
		return err
	}
	nodes, err := graph.ValidateTortuous(b.cfg.Application.Graph, tasks)
	if err != nil {
		b.Log.Errorf("validate application graph error: %s", err.Error())
		return err
	}

	if len(nodes) != 0 {
		nodesName := make([]string, 0, len(nodes))
		for _, n := range nodes {
			nodesName = append(nodesName, n.Name)
		}
		edges := graph.GetLoopEdge(b.cfg.Application.Graph, nodes)
		buf := bytes.NewBuffer(nil)
		for _, e := range edges {
			buf.WriteString(fmt.Sprintf("\"%s\" -> \"%s\"\n", e.From.Name, e.To.Name))
		}
		err = &trait.Error{
			Err:      fmt.Errorf("包中组件依赖图中存在回环，请联系包负责人对架构进行调整，回环与依赖节点有: %#v,回环图:\n%s", nodesName, buf.String()),
			Internal: trait.ErrAPPlicationComponentTortuous,
		}
		b.Log.Error(err)
		return err
	}

	if err := b.WriteConfigTemplate(); err != nil {
		b.Log.Error(err)
		return err
	}

	if err := b.writeMeta(); err != nil {
		b.Log.Error(err)
		return err
	}

	if err := b.tw.Flush(); err != nil {
		b.Log.Error(err)
		return &trait.Error{
			Err:      err,
			Internal: trait.ECNULL,
			Detail:   "flush data from cache when build application",
		}
	}

	if err := b.tw.Close(); err != nil {
		b.Log.Error(err)
		return &trait.Error{
			Err:      err,
			Internal: trait.ECNULL,
			Detail:   "close file writer when build application",
		}
	}

	if _, err := b.w.Write(buf.Bytes()); err != nil {
		b.Log.Error(err)
		return &trait.Error{
			Err:      err,
			Internal: trait.ECNULL,
			Detail:   "write tgz bytes into file",
		}
	}
	return nil
}

// Build build package
func (b *Builder) Build(ctx context.Context) *trait.Error {
	switch b.cfg.Application.Type {
	case trait.AppBetav1Type:
		return b.buildBetaV1(ctx)
	default:
		b.Log.Errorf("appDefinedType: %s, is not suppport", b.cfg.Application.Type)
		return &trait.Error{
			Internal: trait.ErrAppTypeNoDefined,
		}
	}
}

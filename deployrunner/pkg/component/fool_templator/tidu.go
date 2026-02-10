package templator

import (
	// embed template file
	"bytes"
	_ "embed"
	"os"
	"path"
	"text/template"

	"taskrunner/pkg/component"

	"github.com/ghodss/yaml"
	"helm.sh/helm/v3/pkg/chart"
)

//go:embed tidu_template/deploy_template.yaml
var deployTemplate string

// // go:embed tidu_template/deploy_service_template.yaml
// var serviceTemplate string

// DefaultValues default values
//
//go:embed tidu_template/values.yaml
var DefaultValues []byte

// ValuesConf values inject
//
//go:embed tidu_template/values_conf.yaml
var ValuesConf []byte

func printStringSlice(ss []string) string {
	if len(ss) == 0 {
		return "[]"
	}
	buf := bytes.NewBuffer(nil)
	buf.WriteRune('[')
	writeString := func(s string) {
		buf.WriteRune('"')
		buf.WriteString(s)
		buf.WriteRune('"')
	}
	writeString(ss[0])
	for _, s := range ss[1:] {
		buf.WriteRune(',')
		writeString(s)
	}
	buf.WriteRune(']')
	return buf.String()
}

func tiduDeployTemplate(cname string, deploy *component.FoolDeployment) (string, error) {
	t := template.New("deploy")
	t.Funcs(template.FuncMap{
		"printStringSlice": printStringSlice,
	})
	t = t.Delims(`@@`, `@@`)
	t, err := t.Parse(deployTemplate)
	if err != nil {
		return "", err
	}

	buf := bytes.NewBuffer(nil)
	if err = t.Execute(buf, deploy); err != nil {
		return "", err
	}
	return buf.String(), err
}

// func tiduServiceDeploy(deploy *component.FoolDeployment) (string, error) {
// 	t := template.New("deploy")
// 	t.Funcs(template.FuncMap{
// 		"printStringSlice": printStringSlice,
// 	})
// 	t = t.Delims(`@@`, `@@`)
// 	t, err := t.Parse(serviceTemplate)
// 	if err != nil {
// 		return "", err
// 	}

// 	buf := bytes.NewBuffer(nil)
// 	if err = t.Execute(buf, deploy); err != nil {
// 		return "", err
// 	}
// 	return buf.String(), err
// }

// TiduFoolTemplateOne merge into one chart
func TiduFoolTemplateOne(c *component.FoolComponent, dst string, dvs map[string]interface{}) (v map[string]interface{}, err error) {
	chartDir := dst
	templateDir := path.Join(chartDir, "templates")
	if err = os.MkdirAll(templateDir, 0o666); err != nil {
		return
	}

	for _, c := range c.Deploys {
		bs, err0 := tiduDeployTemplate(string(c.Name), &c)
		if err0 != nil {
			err = err0
			return
		}
		if err = os.WriteFile(path.Join(templateDir, string(c.Name)+"_deploy.yaml"), []byte(bs), 0o666); err != nil {
			return
		}

		// bs, err0 = tiduServiceDeploy(&c)
		// if err0 != nil {
		// 	err = err0
		// 	return
		// }

		// if err = os.WriteFile(path.Join(templateDir, string(c.Name)+"_service.yaml"), []byte(bs), 0666); err != nil {
		// 	return
		// }

		cv, err0 := tiduDeployValues(&c)
		if err0 != nil {
			err = err0
			return
		}
		dvs[string(c.Name)] = cv
	}

	v = dvs
	return
}

// TiduFoolTemplate write the fool compoent chart for tidu into the dst dir
func TiduFoolTemplate(c *component.FoolComponent, dst string) error {
	chartName := c.ComponentNode.Name
	chartDir := path.Join(dst, chartName)
	if err := os.MkdirAll(chartDir, 0o666); err != nil {
		return err
	}
	cm := &chart.Metadata{
		APIVersion:  "v1",
		Version:     c.Version,
		AppVersion:  c.Version,
		Name:        c.Name,
		Description: "auto render chart by taskrunner component tools",
	}
	cmbs, err := yaml.Marshal(cm)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path.Join(chartDir, "Chart.yaml"), cmbs, 0o666); err != nil {
		return err
	}

	templateDir := path.Join(chartDir, "templates")
	if err := os.MkdirAll(templateDir, 0o666); err != nil {
		return err
	}

	dvs := make(map[string]interface{})
	for _, c := range c.Deploys {
		bs, err := tiduDeployTemplate(string(c.Name), &c)
		if err != nil {
			return err
		}
		if err := os.WriteFile(path.Join(templateDir, string(c.Name)+"_deploy.yaml"), []byte(bs), 0o666); err != nil {
			return err
		}
		cv, err := tiduDeployValues(&c)
		if err != nil {
			return err
		}
		dvs[string(c.Name)] = cv
	}

	vs := make(map[string]interface{})
	if err := yaml.Unmarshal(DefaultValues, &vs); err != nil {
		return err
	}
	vs["deploy"] = dvs

	bs, err := yaml.Marshal(vs)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path.Join(chartDir, "values.yaml"), bs, 0o666); err != nil {
		return err
	}

	if err := os.WriteFile(path.Join(templateDir, "values_conf.yaml"), ValuesConf, 0o666); err != nil {
		return err
	}

	return nil
}

func tiduDeployValues(d *component.FoolDeployment) (map[string]interface{}, error) {
	vs := make(map[string]interface{})
	if d.Replica.Custom {
		vs["replicaCount"] = d.Replica.DefaultReplica
	}
	init := make(map[string]interface{})
	for _, c := range d.Init {
		cv := make(map[string]interface{})
		// if c.Resources.Custom {
		// 	cv["resources"] = c.Resources
		// }
		init[string(c.Name)] = cv
	}

	vs["init"] = init

	cs := make(map[string]interface{})
	for _, c := range d.Containers {
		cv := make(map[string]interface{})
		if c.Resources.Custom {
			cv["resources"] = c.Resources
		}
		if c.LivenessProbe != nil {
			cv["livenessProbe"] = c.LivenessProbe
		}
		if c.StartupProbe != nil {
			cv["startupProbe"] = c.StartupProbe
		}
		cv["readinessProbe"] = c.ReadinessProbe
		cs[string(c.Name)] = cv
	}
	vs["container"] = cs
	return vs, nil
}

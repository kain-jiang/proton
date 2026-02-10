//nolint:not use
package component

import (
	"fmt"
	"net/url"
	"regexp"

	"taskrunner/trait"

	"k8s.io/apimachinery/pkg/api/resource"
)

type contraint interface {
	check() error
}

type k8sname string

func (n k8sname) check() error {
	length := len(n)
	if length < 1 || length > 63 {
		return fmt.Errorf("name length must in [1,63]")
	}

	rex, err := regexp.Compile("^[a-z]+([0-9A-Z-]+[a-z]+)*$")
	if err != nil {
		panic(err)
	}
	if !rex.MatchString(string(n)) {
		return fmt.Errorf("name: [%s] must match rfc1035 DNS label standard", n)
	}

	return nil
}

// FoolComponent a simple basic component for foolish developer
type FoolComponent struct {
	trait.ComponentMeta
	FoolComponentSpec
}

// Check the component defined
func (f *FoolComponent) Check() error {
	if err := f.FoolComponentSpec.check(); err != nil {
		return fmt.Errorf("[%s] component define check error: %s", f.Name, err.Error())
	}
	return nil
}

// FoolComponentSpec special
type FoolComponentSpec struct {
	Deploys     []FoolDeployment `json:"deploys,omitempty"`
	Statefulset []FoolDeployment `json:"statefulsets,omitempty"`
}

func (f *FoolComponentSpec) check() error {
	for i, d := range f.Deploys {
		if err := d.check(); err != nil {
			return fmt.Errorf("deploy[%d] error: %s", i, err.Error())
		}
	}

	for i, d := range f.Statefulset {
		if err := d.check(); err != nil {
			return fmt.Errorf("deploy[%d] error: %s", i, err.Error())
		}
	}

	return nil
}

// FoolDeployment deployment
type FoolDeployment struct {
	Name         k8sname             `json:"name"`
	Replica      FoolReplica         `json:"replica"`
	ConfigSchema *trait.ConfigSchema `json:"configSchema,omitempty"`
	Init         []FoolInitContainer `json:"initContainer,omitempty"`
	Containers   []FoolContainer     `json:"containers"`
	Services     []FoolService       `json:"service,omitempty"`
}

func (f *FoolDeployment) check() error {
	if err := f.Name.check(); err != nil {
		return fmt.Errorf("deploy name error:%s", err.Error())
	}

	if err := f.Replica.check(); err != nil {
		return fmt.Errorf("deploy [%s] replica error: %s", f.Name, err.Error())
	}

	cindex := map[k8sname]bool{}
	ccount := 0
	checkContainer := func(i *FoolContainer) error {
		if err := i.check(); err != nil {
			return fmt.Errorf("deploy [%s] container error: %s", f.Name, err.Error())
		}
		if _, ok := cindex[i.Name]; ok {
			return fmt.Errorf("deploy [%s] container [%s] is duplicate", f.Name, i.Name)
		}
		cindex[i.Name] = true
		ccount++
		return nil
	}

	for _, i := range f.Containers {
		if err := checkContainer(&i); err != nil {
			return err
		}
	}
	if ccount == 0 {
		return fmt.Errorf("deploy [%s] must set container, current is empty", f.Name)
	}

	for _, c := range f.Init {
		if err := c.check(); err != nil {
			return err
		}
	}

	sindex := map[k8sname]bool{}
	for _, i := range f.Services {
		if err := i.check(); err != nil {
			return fmt.Errorf("deploy [%s] container error: %s", f.Name, err.Error())
		}
		if _, ok := sindex[i.Name]; ok {
			return fmt.Errorf("deploy [%s] container [%s] is duplicate", f.Name, i.Name)
		}
	}

	return nil
}

// FoolReplica fool pod replica
type FoolReplica struct {
	Custom         bool `json:"custom,omitempty"`
	DefaultReplica int  `json:"defaultReplica"`
}

func (f *FoolReplica) check() error {
	if f.DefaultReplica <= 0 {
		return fmt.Errorf("defaultReplica must > 0")
	}
	return nil
}

// RepoImage repo image
type RepoImage struct {
	Registry string `json:"registry"`
	Image    string `json:"imag"`
	Tag      string `json:"tag"`
}

func (i *RepoImage) check() error {
	if i.Registry == "" {
		return fmt.Errorf("image registry must not empty")
	}

	if i.Image == "" {
		return fmt.Errorf("image must not empty")
	}

	if i.Tag == "" {
		return fmt.Errorf("image tag must not empty")
	}
	return nil
}

// FoolInitContainer init container is a container impl but remove some member
type FoolInitContainer struct {
	Name    k8sname   `json:"name"`
	Command []string  `json:"command"`
	Args    []string  `json:"args,omitempty"`
	Image   RepoImage `json:"image"`
}

func (f *FoolInitContainer) check() error {
	if err := f.Name.check(); err != nil {
		return fmt.Errorf("container name error: %s", err.Error())
	}
	if len(f.Command) == 0 {
		return fmt.Errorf("[%s] container command must not empty", f.Name)
	}

	if err := f.Image.check(); err != nil {
		return fmt.Errorf("[%s] container image error: %s", f.Name, err.Error())
	}

	return nil
}

// FoolContainer container
type FoolContainer struct {
	FoolInitContainer
	LivenessProbe  *FoolProbe     `json:"livenessProbe,omitempty"`
	ReadinessProbe FoolProbe      `json:"readinessProbe"`
	StartupProbe   *FoolProbe     `json:"startupProbe,omitempty"`
	Resources      *FoolResources `json:"resources,omitempty"`
	Ports          []int          `json:"ports"`
}

func (f *FoolContainer) check() error {
	if err := f.Name.check(); err != nil {
		return fmt.Errorf("container name error: %s", err.Error())
	}
	if len(f.Command) == 0 {
		return fmt.Errorf("[%s] container command must not empty", f.Name)
	}

	probes := []*FoolProbe{f.LivenessProbe, f.StartupProbe}
	for _, p := range probes {
		if p == nil {
			continue
		}
		if err := p.check(); err != nil {
			return fmt.Errorf("[%s] container probe error: %s", f.Name, err.Error())
		}
		if p.TerminationGracePeriodSeconds < 0 {
			return fmt.Errorf("[%s] container probe TerminationGracePeriodSeconds must is a interger >= 0, currnent is %d", f.Name, p.TerminationGracePeriodSeconds)
		}
	}
	if err := f.ReadinessProbe.check(); err != nil {
		return fmt.Errorf("[%s] container probe error: %s", f.Name, err.Error())
	}

	if err := f.Image.check(); err != nil {
		return fmt.Errorf("[%s] container image error: %s", f.Name, err.Error())
	}

	if f.Resources != nil {
		if err := f.Resources.check(); err != nil {
			return fmt.Errorf("[%s] container resource error: %s", f.Name, err.Error())
		}
	}
	count := 0
	for _, p := range f.Ports {
		if p <= 0 || p > 65536 {
			return fmt.Errorf("[%s] container port must in [1, 65536]", f.Name)
		}
		count++
	}
	if count == 0 {
		return fmt.Errorf("[%s] container port must set one or more", f.Name)
	}
	return nil
}

// FoolService deploy or component service
type FoolService struct {
	Name k8sname    `json:"name"`
	Port []FoolPort `json:"ports"`
}

func (f *FoolService) check() error {
	if err := f.Name.check(); err != nil {
		return err
	}
	pcount := 0
	for _, p := range f.Port {
		if err := p.check(); err != nil {
			return fmt.Errorf("[%s] port error: %s", f.Name, err.Error())
		}
		pcount++
	}

	if pcount == 0 {
		return fmt.Errorf("[%s] port must set, it's empty now", f.Name)
	}
	return nil
}

// FoolPort fool service port
type FoolPort struct {
	Name       k8sname `json:"name"`
	Port       int     `json:"port"`
	TargetPort int     `json:"targetPort,omitepty"`
	Protocol   string  `json:"protocol"`
}

func (f *FoolPort) check() error {
	if err := f.Name.check(); err != nil {
		return err
	}

	if f.Port < 1 || f.Port > 65536 {
		return fmt.Errorf("[%s] service port must in [1, 65536]", f.Name)
	}
	if f.TargetPort < 0 {
		f.TargetPort = 0
	}
	if f.Protocol != "TCP" && f.Protocol != "UDP" && f.Protocol != "SCTP" {
		return fmt.Errorf(`[%s] service protocol must is TCP, UDP or SCTP`, f.Name)
	}
	return nil
}

// FoolResources resources
type FoolResources struct {
	Custom   bool                `json:"custom,omitepty"`
	Requests *FoolResourceObject `json:"requests,omitepty"`
	Limit    *FoolResourceObject `json:"limit,omitepty"`
}

func (f *FoolResources) check() error {
	// if f.Requests != nil {
	// 	if err := f.Requests.check(); err != nil {
	// 		return err
	// 	}
	// }

	// if f.Limit != nil {
	// 	if err := f.Limit.check(); err != nil {
	// 		return err
	// 	}
	// }

	if f.Limit != nil && f.Requests != nil {
		if f.Limit.less(f.Requests) {
			return fmt.Errorf("resoures limit must bigger then requests, limit: %#v, requests: %#v", f.Limit, f.Requests)
		}
	}

	return nil
}

// FoolResourceObject resource object
type FoolResourceObject struct {
	Memory *resource.Quantity `json:"memory"`
	// m      resource.Quantity `json:"-"`
	CPU *resource.Quantity `json:"cpu"`
	// c      resource.Quantity `json:"-"`
}

// func (f *foolResourceObject) check() error {

// 	m, err := resource.ParseQuantity(f.Memory)
// 	if err != nil {
// 		return fmt.Errorf("resource memory obj parse error: [%s]", err.Error())
// 	}
// 	f.m = m
// 	c, err := resource.ParseQuantity(f.Memory)
// 	if err != nil {
// 		return fmt.Errorf("resource cpu obj parse error: [%s]", err.Error())
// 	}
// 	f.c = c

// 	return nil
// }

// func (f *foolResourceObject) check() error {

// 	m, err := resource.ParseQuantity(f.Memory)
// 	if err != nil {
// 		return fmt.Errorf("resource memory obj parse error: [%s]", err.Error())
// 	}
// 	f.m = m
// 	c, err := resource.ParseQuantity(f.Memory)
// 	if err != nil {
// 		return fmt.Errorf("resource cpu obj parse error: [%s]", err.Error())
// 	}
// 	f.c = c

// 	return nil
// }

func (f *FoolResourceObject) less(o *FoolResourceObject) bool {
	less := false
	if f.Memory != nil && o.Memory != nil {
		less = less || f.Memory.Cmp(*o.Memory) == -1
	}

	if f.CPU != nil && o.CPU != nil {
		less = less || f.CPU.Cmp(*o.CPU) == -1
	}
	return less
	// return f.c.Cmp(o.c) == -1 && f.m.Cmp(o.m) == -1
}

// FoolProbe probe
type FoolProbe struct {
	HTTPGet                       *FoolHTTPGetAction   `json:"httpGet,omitempty"`
	Exec                          *FoolExecAction      `json:"exec,omitempty"`
	TCPSocket                     *FoolTCPSocketAction `json:"tcpSocket,omitempty"`
	FailureThreshold              int                  `json:"failureThreshold"`
	InitialDelaySeconds           int                  `json:"initialDelaySeconds"`
	PeriodSeconds                 int                  `json:"periodSeconds"`
	SuccessThreshold              int                  `json:"successThreshold"`
	TerminationGracePeriodSeconds int                  `json:"terminationGracePeriodSeconds,omitempty"`
	TimeoutSeconds                int                  `json:"timeoutSeconds"`
}

func (f *FoolProbe) check() error {
	if f == nil {
		return nil
	}
	probeCount := 0
	var probe contraint
	if f.HTTPGet != nil {
		probe = f.HTTPGet
		probeCount++
	}
	if f.Exec != nil {
		probe = f.Exec
		probeCount++
	}
	if f.TCPSocket != nil {
		probe = f.TCPSocket
		probeCount++
	}

	if probeCount != 1 {
		return fmt.Errorf("probe must set one way, currnet has %d probe", probeCount)
	}

	if err := probe.check(); err != nil {
		return err
	}

	if f.FailureThreshold < 1 {
		return fmt.Errorf("FailureThreshold must is a interger > 0, currnet is %d", f.FailureThreshold)
	}

	if f.PeriodSeconds < 1 {
		return fmt.Errorf("periodSeconds must is a interger > 0, current is %d", f.PeriodSeconds)
	}

	if f.SuccessThreshold < 1 {
		return fmt.Errorf("SuccessThreshold must is a interger >= 1, currnent is %d", f.SuccessThreshold)
	}

	if f.TimeoutSeconds < 1 {
		return fmt.Errorf("TimeoutSeconds must is a interger >= 1, currnent is %d", f.TimeoutSeconds)
	}

	return nil
}

// FoolExecAction command action in probe
type FoolExecAction struct {
	Command string `json:"command"`
}

func (f *FoolExecAction) check() error {
	if f.Command == "" {
		return fmt.Errorf("exec probe command must is a not empty string")
	}
	return nil
}

// FoolHTTPGetAction http get action in probe
type FoolHTTPGetAction struct {
	Path   string `json:"path"`
	Port   int    `json:"port"`
	Scheme string `json:"scheme"`
}

func (f *FoolHTTPGetAction) check() error {
	if f.Scheme != "HTTP" && f.Scheme != "HTTPS" {
		return fmt.Errorf("http probe: schame must is 'HTTPS' or 'HTTP'")
	}
	if f.Port < 1 || f.Port > 65536 {
		return fmt.Errorf("http probe: port Number must be in the range 1 to 65535, current is %d", f.Port)
	}

	testpath := "http://127.0.0.1" + f.Path
	_, err := url.Parse(testpath)
	if err != nil {
		return fmt.Errorf("http probe: path must is a uri format string, path: [%s], testpath: [%s], error:%s", f.Path, testpath, err.Error())
	}
	return nil
}

// FoolTCPSocketAction tcp action in probe
type FoolTCPSocketAction struct {
	Port int `json:"port"`
}

func (f *FoolTCPSocketAction) check() error {
	if f.Port < 1 || f.Port > 65536 {
		return fmt.Errorf("tcp probe: port Number must be in the range 1 to 65535, current is %d", f.Port)
	}
	return nil
}

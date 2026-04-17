package testing

import (
	"context"
	"reflect"
	stdtime "time"
	"unsafe"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
	helmtime "helm.sh/helm/v3/pkg/time"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/helm3"
)

type FakeHelm3 struct {
	storage.Storage

	Namespace string

	CallUpgrade          bool
	CallReconcileRelease bool

	ErrUninstall        error
	ErrGetRelease       error
	ErrUpgrade          error
	ErrReconcileRelease error
	UpgradeCalls        []UpgradeCall
	UninstallCalls      []string
	UninstallCallDetails []UninstallCall
}

type UpgradeCall struct {
	Release string
	Wait    bool
	Timeout stdtime.Duration
	Values  map[string]interface{}
}

type UninstallCall struct {
	Release string
	Timeout stdtime.Duration
}

// ReconcileRelease implements helm3.Client.
func (c *FakeHelm3) ReconcileRelease(ctx context.Context, release string, chart string, values map[string]any) error {
	c.CallReconcileRelease = true
	return c.ErrReconcileRelease
}

func New(namespace string, log func(string, ...interface{}), releases ...*release.Release) *FakeHelm3 {
	var h = FakeHelm3{
		Storage: storage.Storage{
			Driver: driver.NewMemory(),
			Log:    log,
		},
		Namespace: namespace,
	}
	for _, r := range releases {
		if r.Info == nil {
			r.Info = &release.Info{Status: release.StatusDeployed}
		}
		if err := h.Storage.Create(r); err != nil {
			panic(err)
		}
	}
	return &h
}

func (f *FakeHelm3) Install(name string, chartRef *helm3.ChartRef, opts ...helm3.InstallOption) error {
	now := helmtime.Now()
	return f.Storage.Create(&release.Release{
		Name: name,
		Info: &release.Info{
			FirstDeployed: now,
			LastDeployed:  now,
			Status:        release.StatusDeployed,
		},
		Chart: &chart.Chart{
			Metadata: &chart.Metadata{
				Name: name,
			},
		},
		Version:   1,
		Namespace: f.Namespace,
	})
}

func (f *FakeHelm3) Upgrade(release string, chartRef *helm3.ChartRef, opts ...helm3.UpgradeOption) error {
	f.CallUpgrade = true
	call := UpgradeCall{Release: release}
	for _, opt := range opts {
		optValue := reflect.ValueOf(opt)
		if optValue.Kind() != reflect.Func || optValue.Type().NumIn() != 1 {
			continue
		}
		paramType := optValue.Type().In(0)
		if paramType.Kind() != reflect.Pointer || paramType.Elem().Kind() != reflect.Struct {
			continue
		}

		params := reflect.New(paramType.Elem())
		optValue.Call([]reflect.Value{params})

		paramValue := params.Elem()
		waitField := paramValue.FieldByName("wait")
		if waitField.IsValid() && waitField.Kind() == reflect.Bool {
			call.Wait = waitField.Bool()
		}
		timeoutField := paramValue.FieldByName("timeout")
		if timeoutField.IsValid() && timeoutField.Type() == reflect.TypeOf(stdtime.Duration(0)) {
			call.Timeout = stdtime.Duration(timeoutField.Int())
		}
		valuesField := paramValue.FieldByName("values")
		if valuesField.IsValid() && valuesField.Type() == reflect.TypeOf(map[string]interface{}{}) && !valuesField.IsNil() {
			call.Values = reflect.NewAt(valuesField.Type(), unsafe.Pointer(valuesField.UnsafeAddr())).Elem().Interface().(map[string]interface{})
		}
	}
	f.UpgradeCalls = append(f.UpgradeCalls, call)
	return f.ErrUpgrade
}

func (f *FakeHelm3) Uninstall(release string, opts ...helm3.UninstallOption) error {
	f.Log("FakeHelm3.Uninstall() release=%v", release)
	call := UninstallCall{Release: release}
	for _, opt := range opts {
		optValue := reflect.ValueOf(opt)
		if optValue.Kind() != reflect.Func || optValue.Type().NumIn() != 1 {
			continue
		}
		paramType := optValue.Type().In(0)
		if paramType.Kind() != reflect.Pointer || paramType.Elem().Kind() != reflect.Struct {
			continue
		}

		params := reflect.New(paramType.Elem())
		optValue.Call([]reflect.Value{params})

		paramValue := params.Elem()
		timeoutField := paramValue.FieldByName("timeout")
		if timeoutField.IsValid() && timeoutField.Type() == reflect.TypeOf(stdtime.Duration(0)) {
			call.Timeout = stdtime.Duration(timeoutField.Int())
		}
	}
	f.UninstallCalls = append(f.UninstallCalls, release)
	f.UninstallCallDetails = append(f.UninstallCallDetails, call)
	return f.ErrUninstall
}

func (f *FakeHelm3) NameSpace(namespace string) helm3.Client {
	f.Namespace = namespace
	return f
}

func (f *FakeHelm3) GetRelease(name string) (*release.Release, error) {
	if f.ErrGetRelease != nil {
		return nil, f.ErrGetRelease
	}
	return f.Storage.Last(name)
}

func (f *FakeHelm3) PullChart(name, version string, reg *helm3.OCIRegistryConfig) (string, func(), error) {
	return "", nil, nil
}

func (f *FakeHelm3) PushChart(file string, reg *helm3.OCIRegistryConfig) error {
	return nil
}

func (f *FakeHelm3) WithRelease(n, v string) {
	if err := f.Storage.Create(&release.Release{
		Name: n,
		Info: &release.Info{
			Status: release.StatusDeployed,
		},
		Chart: &chart.Chart{
			Metadata: &chart.Metadata{
				Name:    n,
				Version: v,
			},
		},
	}); err != nil {
		panic(err)
	}
}

var _ helm3.Client = (*FakeHelm3)(nil)

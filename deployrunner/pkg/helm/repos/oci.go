package helm

import (
	"context"
	"fmt"

	"taskrunner/pkg/component"
	"taskrunner/pkg/helm"
	"taskrunner/trait"

	"helm.sh/helm/v3/pkg/registry"
)

type OCi struct {
	Registry  string           `json:"registry,omitempty"`
	PlainHTTP bool             `json:"plain_http,omitempty"`
	Username  string           `json:"username,omitempty"`
	Password  string           `json:"password,omitempty"`
	RepoName  string           `json:"name,omitempty"`
	cli       *registry.Client `json:"-"`
}

func (r *OCi) Init(ctx context.Context) (helm.Repo, *trait.Error) {
	return r, r.loginRegistry()
}

func (r *OCi) loginRegistry() *trait.Error {
	if r.Username != "" {
		// login
		regClient, err := registry.NewClient()
		if err != nil {
			return &trait.Error{
				Err:      err,
				Detail:   fmt.Sprintf("login oci helm chart repo with %#v", r),
				Internal: trait.ECNULL,
			}
		}
		err = regClient.Login(
			r.Registry,
			registry.LoginOptInsecure(false),
			registry.LoginOptBasicAuth(r.Username, r.Password),
		)
		if err != nil {
			return &trait.Error{
				Err:      err,
				Detail:   fmt.Sprintf("login oci helm chart repo with %#v", r),
				Internal: trait.ECNULL,
			}
		}
	}
	opts := []registry.ClientOption{}
	if r.PlainHTTP {
		opts = append(opts, registry.ClientOptPlainHTTP())
	}
	cli, rerr := registry.NewClient(opts...)
	if rerr != nil {
		return &trait.Error{
			Err:      rerr,
			Detail:   fmt.Sprintf("login oci helm chart repo with %#v", r),
			Internal: trait.ECNULL,
		}
	}
	r.cli = cli

	return nil
}

func (r *OCi) Store(ctx context.Context, chart *component.HelmComponent, data []byte) *trait.Error {
	ref := fmt.Sprintf("%s/%s:%s", r.Registry, chart.Name, chart.Version)
	if _, rerr := r.cli.Push(data, ref); rerr != nil {
		return &trait.Error{
			Err:      rerr,
			Detail:   fmt.Sprintf("push chart  %s:%s into oci helm chart", chart.Name, chart.Version),
			Internal: trait.ECNULL,
		}
	}
	return nil
}

func (r *OCi) Fetch(ctx context.Context, chart *component.HelmComponent) ([]byte, *trait.Error) {
	ref := fmt.Sprintf("%s/%s:%s", r.Registry, chart.Name, chart.Version)
	res, rerr := r.cli.Pull(ref)
	if rerr != nil {
		return nil, &trait.Error{
			Err:      rerr,
			Detail:   fmt.Sprintf("push chart  %s:%s into oci helm chart", chart.Name, chart.Version),
			Internal: trait.ECNULL,
		}
	}
	return res.Chart.Data, nil
}

func (r *OCi) Name() string {
	return r.RepoName
}

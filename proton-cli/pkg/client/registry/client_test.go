package registry

import (
	"context"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/containers/image/v5/types"
	"github.com/sirupsen/logrus"
)

func TestNew(t *testing.T) {
	type args struct {
		config *Config
	}
	tests := []struct {
		name    string
		args    args
		want    *Client
		wantErr bool
	}{
		{
			name: "host only",
			args: args{config: &Config{Address: "registry.example.org"}},
			want: &Client{address: "registry.example.org", sys: &types.SystemContext{DockerInsecureSkipTLSVerify: types.OptionalBoolTrue}},
		},
		{
			name: "host and port",
			args: args{config: &Config{Address: "registry.example.org:5000"}},
			want: &Client{address: "registry.example.org:5000", sys: &types.SystemContext{DockerInsecureSkipTLSVerify: types.OptionalBoolTrue}},
		},
		{
			name:    "invalid address",
			args:    args{config: &Config{Address: "a:b:c"}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_ListRepositoryTags(t *testing.T) {
	if n := strings.ToUpper(t.Name()); os.Getenv(n) == "" {
		t.Skipf("%v is not set", n)
	}

	var logger = logrus.Logger{
		Out:       os.Stderr,
		Hooks:     make(logrus.LevelHooks),
		Formatter: new(logrus.TextFormatter),
		Level:     logrus.DebugLevel,
	}
	type fields struct {
		address string
		sys     *types.SystemContext
	}
	type args struct {
		ctx        context.Context
		repository string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "library/golang",
			fields: fields{
				address: "acr.aishu.cn",
				sys:     &types.SystemContext{},
			},
			args: args{ctx: context.TODO(), repository: "library/golang"},
			want: []string{
				"1",
				"1-alpine",
				"1-alpine3.15",
				"1.16.15",
				"1.17-alpine",
				"1.17-alpine3.15",
				"1.17.11",
				"1.17.6-alpine",
				"1.17.6-alpine3.15",
				"1.19",
				"1.19.1",
				"1.19.2",
				"1.20",
				"1.20.0",
				"1.20.4",
				"alpine",
				"alpine3.15",
				"latest",
			},
			wantErr: false,
		},
		{
			name:    "repository not found",
			fields:  fields{address: "acr.aishu.cn", sys: &types.SystemContext{}},
			args:    args{ctx: context.TODO(), repository: "library/nothing"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				address: tt.fields.address,
				sys:     tt.fields.sys,
				log:     &logger,
			}
			got, err := c.ListRepositoryTags(tt.args.ctx, tt.args.repository)
			if err != nil {
				t := reflect.TypeOf(err)
				logger.Errorf("Client.ListRepositoryTags() error = %#v, type = %v", err, t)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.ListRepositoryTags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.ListRepositoryTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

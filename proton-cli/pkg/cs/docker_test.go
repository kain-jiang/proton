package cs

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

func readFile(t *testing.T, name string) []byte {
	c, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func Test_addDockerConfigInsecureHost(t *testing.T) {
	type args struct {
		crhosts []string
		port    string
		b       []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "NotExist",
			args: args{
				crhosts: []string{"node-45-71", "node-45-72"},
				port:    "5000",
				b:       readFile(t, "testdata/docker-daemon-not-exist-crhost.json"),
			},
			want:    readFile(t, "testdata/docker-daemon-exist-crhost.json"),
			wantErr: false,
		},
		{
			name: "Exist",
			args: args{
				crhosts: []string{"node-45-71", "node-45-72"},
				port:    "5000",
				b:       readFile(t, "testdata/docker-daemon-exist-crhost.json"),
			},
			want:    readFile(t, "testdata/docker-daemon-exist-crhost.json"),
			wantErr: false,
		},
		{
			name: "HalfExist",
			args: args{
				crhosts: []string{"node-45-71"},
				port:    "5000",
				b:       readFile(t, "testdata/docker-daemon-exist-crhost.json"),
			},
			want:    readFile(t, "testdata/docker-daemon-exist-crhost.json"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := addDockerConfigInsecureHost(tt.args.crhosts, tt.args.port, tt.args.b)
			if (err != nil) != tt.wantErr {
				t.Errorf("addDockerConfigInsecureHost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			conf := make(map[string]interface{})
			if err := json.Unmarshal(got, &conf); err != nil {
				t.Errorf("failed to Unmarshal docker daemon.json: %v", err)
			}
			if !reflect.DeepEqual(conf["insecure-registries"], []interface{}{"node-45-73:5000", "registry.aishu.cn:15000", "node-45-71:5000", "node-45-72:5000"}) {
				t.Errorf("addDockerConfigInsecureHost() = %s, want %s", conf["insecure-registries"], []string{"node-45-73:5000", "registry.aishu.cn:15000", "node-45-71:5000", "node-45-72:5000"})
			}
		})
	}
}

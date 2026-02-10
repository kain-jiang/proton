package utils_test

import (
	"path/filepath"
	"slices"
	"testing"

	"taskrunner/pkg/utils"
)

func TestOrasListTags(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		ociPath string
		want    []string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:    "test list oci tar tags",
			ociPath: filepath.Join("testdata", "oci-file.tar"),
			want:    []string{"registry.aishu.cn:15000/public/pause:3.6", "docker.io/library/test:latest"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := utils.OrasListTags(tt.ociPath)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("OrasListTags() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("OrasListTags() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			slices.Sort(got)
			slices.Sort(tt.want)
			// Simple comparison
			if !slices.Equal(got, tt.want) {
				t.Errorf("OrasListTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOrasPushImage(t *testing.T) {
	t.SkipNow() // 依赖CR仓库
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		ociPath  string
		srcImage string
		dest     string
		username string
		password string
		wantErr  bool
	}{
		// TODO: Add test cases.
		{
			name:     "push image to registry",
			ociPath:  filepath.Join("testdata", "oci-file.tar"),
			srcImage: "registry.aishu.cn:15000/public/pause:3.6",
			dest:     "localhost:5000/public/pause:3.6",
			username: "",
			password: "",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := utils.OrasPushImage(tt.ociPath, tt.srcImage, tt.dest, tt.username, tt.password)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("OrasPushImage() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("OrasPushImage() succeeded unexpectedly")
			}
		})
	}
}

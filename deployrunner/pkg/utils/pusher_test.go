package utils_test

import (
	"path/filepath"
	"testing"

	"taskrunner/pkg/utils"
)

func TestPuserPushChart(t *testing.T) {
	t.SkipNow() // 依赖Chart/OCI仓库
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		chartPath string
		repo      string
		username  string
		password  string
		wantErr   bool
	}{
		// TODO: Add test cases.
		{
			name:      "push proton-cr to haproxy http repo",
			chartPath: filepath.Join("testdata", "demo-0.1.0.tgz"),
			repo:      "http://chartmuseum.aishu.cn:15001",
			username:  "",
			password:  "",
			wantErr:   false,
		},
		{
			name:      "push proton-cr to http repo",
			chartPath: filepath.Join("testdata", "demo-0.1.0.tgz"),
			repo:      "http://localhost:5001",
			username:  "",
			password:  "",
			wantErr:   false,
		},
		{
			name:      "push proton-cr to oci repo",
			chartPath: filepath.Join("testdata", "demo-0.1.0.tgz"),
			repo:      "oci://localhost:5000",
			username:  "",
			password:  "",
			wantErr:   false,
		},
		{
			name:      "push aishu acr to oci repo with invalid auth",
			chartPath: filepath.Join("testdata", "demo-0.1.0.tgz"),
			repo:      "oci://acr.aishu.cn/ict",
			username:  "invalid",
			password:  "invalid",
			wantErr:   true,
		},
		{
			name:      "push aishu acr to chartmuseum repo with invalid auth",
			chartPath: filepath.Join("testdata", "demo-0.1.0.tgz"),
			repo:      "https://acr.aishu.cn/chartrepo/ict",
			username:  "invalid",
			password:  "invalid",
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := utils.PuserPushChart(tt.chartPath, tt.repo, tt.username, tt.password)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("PuserPushChart() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("PuserPushChart() succeeded unexpectedly")
			}
		})
	}
}

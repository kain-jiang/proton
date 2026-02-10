package cmd

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"reflect"
	"testing"

	"os"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/core/push"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/cr"
)

func TestPushImagesApp(t *testing.T) {
	fout, err := os.CreateTemp(os.TempDir(), "testdeployApp")
	assert.Nil(t, err, "create tmp fail")
	fout.Close()
	defer os.Remove(fout.Name())
	patches := gomonkey.NewPatches()
	tests := []struct {
		name     string
		wantErr  bool
		preHook  func()
		postHook func()
	}{
		{
			name:    "FileNotFound",
			wantErr: true,
			preHook: func() {
				patches.ApplyFunc(filepath.Abs, func(string) (string, error) {
					return "", errors.New("abs error mock")
				})
			},
			postHook: func() {
				patches.Reset()
			},
		},
		{
			name:    "push error",
			wantErr: true,
			preHook: func() {
				packagePath = fout.Name()
			},
		},
		{
			name: "success",
			preHook: func() {
				patches.ApplyFunc(PushImagesAndAppDir, func(ctx context.Context, lg *logrus.Logger, opts push.ImagePushOpts, namespace string, rootDir string) error {
					return nil
				})
			},
			postHook: func() {
				patches.Reset()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.preHook != nil {
				tt.preHook()
			}
			if err := pushImagesApp(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("testcase: %s, CheckFile() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
			if tt.postHook != nil {
				tt.postHook()
			}
		})
	}
}

func TestPushImagesAndAppDir(t *testing.T) {
	t.Skip("need to fix test failed")
	opts := push.ImagePushOpts{}
	lg := logrus.New()
	lg.SetOutput(io.Discard)
	namespace := ""
	rootDir, err := os.MkdirTemp(os.TempDir(), "testroot")
	assert.Nil(t, err, "mkdir temp dir error")
	defer os.Remove(rootDir)
	testDir := rootDir

	// tmp test file or dir
	tmpDir, err := os.MkdirTemp(rootDir, "testdir")
	assert.Nil(t, err, "mkdir temp dir error")
	defer os.Remove(tmpDir)
	tmpNilDir, err := os.MkdirTemp(rootDir, ".testdir")
	assert.Nil(t, err, "mkdir temp dir error")
	defer os.Remove(tmpNilDir)
	tmpFile, err := os.CreateTemp(rootDir, "testtempfile")
	assert.Nil(t, err, "mkdir temp dir error")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	tmpFile0, err := os.CreateTemp(rootDir, ".testtempfile")
	assert.Nil(t, err, "mkdir temp dir error")
	tmpFile0.Close()
	defer os.Remove(tmpFile0.Name())

	tmpFile1, err := os.CreateTemp(rootDir, "testtempfile")
	assert.Nil(t, err, "mkdir temp dir error")
	tmpFile1.Close()
	defer os.Remove(tmpFile1.Name())
	//  tmp test file or dir

	patchesMap := make(map[string]*gomonkey.Patches)
	tests := []struct {
		name     string
		wantErr  bool
		preHook  func()
		postHook func()
	}{
		{
			name:    "NewCrError",
			wantErr: true,
			preHook: func() {
				patchesMap["NewCrError"] = gomonkey.NewPatches().ApplyFunc(push.NewCr, func(push.ImagePushOpts) (*cr.Cr, error) {
					return nil, errors.New("push.NewCr error mock")
				})
			},
			postHook: func() {
				patchesMap["NewCrError"].Reset()
				opts.Registry = "test"
			},
		},
		{
			name:    "GetUploadCacheFail",
			wantErr: true,
			preHook: func() {
				rootDir = tmpFile1.Name()
			},
			postHook: func() {
				rootDir = testDir
			},
		},
		{
			name:    "NewDeployInstallerClientError",
			wantErr: true,
			preHook: func() {
				patchesMap["NewDeployInstallerClientError"] = gomonkey.NewPatches().ApplyFunc(push.NewDeployInstallerClient, func(context.Context, string) (*push.DeployInstaller, error) {
					return nil, errors.New("new deploy installer error mock")
				})
			},
			postHook: func() {
				patchesMap["NewDeployInstallerClientError"].Reset()
			},
		},
		{
			name:    "DeployInstallerUploadError",
			wantErr: true,
			preHook: func() {
				patchesMap["DeployInstallerUploadError"] = gomonkey.NewPatches().ApplyFunc(push.NewDeployInstallerClient, func(context.Context, string) (*push.DeployInstaller, error) {
					return &push.DeployInstaller{}, nil
				}).ApplyMethod(reflect.TypeOf(&push.DeployInstaller{}), "Upload", func(*push.DeployInstaller, context.Context, *logrus.Logger, string) (string, error) {
					return "", errors.New("deploy installer upload ")
				})
			},
			postHook: func() {
				patchesMap["DeployInstallerUploadError"].Reset()
			},
		},
		{
			name:    "DeployInstallerUploadConflict",
			wantErr: false,
			preHook: func() {
				patchesMap["DeployInstallerUploadConflict"] = gomonkey.NewPatches().ApplyFunc(push.NewDeployInstallerClient, func(context.Context, string) (*push.DeployInstaller, error) {
					return &push.DeployInstaller{}, nil
				}).ApplyMethod(reflect.TypeOf(&push.DeployInstaller{}), "Upload", func(*push.DeployInstaller, context.Context, *logrus.Logger, string) (string, error) {
					return "", push.ErrAPPConflict
				})
			},
			postHook: func() {
				patchesMap["DeployInstallerUploadConflict"].Reset()
			},
		},
		{
			name:    "DeployInstallerUploadSucess",
			wantErr: false,
			preHook: func() {
				patchesMap["DeployInstallerUploadSucess"] = gomonkey.NewPatches().ApplyFunc(push.NewDeployInstallerClient, func(context.Context, string) (*push.DeployInstaller, error) {
					return &push.DeployInstaller{}, nil
				}).ApplyMethod(reflect.TypeOf(&push.DeployInstaller{}), "Upload", func(*push.DeployInstaller, context.Context, *logrus.Logger, string) (string, error) {
					return "test", nil
				})
			},
			postHook: func() {
				patchesMap["DeployInstallerUploadSucess"].Reset()
			},
		},
		{
			name:    "DeployInstallerNotInstall",
			wantErr: true,
			preHook: func() {
				patchesMap["DeployInstallerNotInstall"] = gomonkey.NewPatches().ApplyFunc(push.NewDeployInstallerClient, func(ctx context.Context, ns string, check bool) (*push.DeployInstaller, error) {
					if !check {
						return nil, push.ErrDeployInstallerNotInstalled
					} else {
						patchesMap["DeployInstallerNotInstall"].Reset()
						return push.NewDeployInstallerClient(ctx, ns, check)
					}
				})
			},
			postHook: func() {
				patchesMap["DeployInstallerNotInstall"].Reset()
			},
		},
		{
			name:    "DeployInstallerCheckFileMock",
			wantErr: false,
			preHook: func() {
				patchesMap["DeployInstallerCheckFileMock"] = gomonkey.NewPatches().ApplyFunc(push.NewDeployInstallerClient, func(context.Context, string) (*push.DeployInstaller, error) {
					return nil, push.ErrDeployInstallerNotInstalled
				}).ApplyMethod(reflect.TypeOf(&push.DeployInstaller{}), "CheckFile", func(*push.DeployInstaller, string) error {
					return nil
				}).ApplyFunc(getUploadCache, func(fpath string) (*os.File, map[string]bool, error) {
					res := map[string]bool{
						tmpFile1.Name(): true,
					}
					file, err := os.OpenFile(fpath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
					if err != nil {
						return nil, nil, err
					}
					return file, res, nil
				})
			},
			postHook: func() {
				patchesMap["DeployInstallerCheckFileMock"].Reset()
			},
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := os.Remove(filepath.Join(rootDir, uploadRecordFileName))
			if !os.IsNotExist(err) {
				assert.Nil(t, err, "clean work space error")
			}
			if tt.preHook != nil {
				tt.preHook()
			}
			err = PushImagesAndAppDir(ctx, lg, opts, namespace, rootDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("testcase: %s, CheckFile() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
			if tt.postHook != nil {
				tt.postHook()
			}
		})
	}
}

func TestPrintReport(t *testing.T) {
	reports := []uploadResult{
		{
			Type: unknowType,
		},
		{
			Status: uploadIgnoreStatus,
			Type:   deployAppType,
		},
		{
			Status: uploadSucessStatus,
			Type:   deployAppType,
		},
		{
			Status: uploadFailedStatus,
			Type:   deployAppType,
		},
		{
			Status: uploadIgnoreStatus,
			Type:   ociImageType,
		},
		{
			Status: uploadSucessStatus,
			Type:   ociImageType,
		},
		{
			Status: uploadFailedStatus,
			Type:   ociImageType,
		},
	}
	lg := logrus.New()
	lg.SetOutput(io.Discard)
	assert.Equal(t, 3, printReport(lg, reports))

}

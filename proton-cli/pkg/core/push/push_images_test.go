package push

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey"
	"github.com/mholt/archiver/v3"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/configuration"
	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/cr"
)

func TestPushImages(t *testing.T) {
	type args struct {
		opts ImagePushOpts
	}
	fi, _ := os.CreateTemp(os.TempDir(), "testfile")
	defer fi.Close()
	defer os.Remove(filepath.Join(os.TempDir(), fi.Name()))
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "HelmRepoSpecified",
			args: args{opts: ImagePushOpts{
				Registry:       "test.registry",
				OCIPackagePath: "/path/to/oci/pkg",
			}},
		},
		{
			name: "K8sClientSetNil",
			args: args{opts: ImagePushOpts{
				OCIPackagePath: "/path/to/oci/pkg",
			}},
			wantErr: true,
		},
	}
	crPushImagesPatcher := gomonkey.ApplyMethod(reflect.TypeOf(&cr.Cr{}), "PushImages", func(_ *cr.Cr, _ string) error {
		log.Info("Patch method cr.PushCharts")
		return nil
	}).ApplyFunc(client.NewK8sClient, func() (clientDynamic dynamic.Interface, clientSet *kubernetes.Clientset) {
		return nil, nil
	}).ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
		return fi.Stat()
	}).ApplyFunc(archiver.Unarchive, func(src, target string) error {
		return nil
	})
	defer crPushImagesPatcher.Reset()
	// newKClientPatcher := gomonkey.ApplyFunc(client.NewK8sClient, func() (clientDynamic dynamic.Interface, clientSet *kubernetes.Clientset) {
	// 	return nil, nil
	// })
	// defer newKClientPatcher.Reset()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if err := PushImages(tt.args.opts, ""); (err != nil) != tt.wantErr {
				t.Errorf("testcase: %s, PushImages() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}

	// test error code
	crPushImagesPatcher.Reset()
	testCases := []struct {
		name     string
		args     args
		wantErr  bool
		preHook  func()
		postHook func()
	}{
		{
			name: "FileNotfound",
			args: args{opts: ImagePushOpts{
				Registry:       "test.registry",
				OCIPackagePath: "/unknowfile/qweadoq/",
			}},
			wantErr: true,
		},
		{
			name: "MktempError",
			args: args{opts: ImagePushOpts{
				Registry:       "test.registry",
				OCIPackagePath: "/unknowfile/qweadoq/",
			}},
			wantErr: true,
			preHook: func() {
				crPushImagesPatcher.Reset()
				crPushImagesPatcher.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
					return fi.Stat()
				})
				crPushImagesPatcher.ApplyFunc(os.MkdirTemp, func(string, string) (string, error) {
					return "", errors.New("MKdirTemp mock error")
				})
			},
		},
		{
			name: "UnarchiveError",
			args: args{opts: ImagePushOpts{
				OCIPackagePath: "/path/to/oci/pkg",
			}},
			wantErr: true,
			preHook: func() {
				crPushImagesPatcher.Reset()
				crPushImagesPatcher.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
					return fi.Stat()
				})
			},
		},
		{
			name: "PushImagesFileError",
			args: args{opts: ImagePushOpts{
				OCIPackagePath: "/path/to/oci/pkg",
			}},
			wantErr: true,
			preHook: func() {
				crPushImagesPatcher.ApplyMethod(reflect.TypeOf(&cr.Cr{}), "PushImages", func(_ *cr.Cr, _ string) error {
					return cr.ErrSkopeoImagesFile
				}).ApplyFunc(archiver.Unarchive, func(src, target string) error {
					return nil
				})
			},
		},
		{
			name: "PushImagesError",
			args: args{opts: ImagePushOpts{
				OCIPackagePath: "/path/to/oci/pkg",
			}},
			wantErr: true,
			preHook: func() {
				crPushImagesPatcher.Reset()
				crPushImagesPatcher.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
					return fi.Stat()
				}).ApplyMethod(reflect.TypeOf(&cr.Cr{}), "PushImages", func(_ *cr.Cr, _ string) error {
					return errors.New("mock cr push images")
				}).ApplyFunc(archiver.Unarchive, func(src, target string) error {
					return nil
				})
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.preHook != nil {
				tt.preHook()
			}
			if err := PushImagesWithCr(nil, tt.args.opts.OCIPackagePath, os.TempDir()); (err != nil) != tt.wantErr {
				t.Errorf("testcase: %s, PushImages() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
			if tt.postHook != nil {
				tt.postHook()
			}
		})
	}

}

func TestNewCr(t *testing.T) {
	type args struct {
		opts ImagePushOpts
	}
	fi, _ := os.CreateTemp(os.TempDir(), "testfile")
	defer fi.Close()
	defer os.Remove(filepath.Join(os.TempDir(), fi.Name()))

	crPushImagesPatcher := gomonkey.ApplyFunc(client.NewK8sClient, func() (clientDynamic dynamic.Interface, clientSet *kubernetes.Clientset) {
		return nil, nil
	})
	defer crPushImagesPatcher.Reset()

	tests := []struct {
		name     string
		args     args
		wantErr  bool
		preHook  func()
		postHook func()
	}{
		{
			name: "HelmRepoSpecified",
			args: args{opts: ImagePushOpts{
				Registry:       "test.registry",
				OCIPackagePath: "/path/to/oci/pkg",
			}},
		},
		{
			name: "K8sClientSetNil",
			args: args{opts: ImagePushOpts{
				OCIPackagePath: "/path/to/oci/pkg",
			}},
			wantErr: true,
		},
		{
			name: "K8sClientNotNil",
			args: args{opts: ImagePushOpts{
				OCIPackagePath: "/path/to/oci/pkg",
			}},
			wantErr: true,
			preHook: func() {
				crPushImagesPatcher.Reset()
				crPushImagesPatcher.ApplyFunc(client.NewK8sClient, func() (clientDynamic dynamic.Interface, clientSet *kubernetes.Clientset) {
					return nil, &kubernetes.Clientset{}
				}).ApplyFunc(configuration.LoadFromKubernetes, func(context.Context, kubernetes.Interface, ...string) (*configuration.ClusterConfig, error) {
					return nil, errors.New("test error mock")
				})
			},
		},
		{
			name: "LoadFromKubernetesNomal",
			args: args{opts: ImagePushOpts{
				OCIPackagePath: "/path/to/oci/pkg",
			}},
			wantErr: false,
			preHook: func() {
				crPushImagesPatcher.Reset()
				crPushImagesPatcher.ApplyFunc(client.NewK8sClient, func() (clientDynamic dynamic.Interface, clientSet *kubernetes.Clientset) {
					return nil, &kubernetes.Clientset{}
				}).ApplyFunc(configuration.LoadFromKubernetes, func(context.Context, kubernetes.Interface, ...string) (*configuration.ClusterConfig, error) {
					return &configuration.ClusterConfig{
						Cs: &configuration.Cs{},
					}, nil
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.preHook != nil {
				tt.preHook()
			}
			if _, err := NewCr(tt.args.opts); (err != nil) != tt.wantErr {
				t.Errorf("testcase: %s, NewCr() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
			if tt.postHook != nil {
				tt.postHook()
			}
		})
	}
}

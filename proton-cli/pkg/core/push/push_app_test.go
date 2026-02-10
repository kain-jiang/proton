package push

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client"
)

func testNewGzipFile(t *testing.T, bs []byte, fpath string) string {
	fout, err := os.CreateTemp(os.TempDir(), "testdeployApp")
	assert.Nil(t, err, "make temp deploy application file error")
	defer fout.Close()
	zipper, err := gzip.NewWriterLevel(fout, 6)
	assert.Nil(t, err, "new zipper error")
	defer zipper.Close()
	writer := tar.NewWriter(zipper)
	defer writer.Close()
	err = writer.WriteHeader(
		&tar.Header{
			Name: fpath,
			Size: int64(len(bs)),
		},
	)
	assert.Nil(t, err, "write encode app file header into temp error")
	_, err = writer.Write(bs)
	assert.Nil(t, err, "write encode app file into temp error")
	// return filepath.Join(os.TempDir(), fout.Name())
	return fout.Name()
}

func testCreatetmpAppfile(t *testing.T, app deployApplicationMeta) string {
	bs, err := json.Marshal(app)
	assert.Nil(t, err, "encode input application meta fail")
	return testNewGzipFile(t, bs, "appMeta.json")
}

func testCreatetmpErrorAppfile(t *testing.T) (string, string, string) {
	bs := []byte("test for fun")
	fpath0 := testNewGzipFile(t, bs, "appMeta.json")
	fpath1 := testNewGzipFile(t, bs, "someone")
	fout, err := os.CreateTemp(os.TempDir(), "testdeployApp")
	assert.Nil(t, err, "make temp deploy application file error")
	defer fout.Close()
	_, err = fout.Write(bs)
	assert.Nil(t, err, "make temp deploy application file error")

	return fpath0, fpath1, fout.Name()
}
func TestAPPCheckFile(t *testing.T) {
	app := &deployApplicationMeta{}
	fpath := testCreatetmpAppfile(t, *app)
	defer os.Remove(fpath)
	app.AName = "test"
	app.Type = "app/v1betav1"
	fpath0 := testCreatetmpAppfile(t, *app)
	defer os.Remove(fpath0)

	efpath0, efpath1, efpath2 := testCreatetmpErrorAppfile(t)
	defer os.Remove(efpath0)
	defer os.Remove(efpath1)
	defer os.Remove(efpath2)
	pathes := gomonkey.NewPatches()
	defer pathes.Reset()

	tests := []struct {
		name     string
		wantErr  bool
		fpath    string
		preHook  func()
		postHook func()
	}{
		{
			name:    "OpenError",
			fpath:   "unknowqwe0s8s12",
			wantErr: true,
		},
		{
			name:    "ErrUnknowFile",
			fpath:   fpath,
			wantErr: true,
		},
		{
			name:    "sucess",
			fpath:   fpath0,
			wantErr: false,
		},
		{
			name:    "ErrUnknowJsonFile",
			fpath:   efpath0,
			wantErr: true,
		},
		{
			name:    "ErrUnknowNoTargetFile",
			fpath:   efpath1,
			wantErr: true,
		},
		{
			name:    "ErrUnknowGzipFile",
			fpath:   efpath2,
			wantErr: true,
		},
		{
			name:  "IoReadAllError",
			fpath: fpath0,
			preHook: func() {
				pathes.ApplyFunc(io.ReadAll, func(io.Reader) ([]byte, error) {
					return nil, errors.New("io read mock error")
				})
			},
			wantErr: true,
		},
		{
			name:  "TarReaderNextError",
			fpath: fpath0,
			preHook: func() {
				pathes.ApplyMethod(reflect.TypeOf(&tar.Reader{}), "Next", func() (*tar.Header, error) {
					return nil, errors.New("tar reader next mock error")

				})
			},
			wantErr: true,
		},
	}

	cli := &DeployInstallerApp{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.preHook != nil {
				tt.preHook()
			}
			if err := cli.CheckFile(tt.fpath); (err != nil) != tt.wantErr {
				t.Errorf("testcase: %s, CheckFile() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
			if tt.postHook != nil {
				tt.postHook()
			}
		})
	}

}

type testReader struct {
	err error
	bs  []byte
}

func (r *testReader) Read(p []byte) (int, error) {
	n := copy(p, r.bs)
	return n, r.err
}

func TestDeployInstallerUpload(t *testing.T) {
	app := &deployApplicationMeta{
		AName: "qwe",
		Type:  "app/v1betav1",
	}
	fpath := testCreatetmpAppfile(t, *app)
	defer os.Remove(fpath)

	pathes := gomonkey.NewPatches()
	defer pathes.Reset()
	cli := &DeployInstallerApp{}

	tests := []struct {
		name     string
		wantErr  bool
		fpath    string
		preHook  func()
		postHook func()
	}{
		{
			name:    "CheckFileError",
			fpath:   "unknowqwe0s8s12",
			wantErr: true,
		},
		// {
		// 	name:    "OpenError",
		// 	fpath:   fpath,
		// 	wantErr: true,
		// 	preHook: func() {
		// 		pathes.ApplyFunc(os.Open, func(string) (*os.File, error) {
		// 			return nil, errors.New("os.Open mock error")
		// 		})
		// 	},
		// 	postHook: pathes.Reset,
		// },
		{
			name:    "NewRequestWithContextError",
			fpath:   fpath,
			wantErr: true,
			preHook: func() {
				cli.url = ":qcp//qwe"
			},
		},
		{
			name:    "HttpDoError",
			fpath:   fpath,
			wantErr: true,
			preHook: func() {
				cli.url = "http://127.0.0.1"
				pathes.ApplyMethod(reflect.TypeOf(&http.Client{}), "Do", func(*http.Client, *http.Request) (*http.Response, error) {
					return nil, errors.New("http client do mock error")
				})
			},
			postHook: pathes.Reset,
		},
		{
			name:    "ErrAPPConflict",
			fpath:   fpath,
			wantErr: true,
			preHook: func() {
				pathes.ApplyMethod(reflect.TypeOf(&http.Client{}), "Do", func(*http.Client, *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 409,
						Body:       io.NopCloser(nil),
					}, nil
				})
			},
			postHook: pathes.Reset,
		},
		{
			name:    "ErrClient",
			fpath:   fpath,
			wantErr: true,
			preHook: func() {
				pathes.ApplyMethod(reflect.TypeOf(&http.Client{}), "Do", func(*http.Client, *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 400,
						Body:       io.NopCloser(strings.NewReader("test")),
					}, nil
				})
			},
			postHook: pathes.Reset,
		},
		{
			name:    "ErrClientbody",
			fpath:   fpath,
			wantErr: true,
			preHook: func() {
				pathes.ApplyMethod(reflect.TypeOf(&http.Client{}), "Do", func(*http.Client, *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 400,
						Body: io.NopCloser(&testReader{
							err: errors.New("mock readall"),
						}),
					}, nil
				})
			},
			postHook: pathes.Reset,
		}, {
			name:    "ErrClientUnknow",
			fpath:   fpath,
			wantErr: true,
			preHook: func() {
				pathes.ApplyMethod(reflect.TypeOf(&http.Client{}), "Do", func(*http.Client, *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 500,
						Body:       io.NopCloser(strings.NewReader("test")),
					}, nil
				})
			},
			postHook: pathes.Reset,
		},
		{
			name:    "ErrClientUnknowbody",
			fpath:   fpath,
			wantErr: true,
			preHook: func() {
				pathes.ApplyMethod(reflect.TypeOf(&http.Client{}), "Do", func(*http.Client, *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 500,
						Body: io.NopCloser(&testReader{
							err: errors.New("mock readall"),
						}),
					}, nil
				})
			},
			postHook: pathes.Reset,
		},
		{
			name:    "sucess",
			fpath:   fpath,
			wantErr: false,
			preHook: func() {
				pathes.ApplyMethod(reflect.TypeOf(&http.Client{}), "Do", func(*http.Client, *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 200,
						Body:       io.NopCloser(nil),
					}, nil
				})
			},
			postHook: pathes.Reset,
		},
	}
	lg := logrus.New()
	lg.SetOutput(io.Discard)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.preHook != nil {
				tt.preHook()
			}
			if _, err := cli.Upload(context.Background(), lg, tt.fpath); (err != nil) != tt.wantErr {
				t.Errorf("testcase: %s, CheckFile() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
			if tt.postHook != nil {
				tt.postHook()
			}
		})
	}
}

func TestNewDeployInstallerClient(t *testing.T) {
	fc := fake.NewSimpleClientset(
		&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deploy-installer",
				Namespace: "testdeployinstaller",
			},
		},
	)
	patches := gomonkey.NewPatches()
	defer patches.Reset()
	tests := []struct {
		name      string
		namespace string
		wantErr   bool
		preHook   func()
		postHook  func()
	}{
		{
			name:    "GetNilK8sHTTPClient",
			wantErr: true,
			preHook: func() {
				patches.ApplyFunc(client.NewK8sHTTPClient, func() (*http.Client, error) {
					return nil, errors.New("mock get nil k8s httpclient")
				})
			},
			postHook: func() {
				patches.Reset()
			},
		},
		{
			name:    "GetNilK8sClient",
			wantErr: true,
			preHook: func() {
				patches.ApplyFunc(client.NewK8sHTTPClient, func() (*http.Client, error) {
					return &http.Client{}, nil
				})
				patches.ApplyFunc(client.NewK8sClientInterface, func() (dynamic.Interface, kubernetes.Interface) {
					return nil, nil
				})
			},
			postHook: func() {
				patches.Reset()
			},
		},
		{
			name:      "DeployInstallerNotInstalled",
			namespace: "unknowqweouq",
			wantErr:   true,
			preHook: func() {
				patches.ApplyFunc(client.NewK8sClientInterface, func() (dynamic.Interface, kubernetes.Interface) {
					return nil, fc
				})
				patches.ApplyFunc(client.NewK8sHTTPClient, func() (*http.Client, error) {
					return &http.Client{}, nil
				})
				patches.ApplyFunc(rest.NewRequest, func(*rest.RESTClient) *rest.Request {
					return &rest.Request{}
				})
				patches.ApplyMethod(reflect.TypeOf(&rest.Request{}), "URL", func(*rest.Request) *url.URL {
					return &url.URL{}
				})
			},
		},
		{
			name:      "sucess",
			namespace: "testdeployinstaller",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.preHook != nil {
				tt.preHook()
			}
			if _, err := NewDeployInstallerClient(context.Background(), tt.namespace, false); (err != nil) != tt.wantErr {
				t.Errorf("testcase: %s, NewDeployInstallerClient() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
			if tt.postHook != nil {
				tt.postHook()
			}
		})
	}

	{
		u, err := NewDeployInstallerClient(context.Background(), "testdeployinstaller", false)
		if err != nil {
			t.Errorf("init fail")
		}
		app := &deployApplicationMeta{
			AName: "qwe",
			Type:  "app/v1betav1",
		}
		fpath := testCreatetmpAppfile(t, *app)
		defer os.Remove(fpath)
		if err := u.CheckFile(fpath); err != nil {
			t.Errorf("init fail")
		}
		_, _ = u.Upload(context.Background(), logrus.New(), fpath)
	}

}

package push

import (
	"context"
	"encoding/hex"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testCreateTmpFile(t *testing.T, bs []byte) string {
	fout, err := os.CreateTemp(os.TempDir(), "testfile")
	assert.Nil(t, err, "make temp file error")
	_, err = fout.Write(bs)
	assert.Nil(t, err, "write temp file error")
	fout.Close()
	return fout.Name()
}

func TestManifestUpload(t *testing.T) {
	cli := &deployInstaller{
		Service:   "test",
		Namespace: "test",
		khttpCli:  &http.Client{},
	}
	u := DeployInstallerManifest{
		deployInstaller: cli,
		url:             "test",
	}
	// create tmp file
	obj := _DeployInstllerManifestsMagic + "{}" + _DeployInstllerManifestsMagic
	obj = hex.EncodeToString([]byte(obj))
	fpath := testCreateTmpFile(t, []byte(obj))

	obj = _DeployInstllerManifestsMagic + "{}" + _DeployInstllerManifestsMagic + "qweqw"
	obj = hex.EncodeToString([]byte(obj))
	fpath0 := testCreateTmpFile(t, []byte(obj))
	obj = _DeployInstllerManifestsMagic + "{}"
	obj = hex.EncodeToString([]byte(obj))
	fpath1 := testCreateTmpFile(t, []byte(obj))
	// create tmp file

	tests := []struct {
		name     string
		wantErr  bool
		fpath    string
		preHook  func()
		postHook func()
	}{
		{
			name:    "success",
			wantErr: false,
			fpath:   fpath,
		},
		{
			name:    "nofile",
			wantErr: true,
			fpath:   "nofile",
		},
		{
			name:    "nomagic",
			wantErr: true,
			fpath:   fpath0,
		},
		{
			name:    "small",
			wantErr: true,
			fpath:   fpath1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.preHook != nil {
				tt.preHook()
			}
			if err := u.CheckFile(tt.fpath); (err != nil) != tt.wantErr {
				t.Errorf("testcase: %s, CheckFile() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
			if tt.postHook != nil {
				tt.postHook()
			}
		})
	}

	_, _ = u.Upload(context.Background(), log, fpath)
}

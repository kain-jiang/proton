package client

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/go-test/deep"
	utilsexec "k8s.io/utils/exec"
)

func Test_handleHelmStderr(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "nil",
			args: args{
				err: nil,
			},
			wantErr: nil,
		},
		{
			name: "chart-repository-not-found",
			args: args{
				err: &utilsexec.ExitErrorWrapper{
					ExitError: &exec.ExitError{
						Stderr: []byte("Error: no repo named \"stable\" found\n"),
					},
				},
			},
			wantErr: ErrChartRepositoryNotFound,
		},
		{
			name: "unauthorized",
			args: args{
				err: &utilsexec.ExitErrorWrapper{
					ExitError: &exec.ExitError{
						Stderr: []byte("XXXX: 401 Unauthorized"),
					},
				},
			},
			wantErr: ErrHelmRepositoryUnauthorized,
		},
		{
			name: "not-found-tiller",
			args: args{
				err: &utilsexec.ExitErrorWrapper{
					ExitError: &exec.ExitError{
						Stderr: []byte("XXXX: could not find tiller"),
					},
				},
			},
			wantErr: ErrHelmNotFindTiller,
		},
		{
			name: "not-found-ready-tiller-pod",
			args: args{
				err: &utilsexec.ExitErrorWrapper{
					ExitError: &exec.ExitError{
						Stderr: []byte("XXXX: could not find a ready tiller pod"),
					},
				},
			},
			wantErr: ErrHelmNotFindReadyTillerPod,
		},
		{
			name: "release-not-found",
			args: args{
				err: &utilsexec.ExitErrorWrapper{
					ExitError: &exec.ExitError{
						Stderr: []byte(`Error: release: "hello-world" not found`),
					},
				},
			},
			wantErr: ErrHelmReleaseNotFound,
		},
		{
			name: "other",
			args: args{
				err: errors.New("hello"),
			},
			wantErr: errors.New("hello"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := handleHelmStderr(tt.args.err)
			for _, diff := range deep.Equal(gotErr, tt.wantErr) {
				t.Errorf("handleHelmStderr() diff: %v", diff)
			}
		})
	}
}

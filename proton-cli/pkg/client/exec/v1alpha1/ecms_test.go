package v1alpha1

import (
	"context"
	"errors"
	"io"
	"testing"

	ecmsexec "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/ecms/v1alpha1/exec"
)

type fakeECMSExec struct {
	lastOpts ecmsexec.ExecuteOptions
	out      []byte
	err      error
}

func (f *fakeECMSExec) Execute(_ context.Context, _ []string, _ io.Reader, opts ecmsexec.ExecuteOptions) ([]byte, error) {
	f.lastOpts = opts
	return f.out, f.err
}

func TestECMSCommandRunReturnsRemoteStderr(t *testing.T) {
	fake := &fakeECMSExec{
		out: []byte("kubeadm init failed: port 6443 is in use"),
		err: &ecmsexec.ExitError{
			Output:   []byte("kubeadm init failed: port 6443 is in use"),
			ExitCode: 1,
		},
	}

	cmd := &ECMSCommand{
		c:       fake,
		command: []string{"kubeadm", "init"},
	}

	err := cmd.Run()
	if err == nil {
		t.Fatal("expected error")
	}

	var exitErr *ErrExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ErrExitError, got %T", err)
	}
	if exitErr.ExitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitErr.ExitCode)
	}
	if string(exitErr.Stderr) != "kubeadm init failed: port 6443 is in use" {
		t.Fatalf("unexpected stderr: %q", string(exitErr.Stderr))
	}
	if !fake.lastOpts.Stderr {
		t.Fatal("expected Run to request stderr from remote executor")
	}
}

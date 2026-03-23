package slb

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	ecms "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/ecms/v1alpha1"
	ecmsfiles "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/ecms/v1alpha1/files"
	execv1alpha1 "devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/exec/v1alpha1"
)

type Remote struct {
	files ecmsfiles.Interface
	exec  execv1alpha1.Executor
}

func NewRemote(c ecms.Interface) *Remote {
	return &Remote{
		files: c.Files(),
		exec:  execv1alpha1.NewECMSExecutorForHost(c.Exec()),
	}
}

func (r *Remote) readFileIfExists(ctx context.Context, path string) ([]byte, bool, error) {
	data, err := r.files.ReadFile(ctx, path)
	if err == nil {
		return data, true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return nil, false, nil
	}
	return nil, false, err
}

func (r *Remote) ensureDir(ctx context.Context, path string) error {
	if path == "" || path == "." || path == "/" {
		return nil
	}

	info, err := r.files.Stat(ctx, path)
	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("%s is not a directory", path)
		}
		return nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	return r.exec.Command("mkdir", "-p", path).Run()
}

func (r *Remote) writeFile(ctx context.Context, path string, data []byte) error {
	if err := r.ensureDir(ctx, filepath.Dir(path)); err != nil {
		return err
	}
	return r.files.Create(ctx, path, false, data)
}

func (r *Remote) deleteFileIfExists(ctx context.Context, path string) error {
	err := r.files.Delete(ctx, path)
	if err == nil || errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}

func (r *Remote) copyFile(ctx context.Context, src, dst string) error {
	data, err := r.files.ReadFile(ctx, src)
	if err != nil {
		return err
	}
	return r.writeFile(ctx, dst, data)
}

func (r *Remote) shellOutput(script string) ([]byte, error) {
	return r.exec.Command("bash", "-lc", script).Output()
}

func (r *Remote) testCommand(script string) (string, error) {
	out, err := r.shellOutput(script)
	if err == nil {
		return "", nil
	}
	if len(out) != 0 {
		return string(out), nil
	}
	return err.Error(), nil
}

func (r *Remote) resetFailedService(name string) {
	_ = r.exec.Command("systemctl", "reset-failed", name).Run()
}

func (r *Remote) serviceActive(name string) bool {
	return r.exec.Command("systemctl", "status", name).Run() == nil
}

package check

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"strings"

	"devops.aishu.cn/AISHUDevOps/ICT/_git/proton-opensource.git/proton-cli/v3/pkg/client/ecms/v1alpha1/files"
)

// NodeNodeDirAvailableChecker 检查指定指定节点的路径不存在或是目录
type NodeDirAvailableChecker struct {
	// Node name
	Node string

	Path string

	// ecms/v1alpha1/Files
	Files files.Interface
}

// Check implements Checker.
func (c *NodeDirAvailableChecker) Check() (warningList []error, errorList []error) {
	var ctx = context.TODO()
	info, err := c.Files.Stat(ctx, c.Path)
	// If it doesn't exist we are good:
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if !info.IsDir() {
		return nil, []error{fmt.Errorf("%v is not a directory", c.Path)}
	}
	return
}

// Name implements Checker.
func (c *NodeDirAvailableChecker) Name() string {
	return fmt.Sprintf("%s-NodeDirAvailableChecker-%s", c.Node, strings.ReplaceAll(c.Path, "/", "-"))
}

var _ Checker = (*NodeDirAvailableChecker)(nil)

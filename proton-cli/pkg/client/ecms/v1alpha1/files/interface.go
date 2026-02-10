package files

import (
	"context"
	"io/fs"
)

// TODO: move pkg/client/ecms/files/v1alpha1 to pkg/client/ecms/v1alpha1/files
type Interface interface {
	// 创建文件或目录
	Create(ctx context.Context, path string, isDir bool, data []byte) error
	// 删除文件、目录
	Delete(ctx context.Context, path string) error
	// 更新文件内容
	Update(ctx context.Context, path string, data []byte) error
	// 获取文件、目录元数据
	Stat(ctx context.Context, path string) (fs.FileInfo, error)
	// 读取文件内容
	ReadFile(ctx context.Context, path string) ([]byte, error)
	// 读取目录列表
	ListDirectory(ctx context.Context, path string) ([]fs.FileInfo, error)
	// 移动文件、目录
	Rename(ctx context.Context, src, dst string) error
}

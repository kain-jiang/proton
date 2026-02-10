package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// TraverseDirectory 遍历目录，返回所有非隐藏文件的路径，递归处理子目录
func TraverseDirectory(root string, f func(path string, info os.FileInfo, err error) error) error {
	_, err := os.Stat(root)
	if os.IsNotExist(err) {
		return nil
	}

	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err // 遇到错误直接返回
		}

		// 忽略隐藏文件或目录（以"."开头的文件或目录）
		if strings.HasPrefix(filepath.Base(info.Name()), ".") {
			if info.IsDir() {
				return filepath.SkipDir // 跳过隐藏目录
			}
			return nil // 跳过隐藏文件
		}

		// 如果是文件，添加到结果列表
		if info.IsDir() {
			return nil
		}
		return f(path, info, err)
	})
}

// TraverseDirectoryNotSub 遍历下子目录，返回所有非隐藏文件的路径，不递归处理子目录
func TraverseDirectoryNotSub(root string, f func(path string) error) error {
	files, err := os.ReadDir(root)
	if err != nil {
		return err
	}

	for _, file := range files {
		// 如果是隐藏文件或目录，跳过
		if file.Name()[0] == '.' {
			continue
		}
		// 只处理文件，不处理子目录
		if file.IsDir() {
			svcName := file.Name()
			err = f(filepath.Join(root, svcName))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

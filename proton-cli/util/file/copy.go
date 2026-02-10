package file

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

// 判断所给路径文件/文件夹是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	//isnotexist来判断，是不是不存在的错误
	if os.IsNotExist(err) { //如果返回的错误类型使用os.isNotExist()判断为true，说明文件或者文件夹不存在
		return false, nil
	}
	return false, err //如果有错误了，但是不是不存在的错误，所以把这个错误原封不动的返回
}

// 字节的单位转换 保留两位小数
func FormatFileSize(fileSize int64) (size string) {
	if fileSize < 1024 {
		//return strconv.FormatInt(fileSize, 10) + "B"
		return fmt.Sprintf("%.2fB", float64(fileSize)/float64(1))
	} else if fileSize < (1024 * 1024) {
		return fmt.Sprintf("%.2fKB", float64(fileSize)/float64(1024))
	} else if fileSize < (1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fMB", float64(fileSize)/float64(1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fGB", float64(fileSize)/float64(1024*1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fTB", float64(fileSize)/float64(1024*1024*1024*1024))
	} else { //if fileSize < (1024 * 1024 * 1024 * 1024 * 1024 * 1024)
		return fmt.Sprintf("%.2fPB", float64(fileSize)/float64(1024*1024*1024*1024*1024))
	}
}

/**
 * 拷贝文件夹,同时拷贝文件夹中的文件
 * @param srcPath  		需要拷贝的文件夹路径: D:/test
 * @param destPath		拷贝到的位置: D:/backup/
 */
func Copy(srcPath string, destPath string, log *logrus.Logger, exclude []string) error {

	exist, err := PathExists(srcPath)
	if err != nil {
		log.Errorln("检测复制源目录或者文件存在异常：", srcPath, err)
		return err
	}
	if !exist {
		log.Errorln("复制源目录或者文件不存在：", srcPath)
		return errors.New("复制源目录或者文件不存在：" + srcPath)
	}

	err = filepath.Walk(srcPath, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}

		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		path = strings.Replace(path, "\\", "/", -1)
		dst := strings.Replace(path, srcPath, destPath, -1)
		if exclude != nil && (slices.Contains(exclude, path) || slices.Contains(exclude, dst)) {
			log.Println("skip path copy:", path)
			return nil
		}
		switch {
		case info.IsDir():
			// 创建目录
			log.Println("create directory:", dst)
			return os.MkdirAll(dst, info.Mode())
		case info.Mode().IsRegular():
			dir := filepath.Dir(dst)
			if exist, err := PathExists(dir); err != nil {
				return err
			} else if !exist {
				err := os.MkdirAll(dir, os.ModePerm)
				if err != nil {
					return err
				}
			}
			// 创建文件
			log.Println("create file:", dst)
			w, err := os.Create(dst)
			if err != nil {
				return err
			}
			defer w.Close()
			r, err := os.Open(path)
			if err != nil {
				return err
			}
			defer r.Close()
			_, err = io.Copy(w, r)
			return err
		case info.Mode()&fs.ModeSymlink != 0:
			// 创建符号链接
			if exist, err = PathExists(dst); err != nil {
				return err
			}
			if exist {
				err := os.Remove(dst)
				if err != nil {
					return err
				}
			}
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			log.Println("create symbolic link:", dst, "->", target)
			return os.Symlink(target, dst)
		default:
			log.Errorf("invalid: %v, mode: %v", path, info.Mode())
			return errors.New("invalid: " + path + " mode:" + info.Mode().String())
		}
	})
	if err != nil {
		fmt.Print(err.Error())
	}
	return err
}

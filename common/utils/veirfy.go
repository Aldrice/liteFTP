package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// VerifyParams 用于检查参数数量是否符合预期, 返回true说明数量符合预期，并且返回参数数组
func VerifyParams(ps string, n int) ([]string, bool) {
	params := strings.Split(ps, " ")
	// 当截取得到的参数不同于预期时
	if len(params) != n || params[0] == "" {
		return nil, false
	}
	return params, true
}

// VerifyPath 检查路径是否存在
func VerifyPath(p string) (bool, error) {
	_, err := os.Stat(p)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// VerifyFolderName 输入路径, 检查文件夹名是否符合规范
func VerifyFolderName(path string) bool {
	name := filepath.Base(path)
	switch runtime.GOOS {
	case "windows":
		if strings.ContainsAny(name, "\\/:*?\"<>|") {
			return false
		}
		return true
	case "linux":
		if strings.ContainsAny(name, "/ ") {
			return false
		}
		return true

	default:
		return false
	}
}

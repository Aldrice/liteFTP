package utils

import (
	"os"
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

// IsDir 检查是否是文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

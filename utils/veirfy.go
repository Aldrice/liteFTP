package utils

import (
	"errors"
	"os"
	"strings"
	"syscall"
)

// VerifyParams 用于检查参数数量是否符合预期, 返回true说明数量符合预期, 并且返回参数数组 (仅用于无空格参数的情况)
func VerifyParams(ps string, n int) ([]string, bool) {
	params := strings.SplitN(ps, " ", n)
	if len(params) != n {
		return nil, false
	}
	for i, param := range params {
		params[i] = strings.TrimSpace(param)
		// 当截取得到的参数不同于预期时
		if params[i] == "" {
			return nil, false
		}
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

// IsPortInuse 检查错误是否为 端口已被占用
func IsPortInuse(err error) bool {
	// 载入错误
	var eOsSyscall *os.SyscallError
	if !errors.As(err, &eOsSyscall) {
		return false
	}
	var errErrno syscall.Errno
	if !errors.As(eOsSyscall, &errErrno) {
		return false
	}
	// 检查端口是否被使用
	if errErrno == syscall.EADDRINUSE {
		return true
	}
	return false
}

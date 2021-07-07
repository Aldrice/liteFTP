package core

import (
	"fmt"
	"strings"
)

type rsp struct {
	code int
	info string
}

func newResponse(c int, i ...string) *rsp {
	info := i[0]
	if len(i) > 1 {
		return &rsp{
			code: c,
			info: fmt.Sprintf("%s\nError Message: %s", info, i[1]),
		}
	}
	return &rsp{
		code: c,
		info: info,
	}
}

// formatResponse 实现response内容的规范化, 以适应协议要求
// 参考: https://www.cnblogs.com/depend-wind/articles/14026572.html
func (r *rsp) formatResponse() string {
	if strings.Contains(r.info, "\n") {
		return fmt.Sprintf("%d-%s\r\n%d \r\n", r.code, r.info, r.code)
	}
	return fmt.Sprintf("%d %s\r\n", r.code, r.info)
}

var (
	rspTextWelcome      = "KLL FTP server ready."
	rspTextTempReceived = "Input password to login."
	rspTextLoginSuccess = "Login successfully."
)

var (
	rspSyntaxError = &rsp{code: 500, info: "Syntax error, command unrecognized."}
	rspParamsError = &rsp{code: 504, info: "Command not implemented for that parameter."}
)

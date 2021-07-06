package core

import (
	"fmt"
	"strings"
)

type rsp struct {
	code int
	info string
}

func createResponse(c int, i ...string) *rsp {
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
	rspWelcome      = &rsp{code: 220, info: "KLL FTP server ready."}
	rspTempReceived = &rsp{code: 331, info: "Input password to login."}
	rspLoginSuccess = &rsp{code: 200, info: "Login successfully."}

	rspSyntaxError  = &rsp{code: 500, info: "Syntax error, command unrecognized."}
	rspProcessError = &rsp{code: 553, info: "An error occur in the server, requested action not taken."}
	rspParamsError  = &rsp{code: 504, info: "Command not implemented for that parameter."}
	rspAuthError    = &rsp{code: 530, info: "User no auth to execute this command."}
)

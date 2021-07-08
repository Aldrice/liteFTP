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
			info: fmt.Sprintf("%s\r\nError Message: %s\r\n", info, i[1]),
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
	// 肯定回应
	rspTextWelcome      = "KLL FTP server ready."
	rspTextTempReceived = "Input password to login."
	rspTextLoginSuccess = "Login successfully."
	// 否定回应
	rspTextPathError     = "The pathname wasn't exist or no authorization to process."
	rspTextSendError     = "An error occur when sending text to client."
	rspTextDataBaseError = "An error occur when processing the database."
	rspTextProcessError  = "An error occur when server handle os process."
	rspTextAuthError     = "You have no authorization to use this command."
)

var (
	rspSyntaxError = &rsp{code: 500, info: "Syntax error, command unrecognized."}
	rspParamsError = &rsp{code: 504, info: "Command not implemented for that parameter."}
)

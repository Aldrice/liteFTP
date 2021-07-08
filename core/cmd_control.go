package core

import (
	"fmt"
	"github.com/Aldrice/liteFTP/utils"
	"net"
	"os"
	"path/filepath"
	"strings"
)

var OPTS = &command{
	name:        []string{"OPTS"},
	demandAuth:  false,
	demandLogin: false,
	demandParam: true,
	cmdFunction: func(conn *Connection, params string) (*rsp, error) {
		// 参考: https://www.serv-u.com/resource/tutorial/feat-opts-help-stat-nlst-xcup-xcwd-ftp-command#de323b8e-a756-470d-9544-bdab18b5644b
		ps, ok := utils.VerifyParams(params, 2)
		if !ok {
			return rspParamsError, nil
		}
		switch ps[0] {
		case "UTF8":
			if ps[1] != "ON" {
				return newResponse(550, "The server only support utf_8 transferring."), nil
			} else {
				return newResponse(200, "Server are now transmit with utf_8 encoding."), nil
			}
		default:
		}
		return rspSyntaxError, nil
	},
}

var PASV = &command{
	name:        []string{"PASV"},
	demandAuth:  false,
	demandLogin: true,
	demandParam: false,
	cmdFunction: func(conn *Connection, params string) (*rsp, error) {
		conn.isPassive = true
		err := conn.establishConn(conn.linkConn.LocalAddr().(*net.TCPAddr))
		if err != nil {
			return newResponse(421, "An error occur when establishing the connection.", err.Error()), nil
		}
		return newResponse(227,
			fmt.Sprintf("Entering Passive Mode (%s)", utils.FormatAddr(conn.pasvAddr)),
		), nil
	},
}

var FEAT = &command{
	name:        []string{"FEAT"},
	demandAuth:  false,
	demandLogin: false,
	demandParam: false,
	cmdFunction: func(conn *Connection, params string) (*rsp, error) {
		return newResponse(221, "Extensions supported:\r\nUTF8\r\n"), nil
	},
}

var QUIT = &command{
	name:        []string{"QUIT"},
	demandAuth:  false,
	demandLogin: false,
	demandParam: false,
	cmdFunction: func(conn *Connection, params string) (*rsp, error) {
		_ = conn.linkConn.Close()
		return &rsp{code: 221, info: "Bye."}, nil
	},
}

var SYST = &command{
	name:        []string{"SYST"},
	demandAuth:  false,
	demandLogin: false,
	demandParam: false,
	cmdFunction: func(conn *Connection, params string) (*rsp, error) {
		return &rsp{code: 215, info: "UNIX Type: L8"}, nil
	},
}

var NOOP = &command{
	name:        []string{"NOOP"},
	demandAuth:  false,
	demandLogin: false,
	demandParam: false,
	cmdFunction: func(conn *Connection, params string) (*rsp, error) {
		return &rsp{
			code: 200,
			info: "",
		}, nil
	},
}

// TYPE 是否开启二进制传输
// 参考: https://cr.yp.to/ftp/type.html
var TYPE = &command{
	name:        []string{"TYPE"},
	demandAuth:  false,
	demandLogin: true,
	demandParam: true,
	cmdFunction: func(conn *Connection, params string) (*rsp, error) {
		switch strings.ToUpper(strings.TrimSpace(params)) {
		case "A":
			return newResponse(200, "ASCII mode on."), nil
		case "I":
			return newResponse(200, "Binary mode on."), nil
		default:
		}
		return rspSyntaxError, nil
	},
}

var CDUP = &command{
	name:        []string{"CDUP", "XCUP"},
	demandAuth:  false,
	demandLogin: true,
	demandParam: false,
	cmdFunction: func(conn *Connection, params string) (*rsp, error) {
		if conn.workDir != conn.authDir {
			ok, err := conn.setLiedDir(filepath.Dir(conn.workDir))
			if err != nil {
				return newResponse(550, "An error occur when setting the dir.", err.Error()), nil
			}
			if ok {
				return newResponse(200, "Okay."), nil
			}
		}
		return newResponse(550, "No further upper path."), nil
	},
}

// MKD 注: 只允许用户逐级创建文件夹
var MKD = &command{
	name:        []string{"MKD", "XMKD"},
	demandAuth:  false,
	demandLogin: true,
	demandParam: true,
	cmdFunction: func(conn *Connection, params string) (*rsp, error) {
		ps, ok := utils.VerifyParams(params, 1)
		if !ok {
			return rspParamsError, nil
		}
		// 处理路径
		newPath := conn.processPath(ps[0])
		if newPath == "" {
			return newResponse(550, rspTextPathError), nil
		}
		if err := os.Mkdir(newPath, os.ModePerm); err != nil {
			return newResponse(550, "An error occur when the server creating the new file.", err.Error()), nil
		}
		return newResponse(250, "Directory created."), nil
	},
}

var PWD = &command{
	name:        []string{"PWD", "XPWD"},
	demandAuth:  false,
	demandLogin: true,
	demandParam: false,
	cmdFunction: func(conn *Connection, params string) (*rsp, error) {
		return &rsp{
			code: 257,
			info: fmt.Sprintf("\"%s\"", conn.workDir),
		}, nil
	},
}

var CWD = &command{
	name:        []string{"CWD", "XCWD"},
	demandAuth:  false,
	demandLogin: true,
	demandParam: true,
	cmdFunction: func(conn *Connection, params string) (*rsp, error) {
		ps, ok := utils.VerifyParams(params, 1)
		if !ok {
			return rspParamsError, nil
		}
		newPath := conn.processPath(ps[0])
		if newPath == "" {
			return newResponse(550, fmt.Sprintf("%s: No such dictionary.", ps[0])), nil
		}
		ok, err := conn.setLiedDir(newPath)
		if err != nil {
			return newResponse(550, "An error occur when setting the dir.", err.Error()), nil
		}
		if !ok {
			return newResponse(550, fmt.Sprintf("%s: No such dictionary.", ps[0])), nil
		}
		return newResponse(250, "Okay."), nil
	},
}

var RMD = &command{
	name:        []string{"RMD", "XRMD"},
	demandAuth:  false,
	demandLogin: true,
	demandParam: true,
	cmdFunction: func(conn *Connection, params string) (*rsp, error) {
		// 检查参数数量
		ps, ok := utils.VerifyParams(params, 1)
		if !ok {
			return rspParamsError, nil
		}
		// 处理路径
		newPath := strings.Replace(conn.processPath(ps[0]), "/", "\\", -1)
		// 不允许用户删除根目录, 也不允许删除用户的工作路径下的目录, 也不允许删除文件
		if newPath == "" || newPath == conn.authDir || newPath == conn.workDir || !utils.IsDir(newPath) {
			return &rsp{
				code: 550,
				info: rspTextPathError,
			}, nil
		}
		// 执行删除
		if err := os.Remove(newPath); err != nil {
			return newResponse(550, "An error occur when the server removing the specify dictionary.", err.Error()), nil
		}
		return newResponse(250, "Directory removed."), nil
	},
}

var DELE = &command{
	name:        []string{"DELE"},
	demandAuth:  false,
	demandLogin: true,
	demandParam: true,
	cmdFunction: func(conn *Connection, params string) (*rsp, error) {
		// 检查参数数量
		ps, ok := utils.VerifyParams(params, 1)
		if !ok {
			return rspParamsError, nil
		}
		// 处理路径
		newPath := conn.processPath(ps[0])
		if newPath == "" || utils.IsDir(newPath) {
			return newResponse(550, rspTextPathError), nil
		}
		if err := os.Remove(newPath); err != nil {
			return newResponse(550, "An error occur when the server removing the specify file.", err.Error()), nil
		}
		return newResponse(250, "File removed."), nil
	},
}

var PORT = &command{
	name:        []string{"PORT"},
	demandAuth:  false,
	demandLogin: true,
	demandParam: true,
	cmdFunction: func(conn *Connection, params string) (*rsp, error) {
		ps := strings.SplitN(params, ",", 6)
		addr := utils.WrapAddr(ps)
		if addr == nil {
			return newResponse(501, "The host port parameter is invalid."), nil
		}
		if err := conn.establishConn(addr); err != nil {
			return newResponse(550, "An error occur when establishing the connection.", err.Error()), nil
		}
		return newResponse(200, "Establishing connection succeed."), nil
	},
}

var SIZE = &command{
	name:        []string{"SIZE"},
	demandAuth:  false,
	demandLogin: true,
	demandParam: true,
	cmdFunction: func(conn *Connection, params string) (*rsp, error) {
		// 检查参数数量
		ps, ok := utils.VerifyParams(params, 1)
		if !ok {
			return rspParamsError, nil
		}
		// 检查文件是否存在以及是否有权限访问
		newPath := conn.processPath(ps[0])
		if newPath == "" {
			return newResponse(550, rspTextPathError), nil
		}
		fileState, err := os.Stat(newPath)
		if err != nil {
			return newResponse(550, "An error occur when stating the file.", err.Error()), nil
		}
		if fileState.IsDir() {
			return newResponse(550, rspTextPathError), nil
		}
		return newResponse(213, fmt.Sprintf("%d", fileState.Size())), nil
	},
}

var RNFR = &command{
	name:        []string{"RNFR"},
	demandAuth:  false,
	demandLogin: true,
	demandParam: true,
	cmdFunction: func(conn *Connection, params string) (*rsp, error) {
		// 检查参数数量
		ps, ok := utils.VerifyParams(params, 1)
		if !ok {
			return rspParamsError, nil
		}
		// 检查文件是否有权限访问
		newPath := conn.processPath(ps[0])
		if newPath == "" {
			return newResponse(550, rspTextPathError), nil
		}
		// 检查文件或文件夹是否存在
		exist, err := utils.VerifyPath(newPath)
		if err != nil {
			return newResponse(550, "An error occur when verifying the file.", err.Error()), nil
		}
		if !exist {
			return newResponse(550, rspTextPathError), nil
		}
		// 记录原文件或文件夹位置
		conn.renamePath = newPath
		return newResponse(350, "Waiting for next instruction."), nil
	},
}

var RNTO = &command{
	name:        []string{"RNTO"},
	demandAuth:  false,
	demandLogin: true,
	demandParam: true,
	cmdFunction: func(conn *Connection, params string) (*rsp, error) {
		// 检查参数数量
		ps, ok := utils.VerifyParams(params, 1)
		if !ok {
			return rspParamsError, nil
		}
		// 检查文件是否存在以及是否有权限访问
		newPath := conn.processPath(ps[0])
		if newPath == "" {
			return newResponse(550, rspTextPathError), nil
		}
		// 检查现有路径是否已有同名文件
		exist, err := utils.VerifyPath(newPath)
		if err != nil {
			return newResponse(550, "An error occur when verifying the file.", err.Error()), nil
		}
		if exist {
			return newResponse(553, "An file already exist in this path."), nil
		}
		// 移动文件或文件夹
		if err := os.Rename(conn.renamePath, newPath); err != nil {
			return newResponse(553, "An error occur when renaming the file.", err.Error()), nil
		}
		return newResponse(250, "File renamed."), nil
	},
}

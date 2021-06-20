package core

import (
	"fmt"
	"github.com/Aldrice/liteFTP/common/utils"
	"os"
	"path/filepath"
	"strings"
)

var OPTS = &command{
	name:        []string{"OPTS"},
	demandAuth:  false,
	demandLogin: false,
	demandParam: true,
	cmdFunction: func(conn *Connection, params string) (*response, error) {
		// todo: 需要完善
		_, ok := utils.VerifyParams(params, 2)
		if !ok {
			return respParamsError, nil
		}
		// 参考: https://www.serv-u.com/resource/tutorial/feat-opts-help-stat-nlst-xcup-xcwd-ftp-command#de323b8e-a756-470d-9544-bdab18b5644b
		if !conn.server.enableUTF8 {
			return &response{
				code: 202,
				info: "Server do not allow utf_8 encoding transmission.",
			}, nil
		}
		return &response{
			code: 200,
			info: "Server are now transmit with utf_8 encoding.",
		}, nil
	},
}

var PASV = &command{
	name:        []string{"PASV"},
	demandAuth:  false,
	demandLogin: false,
	demandParam: true,
	cmdFunction: func(conn *Connection, params string) (*response, error) {
		return respParamsError, nil
	},
}

var FEAT = &command{
	name:        []string{"FEAT"},
	demandAuth:  false,
	demandLogin: false,
	demandParam: false,
	cmdFunction: func(conn *Connection, params string) (*response, error) {

		return nil, nil
	},
}

var QUIT = &command{
	name:        []string{"QUIT"},
	demandAuth:  false,
	demandLogin: false,
	demandParam: false,
	cmdFunction: func(conn *Connection, params string) (*response, error) {
		_ = conn.linkConn.Close()
		return &response{code: 221, info: "Bye."}, nil
	},
}

var SYST = &command{
	name:        []string{"SYST"},
	demandAuth:  false,
	demandLogin: false,
	demandParam: false,
	cmdFunction: func(conn *Connection, params string) (*response, error) {
		return &response{code: 215, info: "UNIX Type: L8"}, nil
	},
}

var NOOP = &command{
	name:        []string{"NOOP"},
	demandAuth:  false,
	demandLogin: false,
	demandParam: false,
	cmdFunction: func(conn *Connection, params string) (*response, error) {
		return &response{
			code: 200,
			info: "",
		}, nil
	},
}

// TYPE 是否开启二进制传输
// todo: 需要完善
// 参考: https://cr.yp.to/ftp/type.html
var TYPE = &command{
	name:        []string{"TYPE"},
	demandAuth:  false,
	demandLogin: true,
	demandParam: true,
	cmdFunction: func(conn *Connection, params string) (*response, error) {
		if conn.server.binaryFlag {
			return respSyntaxError, nil
		}
		return createResponse(200, "Binary flag off."), nil
	},
}

var CDUP = &command{
	name:        []string{"CDUP", "XCUP"},
	demandAuth:  false,
	demandLogin: true,
	demandParam: false,
	cmdFunction: func(conn *Connection, params string) (*response, error) {
		if conn.workDir != conn.authDir {
			ok, err := conn.setLiedDir(filepath.Dir(conn.workDir))
			if err != nil {
				return respProcessError, err
			}
			if ok {
				return createResponse(200, "Okay."), nil
			}
		}
		return createResponse(550, "No further upper path."), nil
	},
}

// MKD 注: 只允许用户逐级创建文件夹
var MKD = &command{
	name:        []string{"MKD", "XMKD"},
	demandAuth:  false,
	demandLogin: true,
	demandParam: true,
	cmdFunction: func(conn *Connection, params string) (*response, error) {
		ps, ok := utils.VerifyParams(params, 1)
		if !ok {
			return respParamsError, nil
		}
		// 处理路径
		newPath := conn.processPath(ps[0])
		if newPath == "" {
			return &response{
				code: 550,
				info: "The path was not exist or no authorization to be processed.",
			}, nil
		}
		if err := os.Mkdir(newPath, os.ModePerm); err != nil {
			return &response{
				code: 550,
				info: "An error occur when the server creating the new file: " + err.Error(),
			}, err
		}
		return createResponse(250, "Directory created."), nil
	},
}

var PWD = &command{
	name:        []string{"PWD", "XPWD"},
	demandAuth:  false,
	demandLogin: true,
	demandParam: false,
	cmdFunction: func(conn *Connection, params string) (*response, error) {
		return &response{
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
	cmdFunction: func(conn *Connection, params string) (*response, error) {
		ps, ok := utils.VerifyParams(params, 1)
		if !ok {
			return respParamsError, nil
		}
		newPath := conn.processPath(ps[0])
		if newPath == "" {
			return createResponse(550, fmt.Sprintf("%s: No such dictionary.", ps[0])), nil
		}
		ok, err := conn.setLiedDir(newPath)
		if err != nil {
			return respProcessError, err
		}
		if !ok {
			return createResponse(550, fmt.Sprintf("%s: No such dictionary.", ps[0])), nil
		}
		return createResponse(250, "Okay."), nil
	},
}

var RMD = &command{
	name:        []string{"RMD", "XRMD"},
	demandAuth:  false,
	demandLogin: true,
	demandParam: true,
	cmdFunction: func(conn *Connection, params string) (*response, error) {
		// 检查参数数量
		ps, ok := utils.VerifyParams(params, 1)
		if !ok {
			return respParamsError, nil
		}
		// 处理路径
		newPath := strings.Replace(conn.processPath(ps[0]), "/", "\\", -1)
		// 不允许用户删除根目录, 也不允许删除用户的工作路径下的目录, 也不允许删除文件
		if newPath == "" || newPath == conn.authDir || newPath == conn.workDir || !utils.IsDir(newPath) {
			return &response{
				code: 550,
				info: "The path was not exist or no authorization to be processed.",
			}, nil
		}
		// 执行删除
		if err := os.Remove(newPath); err != nil {
			return &response{
				code: 550,
				info: "An error occur when the server removing the specify dictionary: " + err.Error(),
			}, err
		}
		return createResponse(250, "Directory removed."), nil
	},
}

var DELE = &command{
	name:        []string{"DELE"},
	demandAuth:  false,
	demandLogin: true,
	demandParam: true,
	cmdFunction: func(conn *Connection, params string) (*response, error) {
		// 检查参数数量
		ps, ok := utils.VerifyParams(params, 1)
		if !ok {
			return respParamsError, nil
		}
		// 处理路径
		newPath := conn.processPath(ps[0])
		if newPath == "" || utils.IsDir(newPath) {
			return createResponse(550, "The file was not exist or no authorization to be processed."), nil
		}
		if err := os.Remove(newPath); err != nil {
			return &response{
				code: 550,
				info: "An error occur when the server removing the specify file: " + err.Error(),
			}, err
		}
		return createResponse(250, "File removed."), nil
	},
}

var PORT = &command{
	name:        []string{"PORT"},
	demandAuth:  false,
	demandLogin: true,
	demandParam: true,
	cmdFunction: func(conn *Connection, params string) (*response, error) {
		ps := strings.SplitN(params, ",", 6)
		addr := utils.ProcessAddr(ps)
		if addr == nil {
			return createResponse(501, "The host port parameter is invalid."), nil
		}
		if err := conn.establishConn(addr); err != nil {
			return createResponse(550, "An error occur when establishing the connection: "+err.Error()), err
		}
		return createResponse(200, "Establishing connection succeed."), nil
	},
}

var SIZE = &command{
	name:        []string{"SIZE"},
	demandAuth:  false,
	demandLogin: true,
	demandParam: true,
	cmdFunction: func(conn *Connection, params string) (*response, error) {
		// 检查参数数量
		ps, ok := utils.VerifyParams(params, 1)
		if !ok {
			return respParamsError, nil
		}
		// 检查文件是否存在以及是否有权限访问
		newPath := conn.processPath(ps[0])
		if newPath == "" {
			return createResponse(550, "The file wasn't exist or no authorization to process."), nil
		}
		fileState, err := os.Stat(newPath)
		if err != nil {
			return createResponse(550, "An error occur when stating the file: "+err.Error()), err
		}
		if fileState.IsDir() {
			return createResponse(550, "The file wasn't exist or no authorization to process."), nil
		}
		return createResponse(213, fmt.Sprintf("%d", fileState.Size())), nil
	},
}

var RNFR = &command{
	name:        []string{"RNFR"},
	demandAuth:  false,
	demandLogin: true,
	demandParam: true,
	cmdFunction: func(conn *Connection, params string) (*response, error) {
		// 检查参数数量
		ps, ok := utils.VerifyParams(params, 1)
		if !ok {
			return respParamsError, nil
		}
		// 检查文件是否有权限访问
		newPath := conn.processPath(ps[0])
		if newPath == "" {
			return createResponse(550, "The path wasn't exist or no authorization to process."), nil
		}
		// 检查文件或文件夹是否存在
		exist, err := utils.VerifyPath(newPath)
		if err != nil {
			return createResponse(550, "An error occur when verifying the file: "+err.Error()), err
		}
		if !exist {
			return createResponse(550, "The path wasn't exist or no authorization to process."), nil
		}
		// 记录原文件或文件夹位置
		conn.renamePath = newPath
		return createResponse(350, "Waiting for next instruction."), nil
	},
}

var RNTO = &command{
	name:        []string{"RNTO"},
	demandAuth:  false,
	demandLogin: true,
	demandParam: true,
	cmdFunction: func(conn *Connection, params string) (*response, error) {
		// 检查参数数量
		ps, ok := utils.VerifyParams(params, 1)
		if !ok {
			return respParamsError, nil
		}
		// 检查文件是否存在以及是否有权限访问
		newPath := conn.processPath(ps[0])
		if newPath == "" {
			return createResponse(550, "The path wasn't exist or no authorization to process."), nil
		}
		// 检查现有路径是否已有同名文件
		exist, err := utils.VerifyPath(newPath)
		if err != nil {
			return createResponse(550, "An error occur when verifying the file: "+err.Error()), err
		}
		if exist {
			return createResponse(553, "An file already exist in this path."), nil
		}
		// 移动文件或文件夹
		if err := os.Rename(conn.renamePath, newPath); err != nil {
			return createResponse(553, "An error occur when renaming the file: "+err.Error()), err
		}
		return createResponse(250, "File renamed."), nil
	},
}

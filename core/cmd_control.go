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

var PWD = &command{
	name:        []string{"PWD", "XPWD"},
	demandAuth:  false,
	demandLogin: true,
	demandParam: false,
	cmdFunction: func(conn *Connection, params string) (*response, error) {
		return &response{
			code: 257,
			info: fmt.Sprintf("\"%s\"", conn.liedDir),
		}, nil
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
		return &response{
			code: 200,
			info: "Binary flag off.",
		}, nil
	},
}

var CDUP = &command{
	name:        []string{"CDUP", "XCUP"},
	demandAuth:  false,
	demandLogin: true,
	demandParam: false,
	cmdFunction: func(conn *Connection, params string) (*response, error) {
		if conn.liedDir != conn.authDir {
			ok, err := conn.setLiedDir(filepath.Dir(conn.liedDir))
			if err != nil {
				return respProcessError, err
			}
			if ok {
				return &response{code: 200, info: "Okay."}, nil
			}
		}
		return &response{code: 550, info: "No further upper path."}, nil
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
		// 检查文件夹名是否正常
		if !utils.VerifyFolderName(ps[0]) {
			return &response{
				code: 550,
				info: "The directory's name was unacceptable for the server's os.",
			}, nil
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
			return &response{
				code: 550,
				info: fmt.Sprintf("%s: No such dictionary.", ps[0]),
			}, nil
		}
		ok, err := conn.setLiedDir(newPath)
		if err != nil {
			return respProcessError, err
		}
		if !ok {
			return &response{
				code: 550,
				info: fmt.Sprintf("%s: No such dictionary.", ps[0]),
			}, nil
		}
		return &response{
			code: 250,
			info: "Okay.",
		}, nil
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
		// 不允许用户删除根目录, 也不允许删除用户的工作路径下的目录
		if newPath == "" || newPath == conn.authDir || newPath == conn.liedDir {
			return &response{
				code: 550,
				info: "The path was not exist or no authorization to be processed.",
			}, nil
		}

		// 执行删除
		if err := os.Remove(newPath); err != nil {
			return &response{
				code: 550,
				info: "An error occur when the server removing the specify file: " + err.Error(),
			}, err
		}
		return createResponse(250, "Directory removed."), nil
	},
}

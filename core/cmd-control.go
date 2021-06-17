package core

import (
	"github.com/Aldrice/liteFTP/common/utils"
)

var OPTS = &command{
	name:        "OPTS",
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
				info: "server do not allow utf_8 encoding transmission",
			}, nil
		}
		return &response{
			code: 200,
			info: "server are now transmit with utf_8 encoding",
		}, nil
	},
}

var PASV = &command{
	name:        "PASV",
	demandAuth:  false,
	demandLogin: false,
	demandParam: true,
	cmdFunction: func(conn *Connection, params string) (*response, error) {
		return respParamsError, nil
	},
}

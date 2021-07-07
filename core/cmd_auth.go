package core

import (
	"github.com/Aldrice/liteFTP/common/config"
	"strings"
)

var USER = &command{
	name:        []string{"USER"},
	demandAuth:  false,
	demandLogin: false,
	demandParam: true,
	cmdFunction: func(conn *Connection, params string) (*rsp, error) {
		// 检查登录状态, 若已经登录则拒绝执行该指令
		if resp := conn.verifyLogin(); resp != nil {
			return resp, nil
		}
		username := strings.ToLower(params)
		// 检查是否允许匿名访问
		if username == config.Anonymous {
			if !conn.server.enableAnonymous {
				return newResponse(530, "Server do not allow anonymous user."), nil
			}
		} else {
			// 若用户不为匿名用户, 则检查该用户是否存在
			exist, err := conn.server.srvDB.VerifyUser(username)
			if err != nil {
				return newResponse(530, rspDataBaseError, err.Error()), err
			}
			if !exist {
				return newResponse(
					530,
					"The user is not exist in this FTP server.\nUse SITE PSWD -password to register a new user as this name.",
				), nil
			}
		}
		conn.temp = username
		return newResponse(331, rspTextTempReceived), nil
	},
}

var PASS = &command{
	name:        []string{"PASS"},
	demandAuth:  false,
	demandLogin: false,
	demandParam: true,
	cmdFunction: func(conn *Connection, params string) (*rsp, error) {
		// 检查登录状态, 已登录用户不允许使用该指令
		if rsp := conn.verifyLogin(); rsp != nil {
			return rsp, nil
		}
		// 检查temp是否有值, 若用户在输入USER之前便输入该指令, 则也报错
		// 检查该用户是匿名用户还是非匿名用户
		// 实现登录状态下的相关处理
		switch conn.temp {
		case "":
			return newResponse(502, "You need to input your username before using this command."), nil
		case config.Anonymous:
			if !conn.server.enableAnonymous {
				return newResponse(530, "Server do not allow anonymous user."), nil
			}
		default:
			exist, err := conn.server.srvDB.VerifyUser(conn.temp, params)
			if err != nil {
				return newResponse(530, rspDataBaseError, err.Error()), err
			}
			if !exist {
				return newResponse(530, "The password is not match with this user."), nil
			}
			isAdmin, err := conn.server.srvDB.IsAdmin(conn.temp)
			if err != nil {
				return newResponse(530, rspDataBaseError, err.Error()), err
			}
			if isAdmin {
				conn.isAdmin = true
			}
		}
		conn.isLogin = true
		conn.refreshDir()
		return newResponse(200, rspTextLoginSuccess), nil
	},
}

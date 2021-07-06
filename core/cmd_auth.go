package core

import (
	"github.com/Aldrice/liteFTP/common/config"
	"path/filepath"
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
				return createResponse(530, "Server do not allow anonymous user."), nil
			}
		} else {
			// 若用户不为匿名用户, 则检查该用户是否存在
			exist, err := conn.server.srvDB.VerifyUser(username)
			if err != nil {
				return createResponse(530, "An error occur when processing in the database.", err.Error()), err
			}
			// todo: 可能要给予用户渠道去注册账户
			if !exist {
				return createResponse(530, "The user is not exist in this FTP server."), nil
			}
		}
		conn.temp = username
		return rspTempReceived, nil
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
			return createResponse(502, "You need to input your username before using this command."), nil
		case config.Anonymous:
			if !conn.server.enableAnonymous {
				return createResponse(530, "Server do not allow anonymous user."), nil
			}
		default:
			exist, err := conn.server.srvDB.VerifyUser(conn.temp, strings.ToLower(params))
			if err != nil {
				return createResponse(530, "An error occur when processing in the database.", err.Error()), err
			}
			// todo: 可能要给予用户渠道去注册账户
			if !exist {
				return createResponse(530, "The password is not match with this user."), nil
			}
		}
		// todo: 用户可能有多处分叉的最大根目录 (如 user/anonymous, user/xxx)
		conn.isLogin = true
		conn.authDir = filepath.Join(conn.server.userDir, conn.temp)
		conn.workDir = conn.authDir
		return rspLoginSuccess, nil
	},
}

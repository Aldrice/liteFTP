package core

import (
	"path/filepath"
	"strings"
)

var USER = &command{
	name:        []string{"USER"},
	demandAuth:  false,
	demandLogin: false,
	demandParam: true,
	cmdFunction: func(conn *Connection, params string) (*response, error) {
		// 检查登录状态
		if resp := conn.verifyLogin(); resp != nil {
			return resp, nil
		}
		username := strings.ToLower(params)
		// 检查是否允许匿名访问
		if username == anonymous {
			if !conn.server.enableAnonymous {
				return &response{
					code: 530,
					info: "Server do not allow anonymous user.",
				}, nil
			}
			conn.isAnonymous = true
		}
		conn.temp = params
		return respTempReceived, nil
	},
}

var PASS = &command{
	name:        []string{"PASS"},
	demandAuth:  false,
	demandLogin: false,
	demandParam: true,
	cmdFunction: func(conn *Connection, params string) (*response, error) {
		// 检查登录状态
		if resp := conn.verifyLogin(); resp != nil {
			return resp, nil
		}
		if conn.isAnonymous {
			conn.isLogin = true
			conn.authDir = filepath.Join(conn.server.rootDir, conn.temp)
			conn.liedDir = conn.authDir
			return respLoginSuccess, nil
		}
		// todo: 实现登录状态下的相关处理
		// todo: 明确该用户有权限的最大根目录

		return respParamsError, nil
	},
}

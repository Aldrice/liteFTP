package core

import (
	"github.com/Aldrice/liteFTP/common/config"
	"os"
	"path/filepath"
	"strings"
)

var SITE = &command{
	name:        []string{"SITE"},
	demandAuth:  false,
	demandLogin: false,
	demandParam: true,
	cmdFunction: func(conn *Connection, params string) (*rsp, error) {
		ps := strings.Split(params, " ")
		// 实现用户的注册, 改密, 发站内信, 帮助
		switch strings.ToUpper(ps[0]) {
		case "HELP":
			return conn.siteHELP()
		case "PSWD":
			return conn.sitePSWD(ps[1:])
		case "REGT":
			return conn.siteREGT(ps[1:])
		case "MSGE":
			return conn.siteMSGE(ps[1:])
		default:
			return conn.siteHELP()
		}
	},
}

// siteHELP 用户请求SITE指令列表
func (conn *Connection) siteHELP() (*rsp, error) {
	info :=
		"\r\nSITE PSWD -PASSWORD , USED TO CHANGE YOUR PASSWORD\r\n" +
			"SITE REGT -USERNAME -PASSWORD , USED TO REGISTER YOUR ACCOUNT WITH PASSWORD\r\n" +
			"SITE MSGE -MESSAGE  , SEND MESSAGE TO SERVER MANAGER"
	return newResponse(220, info), nil
}

// sitePSWD 用户自助更改密码服务
func (conn *Connection) sitePSWD(params []string) (*rsp, error) {
	// 只有当用户为非匿名用户并且已登录的情况下才可以使用该指令
	if !(conn.temp != config.Anonymous && conn.isLogin) {
		return newResponse(550, "No authorization to use this command."), nil
	}
	if err := conn.server.srvDB.ChangePassword(conn.temp, params[0]); err != nil {
		return newResponse(550, rspDataBaseError), err
	}
	return newResponse(220, "Change password Success."), nil
}

// siteREGT 用户自助注册服务
func (conn *Connection) siteREGT(params []string) (*rsp, error) {
	// 只有当用户在未登录的情况下才可以使用该指令
	if conn.isLogin {
		return newResponse(550, "No authorization to use this command."), nil
	}
	// 检查该用户名是否已被注册
	exist, err := conn.server.srvDB.VerifyUser(params[0])
	if err != nil {
		return newResponse(550, rspDataBaseError), err
	}
	// 当该用户已经被注册时, 报错
	if exist {
		return newResponse(550, "This user already exist."), err
	}
	// 创建该用户
	if err := conn.server.srvDB.CreateUser(params[0], params[1]); err != nil {
		return newResponse(550, rspDataBaseError), err
	}
	username := strings.ToLower(params[0])
	// 在此处新建文件夹
	if err := os.Mkdir(filepath.Join(conn.server.userDir, username), os.ModePerm); err != nil {
		return newResponse(550, rspProcessError), err
	}
	// 让用户处于登录状态
	conn.isLogin = true
	conn.temp = username
	conn.refreshDir()
	return newResponse(200, rspTextLoginSuccess), nil
}

// siteMSGE 用户发送站内信
func (conn *Connection) siteMSGE(params []string) (*rsp, error) {

	return nil, nil
}

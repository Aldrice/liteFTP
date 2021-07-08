package core

import (
	"fmt"
	"github.com/Aldrice/liteFTP/common/config"
	"github.com/Aldrice/liteFTP/common/datebase"
	"github.com/Aldrice/liteFTP/utils"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var SITE = &command{
	name:        []string{"SITE"},
	demandAuth:  false,
	demandLogin: false,
	demandParam: true,
	cmdFunction: func(conn *Connection, params string) (*rsp, error) {
		components := strings.SplitN(params, " ", 2)
		cmd, ok := conn.server.udmCommand[strings.ToUpper(components[0])]
		// 若无匹配的指令, 默认返回HELP指令结果
		if !ok {
			return siteHELP.cmdFunction(conn, "")
		}
		if cmd.demandLogin && !conn.isLogin {
			return newResponse(550, rspTextAuthError), nil
		}
		if cmd.demandAuth && (conn.temp == config.Anonymous) {
			return newResponse(550, rspTextAuthError), nil
		}
		if cmd.demandAdmin && !conn.isAdmin {
			return newResponse(550, rspTextAuthError), nil
		}
		if cmd.demandParam {
			if len(components) == 1 {
				return rspParamsError, nil
			}
			return cmd.cmdFunction(conn, components[1])
		} else {
			return cmd.cmdFunction(conn, "")
		}
	},
}

var siteHELP = &command{
	name:        []string{"HELP"},
	demandAuth:  false,
	demandLogin: false,
	demandParam: false,
	demandAdmin: false,
	cmdFunction: func(conn *Connection, params string) (*rsp, error) {
		info :=
			"\r\nSITE PSWD [PASSWORD] , CHANGE YOUR PASSWORD\r\n" +
				"SITE REGT [USERNAME] [PASSWORD] , REGISTER YOUR ACCOUNT WITH USERNAME & PASSWORD\r\n" +
				"SITE MSGE [MESSAGE]  , SEND MESSAGE TO SERVER MANAGER\r\n" +
				"SITE READ [MESSAGE ID] , READ MESSAGE BY IT'S ID (ONLY FOR ADMIN)\r\n" +
				"SITE LIST [TYPE] [PAGE NUM]\r\n" +
				"LIST MESSAGES BY THE TYPE(1 or 0) OF MESSAGE AND PAGE NUMBERS (ONLY FOR ADMIN)\r\n" +
				"TYPE: 1, READ MESSAGE; 0, UNREAD MESSAGE\r\n" +
				"SITE USER [PAGE NUM]\r\n" +
				"LIST USERS BY THE PAGE NUMBERS (ONLY FOR ADMIN)"
		return newResponse(220, info), nil
	},
}

var sitePSWD = &command{
	name:        []string{"PSWD"},
	demandAuth:  true,
	demandLogin: true,
	demandParam: true,
	demandAdmin: false,
	cmdFunction: func(conn *Connection, params string) (*rsp, error) {
		// 检查密码是否为有效值
		if !utils.StringsIsNotNull(params) {
			return rspParamsError, nil
		}
		if err := conn.server.srvDB.ChangePassword(conn.temp, params); err != nil {
			return newResponse(550, rspTextDataBaseError, err.Error()), nil
		}
		return newResponse(220, "Change password Success."), nil
	},
}

// siteREGT 用户自助注册服务
var siteREGT = &command{
	name:        []string{"REGT"},
	demandAuth:  false,
	demandLogin: false,
	demandParam: true,
	demandAdmin: false,
	cmdFunction: func(conn *Connection, params string) (*rsp, error) {
		// 检查用户输入的值是否有username和password, 且其是否为有效值
		ps := strings.SplitN(params, " ", 2)
		for i, p := range ps {
			ps[i] = strings.ToLower(strings.TrimSpace(p))
			if ps[i] == "" {
				return rspParamsError, nil
			}
		}
		// 只有当用户在未登录的情况下才可以使用该指令
		if conn.isLogin {
			return newResponse(550, "No authorization to use this command."), nil
		}
		// 检查该用户名是否已被注册
		exist, err := conn.server.srvDB.VerifyUser(ps[0])
		if err != nil {
			return newResponse(550, rspTextDataBaseError, err.Error()), nil
		}
		// 当该用户已经被注册时, 报错
		if exist {
			return newResponse(550, "This user already exist."), nil
		}
		// 创建该用户
		if err := conn.server.srvDB.CreateUser(ps[0], ps[1]); err != nil {
			return newResponse(550, rspTextDataBaseError, err.Error()), nil
		}
		// 在此处新建文件夹
		if err := os.Mkdir(filepath.Join(conn.server.userDir, ps[0]), os.ModePerm); err != nil {
			return newResponse(550, rspTextProcessError, err.Error()), nil
		}
		// 让用户处于登录状态
		conn.isLogin = true
		conn.temp = ps[0]
		conn.refreshDir()
		return newResponse(200, rspTextLoginSuccess), nil
	},
}

// siteMSGE 用户发送站内信, 单向发送, 只允许用户发给管理员
var siteMSGE = &command{
	name:        []string{"MSGE"},
	demandAuth:  true,
	demandLogin: true,
	demandParam: true,
	demandAdmin: false,
	cmdFunction: func(conn *Connection, params string) (*rsp, error) {
		if !utils.StringsIsNotNull(params) {
			return rspParamsError, nil
		}
		// 管理员不可发站内信给自己
		if conn.isAdmin {
			return newResponse(550, "You can't send message to yourself."), nil
		}
		if err := conn.server.srvDB.CreateMessage(conn.temp, params); err != nil {
			return newResponse(550, rspTextDataBaseError, err.Error()), nil
		}
		return newResponse(220, "Send completed."), nil
	},
}

// siteREAD 管理员根据 Message ID 去读取具体的站内信内容
var siteREAD = &command{
	name:        []string{"READ"},
	demandAuth:  true,
	demandLogin: true,
	demandParam: true,
	demandAdmin: true,
	cmdFunction: func(conn *Connection, params string) (*rsp, error) {
		id, err := strconv.Atoi(strings.TrimSpace(params))
		if err != nil {
			return rspParamsError, err
		}
		m, err := conn.server.srvDB.GetMessage(id)
		if err != nil {
			return newResponse(550, rspTextDataBaseError, err.Error()), nil
		}
		return newResponse(220, m.FormatToEntity()), nil
	},
}

// siteLIST 管理员 已读或未读 去列出所有的符合条件的站内信
var siteLIST = &command{
	name:        []string{"LIST"},
	demandAuth:  true,
	demandLogin: true,
	demandParam: true,
	demandAdmin: true,
	cmdFunction: func(conn *Connection, params string) (*rsp, error) {
		ps, ok := utils.VerifyParams(params, 2)
		if !ok {
			return rspParamsError, nil
		}
		mTypeInt, err := strconv.Atoi(ps[0])
		if err != nil {
			return rspParamsError, err
		}
		page, err := strconv.Atoi(ps[1])
		if err != nil {
			return rspParamsError, err
		}
		mType := true
		if mTypeInt == 0 {
			mType = false
		}
		cnt, ms, err := conn.server.srvDB.ListMessages(page, mType)
		if err != nil {
			return newResponse(550, rspTextDataBaseError, err.Error()), nil
		}
		minPage, maxPage := utils.CalPageExtreme(config.PageSize, cnt)
		text := datebase.FormatMessages(ms) + fmt.Sprintf("\r\n[%d]<--(%d)-->[%d]", minPage, page, maxPage)
		return newResponse(220, text), nil
	},
}

// siteUSER 管理员 列出用户列表
var siteUSER = &command{
	name:        []string{"USER"},
	demandAuth:  true,
	demandLogin: true,
	demandParam: true,
	demandAdmin: true,
	cmdFunction: func(conn *Connection, params string) (*rsp, error) {
		ps, ok := utils.VerifyParams(params, 1)
		if !ok {
			return rspParamsError, nil
		}
		page, err := strconv.Atoi(ps[0])
		if err != nil {
			return rspParamsError, err
		}
		cnt, us, err := conn.server.srvDB.ListUsers(page)
		if err != nil {
			return newResponse(550, rspTextDataBaseError, err.Error()), nil
		}
		minPage, maxPage := utils.CalPageExtreme(config.PageSize, cnt)
		text := datebase.FormatUsers(us) + fmt.Sprintf("\r\n[%d]<--(%d)-->[%d]", minPage, page, maxPage)
		return newResponse(220, text), nil
	},
}

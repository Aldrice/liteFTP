package core

// todo: 成员变量 name 可删除
type command struct {
	name        string
	demandAuth  bool
	demandLogin bool
	demandParam bool
	cmdFunction func(conn *Connection, params string) (*response, error)
}

func loadAllCommands() map[string]*command {
	return map[string]*command{
		// auth
		USER.name: USER, // USER 认证用户名
		PASS.name: PASS, // PASS 认证密码
		// control
		QUIT.name: QUIT,
		OPTS.name: OPTS,
		PASV.name: PASV, // PASV 进入被动模式
		// transmit
		MKD.name: MKD, // MKD 创建目录
		// todo: XMKD
		"XMKD": MKD, // MKD 创建目录
	}
}

// PORT 指定服务器要连接的地址和端口
// QUIT 断开连接
// CDUP 改变到父目录
// DELE 删除文件
// LIST 如果指定了命令，返回命令使用文档；否则返回一个通用帮助文档

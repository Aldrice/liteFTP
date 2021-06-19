package core

type commandList map[string]*command

type command struct {
	name        []string
	demandAuth  bool
	demandLogin bool
	demandParam bool
	cmdFunction func(conn *Connection, params string) (*response, error)
}

// todo: 仍需要优化, 指令集合 -> 载入指令集合
func loadAllCommands() commandList {
	cmdList := make(commandList)
	// auth
	{
		cmdList.loadCommand(USER)
		cmdList.loadCommand(PASS)
	}
	// control
	{
		cmdList.loadCommand(OPTS)
		cmdList.loadCommand(PASV)
		cmdList.loadCommand(QUIT)
		cmdList.loadCommand(PWD)
		cmdList.loadCommand(SYST)
		cmdList.loadCommand(NOOP)
		cmdList.loadCommand(TYPE)
		cmdList.loadCommand(CDUP)
		cmdList.loadCommand(RMD)
		cmdList.loadCommand(MKD)
		cmdList.loadCommand(CWD)
		cmdList.loadCommand(DELE)
	}
	// transmit
	{
		cmdList.loadCommand(STOR)
	}
	return cmdList
}

func (cmd *commandList) loadCommand(c *command) {
	for _, n := range c.name {
		(*cmd)[n] = c
	}
}

// PORT 指定服务器要连接的地址和端口
// QUIT 断开连接
// CDUP 改变到父目录
// DELE 删除文件
// LIST 如果指定了命令，返回命令使用文档；否则返回一个通用帮助文档

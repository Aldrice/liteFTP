package core

type command struct {
	name        []string
	demandAuth  bool
	demandLogin bool
	demandParam bool
	cmdFunction func(conn *Connection, params string) (*rsp, error)
}

type cmdMap map[string]*command

// todo: 可以选用标准指令集
var cmdList = []*command{
	// auth
	USER,
	PASS,
	// control
	OPTS,
	PASV,
	FEAT,
	QUIT,
	PWD,
	SYST,
	NOOP,
	TYPE,
	CDUP,
	RMD,
	MKD,
	CWD,
	DELE,
	PORT,
	SIZE,
	RNFR,
	RNTO,
	// transmit
	STOR,
	LIST,
	RETR,
}

func loadAllCommands(commands []*command) cmdMap {
	commandMap := make(cmdMap)
	for _, cmd := range commands {
		commandMap.loadCommand(cmd)
	}
	return commandMap
}

func (cmd *cmdMap) loadCommand(c *command) {
	for _, n := range c.name {
		(*cmd)[n] = c
	}
}

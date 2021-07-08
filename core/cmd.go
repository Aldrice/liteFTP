package core

type command struct {
	name        []string
	demandAuth  bool
	demandLogin bool
	demandParam bool
	demandAdmin bool
	cmdFunction func(conn *Connection, params string) (*rsp, error)
}

type cmdMap map[string]*command

//  标准指令集
var stdCMD = []*command{
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
	// extend
	SITE,
}

// 服务器自定义指令集
var udmCMD = []*command{
	// all
	siteHELP,
	siteREGT,
	// user
	sitePSWD,
	siteMSGE,
	// admin
	siteREAD,
	siteLIST,
	siteUSER,
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

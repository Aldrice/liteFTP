package core

import (
	"bufio"
	"fmt"
	"github.com/Aldrice/liteFTP/common/utils"
	"log"
	"net"
	"path/filepath"
	"strings"
)

// 连接的ID号
var cid = 0

type Connection struct {
	connectID int
	server    *server

	linkConn *net.TCPConn
	wt       *bufio.Writer
	rt       *bufio.Reader

	dataConn *net.TCPConn

	// todo: 该用户的信息, 权限, IP地址, 是否开启被动模式等等
	isLogin     bool
	isAnonymous bool
	temp        string
	authDir     string // 有权限的最大根目录
	liedDir     string // 目前所在的目录
}

func (conn *Connection) handle() {
	defer func() {
		log.Print(fmt.Sprintf("断开一个连接, 连接ID: %d", conn.connectID))
		_ = conn.linkConn.Close()
	}()

	if err := conn.sendText(respWelcome); err != nil {
		log.Print("出错了")
	}

	for {
		// 读指令
		res, err := conn.readCommand()
		if err != nil && res == nil {
			_ = conn.linkConn.Close()
			return
		}
		log.Printf("response: %d %s", res.code, res.info)
		if err := conn.sendText(res); err != nil {
			_ = conn.linkConn.Close()
			return
		}
	}
}

func (conn *Connection) sendText(r *response) error {
	_, err := conn.wt.WriteString(fmt.Sprintf("%d %s\r\n", r.code, r.info))
	if err != nil {
		return err
	}
	return conn.wt.Flush()
}

func (conn *Connection) readCommand() (*response, error) {
	statement, err := conn.rt.ReadString('\n')
	if err != nil {
		// tcp连接中断
		return nil, err
	}
	statement = strings.TrimRight(statement, "\r\n")
	components := strings.SplitN(statement, " ", 2)
	log.Println("request: " + statement)
	// 指令匹配
	c, exist := conn.server.command[strings.ToUpper(components[0])]
	if !exist {
		return respSyntaxError, nil
	}
	// 检查是否需要权限
	if c.demandAuth && (!conn.isLogin || conn.isAnonymous) {
		return respAuthError, nil
	}
	// 检查输入参数数量是否符合要求
	if c.demandParam {
		if len(components) == 1 {
			return respParamsError, nil
		}
		return c.cmdFunction(conn, components[1])
	}
	return c.cmdFunction(conn, "")
}

// verifyLogin 检查用户登录状态，若已经登录，则返回错误信息
func (conn *Connection) verifyLogin() *response {
	if conn.isLogin {
		return &response{
			code: 502,
			info: "user already login",
		}
	}
	return nil
}

// processPath 返回有效路径, 若返回的路径为空, 说明输入路径有误
func (conn *Connection) processPath(path string) string {
	// todo: 处理windows file explorer路径格式问题
	if filepath.IsAbs(path) {
		relPath, err := filepath.Rel(conn.authDir, path)
		if err != nil || strings.Contains(relPath, "..") {
			return ""
		}
		return path
	}
	return filepath.Join(conn.liedDir, path)
}

// setLiedDir 设置当前连接的工作路径, 若路径不存在则报错
func (conn *Connection) setLiedDir(path string) (bool, error) {
	// 检查该路径是否存在
	exist, err := utils.VerifyPath(path)
	if err != nil {
		return false, err
	}
	if !exist {
		return false, nil
	}
	conn.liedDir = path
	return true, nil
}

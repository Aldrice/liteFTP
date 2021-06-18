package core

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
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
		if err != nil {
			// 如果连接已经断开, 则直接关闭连接
			if errors.Is(err, net.ErrClosed) {
				break
			}
			log.Print("指令处理出错", err.Error())
		}
		if err := conn.sendText(res); err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}
			log.Print("结果回送出错", err.Error())
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
		// 用户输入语句有语法错误
		if errors.Is(err, io.EOF) {
			return respSyntaxError, err
		}
		// 执行出现错误
		return respProcessError, err
	}
	statement = strings.TrimRight(statement, "\r\n")
	components := strings.SplitN(statement, " ", 2)
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

func (conn *Connection) processPath(path string) string {
	// todo: 此处有错误要处理

	return ""
}

/*// processPath 处理用户的路径参数
func (conn *Connection) processPath(path string) string {
	// todo: 处理windows路径

	// todo: 仍然有问题
	// 当路径为绝对路径时
	newPath := path
	if !filepath.IsAbs(path) {
		newPath = filepath.Join(conn.authDir, path)
	}
	rel, err := filepath.Rel(conn.authDir, newPath)
	if err != nil || strings.Contains(rel, "..") {
		return conn.authDir
	}
	return newPath
}*/

package core

import (
	"bufio"
	"context"
	"fmt"
	"github.com/Aldrice/liteFTP/common/utils"
	"io"
	"log"
	"net"
	"path/filepath"
	"strings"
	"time"
)

// 连接的ID号
var cid = 0

type Connection struct {
	connectID int
	server    *server

	linkConn *net.TCPConn
	wt       *bufio.Writer
	rd       *bufio.Reader

	dataConn *net.TCPConn

	isPassive bool

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

// todo: 被动模式下连接的建立
func (conn *Connection) establishConn(addr *net.TCPAddr) error {
	var tcpConn *net.TCPConn
	var err error
	if conn.isPassive {

	} else {
		tcpConn, err = net.DialTCP("tcp", nil, addr)
		if err != nil {
			return err
		}
	}
	// todo: 完善连接的关闭
	conn.dataConn = tcpConn
	return nil
}

func (conn *Connection) sendText(r *response) error {
	_, err := conn.wt.WriteString(fmt.Sprintf("%d %s\r\n", r.code, r.info))
	if err != nil {
		return err
	}
	return conn.wt.Flush()
}

func (conn *Connection) readCommand() (*response, error) {
	statement, err := conn.rd.ReadString('\n')
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

func (conn *Connection) readData(ctx context.Context, wt io.Writer) (*response, error) {
	// todo: 开始前检查连接状况
	if err := conn.sendText(createResponse(125, "Starting a data transport")); err != nil {
		return createResponse(1, "An error occur when sending text to client: "+err.Error()), err
	}
	// 等待连接
	for {
		// 被动情况下的处理
		if conn.dataConn != nil {
			break
		}
		if err := ctx.Err(); err != nil {
			return createResponse(550, "Waiting transport time out."), nil
		}
		time.Sleep(time.Millisecond * 100)
		log.Print("Waiting transport...")
	}
	// 接收数据
	size, err := io.Copy(wt, conn.dataConn)
	if err != nil {
		return createResponse(451, "An error occur when receiving the data: "+err.Error()), err
	}
	return createResponse(226, fmt.Sprintf("Receive complete, data size: %d", size)), nil
}

func (conn *Connection) writeData(ctx context.Context, rd io.Reader, msg string) (*response, error) {
	if err := conn.sendText(createResponse(125, msg)); err != nil {
		return createResponse(1, "An error occur when sending text to client: "+err.Error()), err
	}
	// 等待连接
	for {
		// 被动情况下的处理
		if conn.dataConn != nil {
			break
		}
		if err := ctx.Err(); err != nil {
			return createResponse(550, "Waiting transport time out."), nil
		}
		time.Sleep(time.Millisecond * 100)
		log.Print("Waiting transport...")
	}
	// todo: 连接断开的处理
	defer conn.dataConn.Close()
	buf := make([]byte, 1024*1024)
	size, err := io.CopyBuffer(conn.dataConn, rd, buf)
	if err != nil {
		return createResponse(451, "An error occur when sending the data: "+err.Error()), nil
	}
	return createResponse(226, fmt.Sprintf("Send complete, data size: %d", size)), nil
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
	// 路径为绝对路径的情况
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

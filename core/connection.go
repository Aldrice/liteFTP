package core

import (
	"bufio"
	"context"
	"fmt"
	"github.com/Aldrice/liteFTP/common/config"
	"github.com/Aldrice/liteFTP/utils"
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
	// 连接信息
	connectID int
	server    *server

	linkConn *net.TCPConn
	wt       *bufio.Writer
	rd       *bufio.Reader

	dataConn *net.TCPConn
	pasvAddr *net.TCPAddr

	isPassive  bool   // 是否处于被动模式
	renamePath string // 要移动的文件路径
	authDir    string // 有权限的最大根目录
	workDir    string // 目前所在的目录

	// 用户信息
	temp    string // 用户昵称
	isLogin bool   // 是否处于登录状态
	isAdmin bool   // 是否为管理员
}

func (conn *Connection) handle() {
	defer func() {
		log.Print(fmt.Sprintf("断开一个连接, 连接ID: %d", conn.connectID))
		_ = conn.linkConn.Close()
	}()

	if err := conn.sendText(newResponse(220, rspTextWelcome)); err != nil {
		log.Print("出错了")
	}

	for {
		// 读指令
		res, err := conn.readCommand()
		if err != nil && res == nil {
			_ = conn.linkConn.Close()
			return
		}
		log.Printf("rsp: %d %s", res.code, res.info)
		if err := conn.sendText(res); err != nil {
			_ = conn.linkConn.Close()
			return
		}
	}
}

// establishConn 建立数据连接
func (conn *Connection) establishConn(addr *net.TCPAddr) error {
	var tcpConn *net.TCPConn
	var err error
	if conn.isPassive {
		for {
			pasvAddr := &net.TCPAddr{IP: addr.IP, Port: conn.server.getPassivePort()}
			tcpLsn, err := net.ListenTCP("tcp", pasvAddr)
			if err == nil {
				conn.pasvAddr = pasvAddr
				go func() {
					// 限定监听时间为一分钟
					err = tcpLsn.SetDeadline(time.Now().Add(time.Minute))
					if err != nil {
						return
					}
					// 开始监听
					log.Printf("正在监听一个连接 - 连接端口: %d", pasvAddr.Port)
					conn.dataConn, err = tcpLsn.AcceptTCP()
					if err != nil {
						log.Printf("建立一个连接失败 - 连接端口: %d, 错误信息: %s", pasvAddr.Port, err.Error())
					}
					log.Printf("成功建立一个连接 - 连接端口: %d", pasvAddr.Port)

					// 关闭监听, 限定数据连接时间为一分钟
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()

					_ = conn.dataConn.SetDeadline(time.Now().Add(time.Minute))
					_ = tcpLsn.Close()
					log.Printf("断开一个连接的监听 - 连接端口: %d", pasvAddr.Port)

					<-ctx.Done()

					_ = conn.dataConn.Close()
					log.Printf("断开一个连接 - 连接端口: %d", pasvAddr.Port)
					conn.dataConn = nil
					conn.pasvAddr = nil
				}()
			}
			if !utils.IsPortInuse(err) {
				return err
			}
		}
	} else {
		tcpConn, err = net.DialTCP("tcp", nil, addr)
		if err != nil {
			return err
		}
	}
	conn.dataConn = tcpConn
	return nil
}

func (conn *Connection) sendText(r *rsp) error {
	_, err := conn.wt.WriteString(r.formatResponse())
	if err != nil {
		return err
	}
	return conn.wt.Flush()
}

func (conn *Connection) readCommand() (*rsp, error) {
	statement, err := conn.rd.ReadString('\n')
	if err != nil {
		// tcp连接中断
		return nil, err
	}
	statement = strings.TrimRight(statement, "\r\n")
	log.Println("request: " + statement)
	components := strings.SplitN(statement, " ", 2)
	// 指令匹配
	c, exist := conn.server.command[strings.ToUpper(components[0])]
	if !exist {
		return rspSyntaxError, nil
	}
	// 检查是否需要权限
	if c.demandAuth && (!conn.isLogin || conn.temp == config.Anonymous) {
		return newResponse(530, "User no auth to execute this command."), nil
	}
	// 检查输入参数数量是否符合要求
	if c.demandParam {
		if len(components) == 1 {
			return rspParamsError, nil
		}
		return c.cmdFunction(conn, components[1])
	}
	return c.cmdFunction(conn, "")
}

func (conn *Connection) readData(ctx context.Context, wt io.Writer) (*rsp, error) {
	if err := conn.sendText(newResponse(125, "Starting a data transport")); err != nil {
		return newResponse(550, rspTextSendError, err.Error()), err
	}
	// 等待连接
	for {
		// 被动情况下的处理
		if conn.dataConn != nil {
			break
		}
		if err := ctx.Err(); err != nil {
			return newResponse(550, "Waiting transport time out."), nil
		}
		time.Sleep(time.Millisecond * 100)
		log.Print("Waiting transport...")
	}
	// 接收数据
	size, err := io.Copy(wt, conn.dataConn)
	if err != nil {
		return newResponse(451, "An error occur when receiving the data", err.Error()), err
	}
	return newResponse(226, fmt.Sprintf("Receive complete, data size: %d", size)), nil
}

func (conn *Connection) writeData(ctx context.Context, rd io.Reader) (*rsp, error) {
	// 等待连接
	for {
		// 被动情况下的处理, 若数据链路已建立则跳出等待循环
		if conn.dataConn != nil {
			break
		}
		// 若超时则跳出等待循环并回复用户端错误
		if err := ctx.Err(); err != nil {
			return newResponse(550, "Waiting transport time out."), nil
		}
		time.Sleep(time.Millisecond * 100)
		log.Print("Waiting transport...")
	}
	// 连接断开的处理
	defer func() {
		if conn.dataConn != nil {
			_ = conn.dataConn.Close()
			log.Printf("断开一个连接 - 连接: %s", conn.dataConn.LocalAddr().String())
		}
		conn.dataConn = nil
	}()
	// 限定传输数据的缓存为1024KB
	buf := make([]byte, 1024*1024)
	size, err := io.CopyBuffer(conn.dataConn, rd, buf)
	if err != nil {
		return newResponse(451, "An error occur when sending the data", err.Error()), nil
	}
	return newResponse(226, fmt.Sprintf("Send complete, data size: %d", size)), nil
}

// verifyLogin 检查用户登录状态，若已经登录，则返回错误信息
func (conn *Connection) verifyLogin() *rsp {
	if conn.isLogin {
		return newResponse(502, "User already login.")
	}
	return nil
}

// processPath 返回有效路径, 若返回的路径为空, 说明输入路径有误
func (conn *Connection) processPath(path string) string {
	// 处理windows file explorer路径格式问题
	path = strings.Trim(path, "/")

	// 路径为绝对路径的情况
	if filepath.IsAbs(path) {
		relPath, err := filepath.Rel(conn.authDir, path)
		if err != nil || strings.Contains(relPath, "..") {
			return ""
		}
		return path
	}
	return filepath.Join(conn.workDir, path)
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
	conn.workDir = path
	return true, nil
}

// refreshDir 刷新当前连接关于路径的信息
func (conn *Connection) refreshDir() {
	if conn.isAdmin {
		conn.authDir = conn.server.userDir
	} else {
		conn.authDir = filepath.Join(conn.server.userDir, conn.temp)
	}
	conn.workDir = conn.authDir
}

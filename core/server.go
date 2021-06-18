package core

import (
	"bufio"
	"fmt"
	. "github.com/Aldrice/liteFTP/common/config"
	"github.com/Aldrice/liteFTP/common/utils"
	"log"
	"net"
	"os"
	"path/filepath"
)

const anonymous = "anonymous"

type server struct {
	// 服务器支持的指令集
	command commandList
	// 被动模式下最大端口
	pasvMaxPort int
	// 被动模式下最小端口
	pasvMinPort int
	// 是否启用UTF8编码通信
	enableUTF8 bool
	// 是否允许匿名访问
	enableAnonymous bool
	// 服务器文件存储基址
	rootDir string
	// 服务器是否允许二进制传输
	binaryFlag bool
	// todo: 权限管理
}

func NewServer() *server {
	absPath, err := filepath.Abs(InitCfg.SrvCfg.RootDir)
	if err != nil {
		log.Fatal("路径转换出错")
	}
	s := &server{
		command:         loadAllCommands(),
		pasvMaxPort:     InitCfg.PortCfg.MaxPasvPort,
		pasvMinPort:     InitCfg.PortCfg.MinPasvPort,
		enableUTF8:      InitCfg.SrvCfg.EnableUTF8,
		enableAnonymous: InitCfg.SrvCfg.EnableAnonymous,
		rootDir:         absPath,
		binaryFlag:      InitCfg.SrvCfg.BinaryFlag,
	}

	// 检查服务单元的存储路径是否有效
	exist, err := utils.VerifyPath(s.rootDir)
	if err != nil {
		log.Fatal("服务器根目录 路径检查出错")
	}
	if !exist {
		if err := os.MkdirAll(s.rootDir, os.ModePerm); err != nil {
			log.Fatal("文件夹创建出错")
		}
	}

	// 若开启匿名访问服务，则检查匿名文件夹是否存在
	if s.enableAnonymous {
		anonymousPath := filepath.Join(s.rootDir, anonymous)
		exist, err = utils.VerifyPath(anonymousPath)
		if err != nil {
			log.Fatal("匿名根目录 路径检查出错")
		}
		if !exist {
			if err := os.MkdirAll(anonymousPath, os.ModePerm); err != nil {
				log.Fatal("文件夹创建出错")
			}
		}
	}

	log.Print("服务器搭建成功")
	return s
}

func (s *server) newConnection(conn *net.TCPConn) *Connection {
	r := &Connection{
		linkConn:  conn,
		wt:        bufio.NewWriter(conn),
		rt:        bufio.NewReader(conn),
		connectID: cid,
		server:    s,
	}
	cid++
	return r
}

func (s *server) Listen() error {
	// 开始监听
	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%d", InitCfg.PortCfg.LinkPort))
	if err != nil {
		return err
	}
	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}
	// 持续监听并建立连接
	for {
		tcpConn, err := tcpListener.AcceptTCP()
		if err != nil {
			log.Panic("连接建立失败")
		}
		log.Print("建立新连接")
		go s.newConnection(tcpConn).handle()
	}
}

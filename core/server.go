package core

import (
	"bufio"
	"fmt"
	. "github.com/Aldrice/liteFTP/common/config"
	"github.com/Aldrice/liteFTP/common/datebase"
	"github.com/Aldrice/liteFTP/utils"
	"log"
	"math/rand"
	"net"
	"os"
	"path/filepath"
)

type server struct {
	// 服务器支持的通用指令集
	stdCommand cmdMap
	// 服务器支持的自定义指令集
	udmCommand cmdMap
	// 被动模式下最大端口
	pasvMaxPort int
	// 被动模式下最小端口
	pasvMinPort int
	// 是否允许匿名访问
	enableAnonymous bool
	// 服务器根目录基址
	rootDir string
	// 服务器用户存储目录基址
	userDir string
	// 服务器系统文件目录基址
	systDir string
	// 数据库对象
	srvDB *datebase.SrvDB
}

func NewServer() *server {
	absPath, err := filepath.Abs(InitCfg.SrvCfg.RootDir)
	if err != nil {
		log.Fatal("路径转换出错")
	}
	s := &server{
		stdCommand:      loadAllCommands(stdCMD),
		udmCommand:      loadAllCommands(udmCMD),
		pasvMaxPort:     InitCfg.PortCfg.MaxPasvPort,
		pasvMinPort:     InitCfg.PortCfg.MinPasvPort,
		enableAnonymous: InitCfg.SrvCfg.EnableAnonymous,
		rootDir:         absPath,
		userDir:         filepath.Join(absPath, UserPath),
		systDir:         filepath.Join(absPath, SystPath),
	}

	// 检查服务器的存储路径是否有效
	exist, err := utils.VerifyPath(s.userDir)
	if err != nil {
		log.Fatal("服务器根目录 路径检查出错")
	}
	if !exist {
		if err := os.MkdirAll(s.userDir, os.ModePerm); err != nil {
			log.Fatal("文件夹创建出错")
		}
	}
	// 检查服务器的系统文件路径是否有效
	exist, err = utils.VerifyPath(s.systDir)
	if err != nil {
		log.Fatal("服务器根目录 路径检查出错")
	}
	if !exist {
		if err := os.MkdirAll(s.systDir, os.ModePerm); err != nil {
			log.Fatal("文件夹创建出错")
		}
	}

	// 为该服务端初始化一个数据库
	s.srvDB, err = datebase.InitDB(InitCfg.DBCfg.DriverName, filepath.Join(s.systDir, DataSource))
	if err != nil {
		log.Fatal("服务器数据库 初始化失败")
	}

	// 若开启匿名访问服务，则检查匿名文件夹是否存在
	if s.enableAnonymous {
		anonymousPath := filepath.Join(s.userDir, Anonymous)
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
		rd:        bufio.NewReader(conn),
		connectID: cid,
		server:    s,
		isPassive: false,
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

func (s *server) getPassivePort() int {
	return s.pasvMinPort + rand.Intn(s.pasvMaxPort-s.pasvMinPort)
}

func (s *server) CloseDB() {
	s.srvDB.CloseDB()
}

package main

import (
	"encoding/json"
	. "github.com/Aldrice/liteFTP/common/config"
	"github.com/Aldrice/liteFTP/core"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// todo: 实现多FTP服务器在同一个机器上运行
func main() {
	x, err := os.ReadFile("./config.json")
	if err != nil {
		log.Fatal("配置文件读取失败")
	}
	if err := json.Unmarshal(x, &InitCfg); err != nil {
		log.Fatal("配置文件解析失败")
	}
	// 播随机种子
	rand.Seed(time.Now().UnixNano())

	// 检查导入的参数是否正确
	verifyConfig()

	// 新建一个FTP服务器对象
	srv := core.NewServer()
	// 在程序退出时关闭数据库
	defer srv.CloseDB()

	// ftp服务
	go func() {
		if err := srv.Listen(); err != nil {
			log.Fatal("建立监听失败, 错误信息: ", err.Error())
		}
	}()

	{
		osSignals := make(chan os.Signal, 1)
		signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)
		<-osSignals
		log.Print("成功关闭程序")
		close(osSignals)
	}
}

func verifyConfig() {
	if InitCfg.DBCfg.ForeignKey != 1 && InitCfg.DBCfg.ForeignKey != 0 {
		log.Fatal("配置文件有误 外键启用设置只能为 1 或 0")
	}
	if InitCfg.PortCfg.MinPasvPort >= InitCfg.PortCfg.MaxPasvPort {
		log.Fatal("配置文件有误 最大端口必须大于最大端口")
	}
}

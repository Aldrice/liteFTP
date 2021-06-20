package core

import (
	"context"
	"github.com/Aldrice/liteFTP/common/utils"
	"io/fs"
	"os"
	"strings"
	"time"
)

// STOR 传输文件, 若文件已有, 则覆盖旧文件
var STOR = &command{
	name:        []string{"STOR"},
	demandAuth:  false,
	demandLogin: true,
	demandParam: true,
	cmdFunction: func(conn *Connection, params string) (*response, error) {
		// 检查参数数量
		ps, ok := utils.VerifyParams(params, 1)
		if !ok {
			return respParamsError, nil
		}
		// 处理路径
		newPath := conn.processPath(ps[0])
		if newPath == "" {
			return createResponse(550, "The pathname wasn't exist or no authorization to process."), nil
		}
		// 打开已有文件 或 新建文件
		f, err := os.OpenFile(newPath, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return createResponse(550, "An error occur when creating or opening the file: "+err.Error()), err
		}
		defer f.Close()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		return conn.readData(ctx, f)
	},
}

var LIST = &command{
	name:        []string{"LIST"},
	demandAuth:  false,
	demandLogin: true,
	demandParam: false,
	cmdFunction: func(conn *Connection, params string) (*response, error) {
		path := conn.liedDir
		if params != "" {
			ps, ok := utils.VerifyParams(params, 1)
			if !ok {
				return respParamsError, nil
			}
			exist, err := utils.VerifyPath(ps[0])
			if err != nil {
				return createResponse(550, "An error occur when verifying the data: "+err.Error()), err
			}
			if !exist {
				return createResponse(550, "The path wasn't exist or no authorization to process."), err
			}
			path = ps[0]
		}
		dir, err := os.ReadDir(path)
		if err != nil {
			return createResponse(550, "An error occur when opening the dictionary: "+err.Error()), err
		}

		var files []fs.FileInfo
		for _, entry := range dir {
			info, err := entry.Info()
			if err == nil {
				files = append(files, info)
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		return conn.writeData(ctx, strings.NewReader(utils.FormatFileList(files)), "Transferring the dir entries.")
	},
}

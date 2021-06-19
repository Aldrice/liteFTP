package core

import (
	"context"
	"github.com/Aldrice/liteFTP/common/utils"
	"os"
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
			return createResponse(550, "The pathname wasn't exist or no authorization to process"), nil
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

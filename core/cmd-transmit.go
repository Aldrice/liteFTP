package core

import (
	"github.com/Aldrice/liteFTP/common/utils"
	"os"
	"path/filepath"
)

// MKD 注: 只允许用户逐级创建文件夹
var MKD = &command{
	name:        "MKD",
	demandAuth:  false,
	demandLogin: true,
	demandParam: true,
	// todo: 仍然不可用
	cmdFunction: func(conn *Connection, params string) (*response, error) {
		ps, ok := utils.VerifyParams(params, 1)
		if !ok {
			return respParamsError, nil
		}
		// 检查文件夹名是否正常
		if !utils.VerifyFolderName(ps[0]) {
			return &response{
				code: 550,
				info: "the folder's name was unacceptable for the server's os",
			}, nil
		}
		if err := os.Mkdir(filepath.Join(conn.liedDir, ps[0]), os.ModePerm); err != nil {
			return &response{
				code: 550,
				info: "an error occur when the server creating the new file, " + err.Error(),
			}, err
		}
		return createResponse(250, "directory created"), nil
	},
}

package datebase

import (
	"github.com/Aldrice/liteFTP/utils"
)

// CreateUser 创建用户
func (db *SrvDB) CreateUser(username, password string) error {
	_, err := db.client.Exec("INSERT INTO user(username, password) values(?,?)", username, utils.HashStrings(password))
	if err != nil {
		return err
	}
	return nil
}

// DeleteUser 删除用户
func (db *SrvDB) DeleteUser(username string) error {
	_, err := db.client.Exec("delete from user where username=?", username)
	if err != nil {
		return err
	}
	return nil
}

// VerifyUser 根据用户名和密码检查该用户是否可以登录, 或者根据用户名检查该用户是否存在
func (db *SrvDB) VerifyUser(u ...string) (bool, error) {
	var cnt int64
	username := u[0]
	if len(u) > 1 {
		err := db.client.QueryRow("select count(*) from user where username = ? and password = ?", username, utils.HashStrings(u[1])).Scan(&cnt)
		if err != nil {
			return false, err
		}
	} else {
		err := db.client.QueryRow("select count(*) from user where username = ?", username).Scan(&cnt)
		if err != nil {
			return false, err
		}
	}
	return cnt > 0, nil
}

// ChangePassword 根据用户名更改该用户的密码
func (db *SrvDB) ChangePassword(username, password string) error {
	_, err := db.client.Exec("update user set password = ? where username = ?", utils.HashStrings(password), username)
	if err != nil {
		return err
	}
	return nil
}

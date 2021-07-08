package datebase

import (
	"fmt"
	"github.com/Aldrice/liteFTP/common/config"
	"github.com/Aldrice/liteFTP/utils"
	"strings"
)

type User struct {
	Uid      int
	Username string
	Password string
	IsAdmin  int
}

// CreateUser 创建用户
func (db *SrvDB) CreateUser(username, password string) error {
	_, err := db.client.Exec("INSERT INTO user(username, password) values(?,?)",
		strings.ToLower(username),
		utils.HashStrings(strings.ToLower(password)),
		0,
	)
	if err != nil {
		return err
	}
	return nil
}

func (u *User) FormatToList() string {
	isAdmin := "true"
	if u.IsAdmin == 0 {
		isAdmin = "false"
	}
	format := "\r\n===============\r\n\tUser ID: %d\r\n\tUsername: %s\r\n\tIs Admin: %s\r\n==============="
	return fmt.Sprintf(format, u.Uid, u.Username, isAdmin)
}

func FormatUsers(us []User) (text string) {
	for _, u := range us {
		text += u.FormatToList()
	}
	return text
}

// VerifyUser 根据用户名和密码检查该用户是否可以登录, 或者根据用户名检查该用户是否存在
func (db *SrvDB) VerifyUser(u ...string) (bool, error) {
	var cnt int64
	username := strings.ToLower(u[0])
	if len(u) > 1 {
		err := db.client.QueryRow("select count(*) from user where username = ? and password = ?", username, utils.HashStrings(strings.ToLower(u[1]))).Scan(&cnt)
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
	_, err := db.client.Exec("update user set password = ? where username = ?", utils.HashStrings(strings.ToLower(password)), strings.ToLower(username))
	if err != nil {
		return err
	}
	return nil
}

// IsAdmin 根据用户名检查该用户是否为管理员
func (db *SrvDB) IsAdmin(username string) (bool, error) {
	var cnt int64
	if err := db.client.QueryRow("SELECT COUNT(*) FROM user WHERE username = ? AND is_admin = ?", username, 1).Scan(&cnt); err != nil {
		return false, err
	}
	return cnt > 0, nil
}

// DeleteUser 删除用户
func (db *SrvDB) DeleteUser(username string) error {
	_, err := db.client.Exec("delete from user where username=?", strings.ToLower(username))
	if err != nil {
		return err
	}
	return nil
}

// SetAdmin 根据用户名设置某个用户为管理员
func (db *SrvDB) SetAdmin(username string) error {
	_, err := db.client.Exec("UPDATE user SET is_admin = ? WHERE username = ?", 1, strings.ToLower(username))
	if err != nil {
		return err
	}
	return nil
}

// ListUsers 列出所有用户ID和其昵称
func (db *SrvDB) ListUsers(page int) (int, []User, error) {
	cnt := 0
	if err := db.client.QueryRow("SELECT COUNT(*) FROM user").Scan(&cnt); err != nil {
		return 0, nil, err
	}
	if cnt > 0 {
		limit, offset := config.PageSize, utils.TransferPageToOffset(page, config.PageSize, cnt)
		rows, err := db.client.Query("SELECT * FROM user LIMIT ? OFFSET ?", limit, offset)
		if err != nil {
			return 0, nil, err
		}
		var us []User
		for rows.Next() {
			var u User
			if err := rows.Scan(&u.Uid, &u.Username, &u.Password, &u.IsAdmin); err != nil {
				return 0, nil, err
			}
			u.Password = ""
			us = append(us, u)
		}
		return cnt, us, nil
	}
	return 0, nil, nil
}

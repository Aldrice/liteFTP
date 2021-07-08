package datebase

import (
	"fmt"
	"github.com/Aldrice/liteFTP/common/config"
	"github.com/Aldrice/liteFTP/utils"
	"strings"
)

type Msg struct {
	Mid      int
	Username string
	Message  string
	IsRead   int
}

func (m *Msg) FormatToEntity() string {
	isRead := "true"
	if m.IsRead == 0 {
		isRead = "false"
	}
	format := "\r\n===============\r\n\tMessage ID: %d\r\n\tSend By: %s\r\n\tMessage Text: %s\r\n\tIs Read: %s\r\n==============="
	return fmt.Sprintf(format, m.Mid, m.Username, m.Message, isRead)
}

func (m *Msg) FormatToList() string {
	isRead := "true"
	if m.IsRead == 0 {
		isRead = "false"
	}
	format := "\r\n===============\r\n\tMessage ID: %d\r\n\tSend By: %s\r\n\tIs Read: %s\r\n==============="
	return fmt.Sprintf(format, m.Mid, m.Username, isRead)
}

func FormatMessages(ms []Msg) (text string) {
	for _, m := range ms {
		text += m.FormatToList()
	}
	return text
}

// CreateMessage 新建站内信
func (db *SrvDB) CreateMessage(username, message string) error {
	_, err := db.client.Exec("INSERT INTO message(username, message) VALUES (?,?)",
		strings.ToLower(username),
		message,
	)
	if err != nil {
		return err
	}
	return nil
}

// DeleteMessage 删除站内信
func (db *SrvDB) DeleteMessage(id int) error {
	_, err := db.client.Exec("DELETE FROM message WHERE mid = ?", id)
	if err != nil {
		return err
	}
	return nil
}

// ListMessages 列出所有的站内信, 可以选择列出已读或未读的站内信
func (db *SrvDB) ListMessages(page int, isRead bool) (int, []Msg, error) {
	read, cnt := 0, 0
	if isRead {
		read = 1
	}
	if err := db.client.QueryRow("SELECT COUNT(*) FROM message WHERE is_read = ?", read).Scan(&cnt); err != nil {
		return 0, nil, err
	}
	if cnt > 0 {
		limit, offset := config.PageSize, utils.TransferPageToOffset(page, config.PageSize, cnt)
		rows, err := db.client.Query("SELECT * FROM message WHERE is_read = ? LIMIT ? OFFSET ?", read, limit, offset)
		if err != nil {
			return 0, nil, err
		}
		var ms []Msg
		for rows.Next() {
			var m Msg
			if err := rows.Scan(&m.Mid, &m.Username, &m.Message, &m.IsRead); err != nil {
				return 0, nil, err
			}
			ms = append(ms, m)
		}
		return cnt, ms, nil
	}
	return 0, nil, nil
}

// GetMessage 根据 站内信id 获得 站内信
func (db *SrvDB) GetMessage(id int) (*Msg, error) {
	m := new(Msg)
	if err := db.client.QueryRow("SELECT * FROM message WHERE mid = ?", id).Scan(&m.Mid, &m.Username, &m.Message, &m.IsRead); err != nil {
		return nil, err
	}
	_, err := db.client.Exec("UPDATE message SET is_read = ? WHERE mid = ?", 1, id)
	if err != nil {
		return nil, err
	}
	return m, nil
}

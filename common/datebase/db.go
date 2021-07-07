package datebase

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type SrvDB struct {
	client     *sql.DB
	driverName string
	dataSource string
}

// InitDB 给出数据库文件路径, 根据数据库路径生成一个DB对象
func InitDB(driverName, dataSource string) (*SrvDB, error) {
	db, err := sql.Open(driverName, dataSource)
	if err != nil {
		return nil, err
	}

	// 若表不存在，则创建一个表
	table := `
    CREATE TABLE IF NOT EXISTS user (
        uid INTEGER PRIMARY KEY AUTOINCREMENT,
        username VARCHAR(256) NOT NULL UNIQUE,
        password BLOB(16) NOT NULL,
        is_admin INTEGER NOT NULL                        
    );`
	_, err = db.Exec(table)
	if err != nil {
		return nil, err
	}

	return &SrvDB{client: db, driverName: driverName, dataSource: dataSource}, nil
}

// CloseDB 关闭DB
func (db *SrvDB) CloseDB() {
	_ = db.client.Close()
}

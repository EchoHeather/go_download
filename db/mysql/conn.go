package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
)

var db *sql.DB

//init mysql初始化
func init() {
	db, _ = sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/fileserver?charset=utf8")
	db.SetMaxOpenConns(1000)
	err := db.Ping()
	if err != nil {
		fmt.Println("Failed to connect to mysql, err : " + err.Error())
		os.Exit(1)
	}
}

//DBConn 返回数据库对象实例,内部是基于连接池实现的，因此不用过于担心并发的问题
func DBConn() *sql.DB {
	return db
}

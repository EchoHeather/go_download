package db

import (
	"fmt"
	mydb "goWork/db/mysql"
)

//UserSignup 注册用户名、密码
func UserSignup(username, password string) bool {
	stmt, err := mydb.DBConn().Prepare("insert ignore into tbl_user(`user_name`, `user_pwd`) values (?, ?)")
	if err != nil {
		fmt.Println("Failed to insert , err :" + err.Error())
		return false
	}
	defer stmt.Close()

	ret, err := stmt.Exec(username, password)
	if err != nil {
		fmt.Println("Failed to insert , err :" + err.Error())
		return false
	}

	if rowAffected, err := ret.RowsAffected(); err == nil && rowAffected > 0 {
		return true
	}
	return false
}

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

//UserSignIn 判断密码是否一致
func UserSignIn(username, password string) bool {
	stmt, err := mydb.DBConn().Prepare("select * from tbl_user where user_name = ? limit 1")
	if err != nil {
		fmt.Println("Failed to select , err :" + err.Error())
		return false
	}
	defer stmt.Close()

	rows, err := stmt.Query(username)

	if err != nil {
		fmt.Println("Failed to select , err :" + err.Error())
		return false
	} else if rows == nil {
		fmt.Println("username not found")
		return false
	}

	pRows := mydb.ParseRows(rows)
	if len(pRows) > 0 && string(pRows[0]["user_pwd"].([]byte)) == password {
		return true
	}
	return false
}

//UpdateToken 刷新用户登录token
func UpdateToken(username, token string) bool {
	stmt, err := mydb.DBConn().Prepare("replace into tbl_user_token(`user_name`, `user_token`) VALUES (?, ?)")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, token)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

//GetToken 获取用户登录token
func GetToken(username string) (string, error) {
	stmt, err := mydb.DBConn().Prepare("select user_token from tbl_user_token where user_name = ? limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	defer stmt.Close()
	var token string
	err = stmt.QueryRow(username).Scan(&token)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	return token, nil
}

type User struct {
	Username   string
	Email      string
	Phone      string
	SignupAt   string
	LastActive string
	Status     int
}

//GetUserInfo 获取用户信息
func GetUserInfo(username string) (User, error) {
	user := User{}

	stmt, err := mydb.DBConn().Prepare("select user_name, signup_at from tbl_user where user_name = ? limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return user, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(username).Scan(&user.Username, &user.SignupAt)
	if err != nil {
		fmt.Println(err.Error())
		return user, err
	}

	return user, nil
}

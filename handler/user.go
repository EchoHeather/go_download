package handler

import (
	dblayer "goWork/db"
	"goWork/util"
	"io/ioutil"
	"net/http"
)

const pwd_salt = "*#890"

//SignupHandler 处理用户请求
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data, err := ioutil.ReadFile("static/view/signup.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)
		return
	}
	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")

	if len(username) < 3 || len(password) < 5 {
		w.Write([]byte("invalid parameter"))
		return
	}
	enc_password := util.Sha1([]byte(password + pwd_salt))
	suc := dblayer.UserSignup(username, enc_password)
	if suc {
		w.Write([]byte("SUCCESS"))
		return
	}
	w.Write([]byte("FAILED"))
	return
}

func SignInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		//获取账号密码
		r.ParseForm()
		username := r.Form.Get("username")
		password := r.Form.Get("password")
		encPassWd := util.Sha1([]byte(password + pwd_salt))

		//验证账号密码
		dblayer.UserSignIn(username, encPassWd)

		//登陆成功并跳转

	}
}

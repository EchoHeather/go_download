package handler

import (
	"fmt"
	dblayer "goWork/db"
	"goWork/util"
	"io/ioutil"
	"net/http"
	"time"
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

//SignInHandler 用户登陆
func SignInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		//获取账号密码
		r.ParseForm()
		username := r.Form.Get("username")
		password := r.Form.Get("password")
		encPassWd := util.Sha1([]byte(password + pwd_salt))

		//验证账号密码
		pwdChecked := dblayer.UserSignIn(username, encPassWd)

		if !pwdChecked {
			w.Write([]byte("FAILED"))
			return
		}
		//生成token
		token := GenToken(username)
		upRes := dblayer.UpdateToken(username, token)
		if !upRes {
			w.Write([]byte("FAILED"))
			return
		}
		//登陆成功并跳转
		resp := util.RespMsg{
			Code: 0,
			Msg:  "OK",
			Data: struct {
				Location string
				Username string
				Token    string
			}{
				Location: "http://" + r.Host + "/static/view/home.html",
				Username: username,
				Token:    token,
			},
		}
		w.Write(resp.JSONBytes())
		return
	}
	data, err := ioutil.ReadFile("static/view/signin.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
	return
}

//GenToken 获取token 40位字符md5(username + timestamp + token_salt) + timestamp[:8]
func GenToken(username string) string {
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username + ts + "_tokensalt"))
	return tokenPrefix + ts[:8]
}

//UserInfoHandler 获取用户信息
func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.Form.Get("username")
	//验证token是否有效
	//token := r.Form.Get("token")
	//isValidToken := IsTokenValid(username, token)
	//if !isValidToken {
	//	w.WriteHeader(http.StatusForbidden)
	//	return
	//}
	user, err := dblayer.GetUserInfo(username)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	resp := util.RespMsg{
		Code: 0,
		Msg:  "ok",
		Data: user,
	}
	w.Write(resp.JSONBytes())
}

//IsTokenValid 验证token
func IsTokenValid(username, token string) bool {
	//判断长度
	if len(token) != 40 {
		fmt.Println("token invalid: " + token)
		return false
	}
	//判断token后8位的时间Hex2Dec
	t := token[32:40]
	if util.Hex2Dec(t) < time.Now().Unix()-86400 {
		fmt.Println("token time out: " + token)
		return false
	}

	//判断数据库token是否一致
	tokenDb, err := dblayer.GetToken(username)
	if err != nil {
		fmt.Println("token invalid: " + err.Error())
		return false
	}
	if token != tokenDb {
		return false
	}
	return true
}

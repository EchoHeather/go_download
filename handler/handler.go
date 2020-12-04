package handler

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

//UploadHandler 处理上传文件
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		//返回上传html
		data, err := ioutil.ReadFile("./static/view/index.html")
		if err != nil {
			io.WriteString(w, "internel server error")
			return
		}
		io.WriteString(w, string(data))
	} else if r.Method == "POST" {
		//接收文件流及存储到本地目录
		file, header, err := r.FormFile("file")
		if err != nil {
			fmt.Printf("Faield to get data, err : %s", err.Error())
			return
		}
		defer file.Close()

		//创建文件
		newFile, err := os.Create("D:\\goWork\\static\\log\\" + header.Filename)
		if err != nil {
			fmt.Printf("Failed to create file, err : %s", err.Error())
			return
		}
		defer newFile.Close()

		//copy源文件内容存入新文件内
		_, err = io.Copy(newFile, file)
		if err != nil {
			fmt.Printf("Failed to save data into file, err : %s", err.Error())
			return
		}
		http.Redirect(w, r, "/file/upload/suc", http.StatusFound)
	}
}

//UploadSucHandler 上传成功
func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload finished!")
}

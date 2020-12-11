package main

import (
	"fmt"
	"goWork/handler"
	"net/http"
)

func main() {
	// 静态资源处理
	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("./static"))))
	//路由
	http.HandleFunc("/file/upload", handler.UploadHandler)  //根目录
	http.HandleFunc("/file/upload/suc", handler.UploadSucHandler)  //上传成功
	http.HandleFunc("/file/meta", handler.GetFileMetaHandler) //获取元文件信息

	//监听端口
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Failed to start server, err : %s", err.Error())
	}
}

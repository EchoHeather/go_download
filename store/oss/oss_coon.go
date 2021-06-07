package oss

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	cfg "goWork/config"
)

var ossCli *oss.Client

//OssClient oss实例化
func OssClient() *oss.Client {
	if ossCli != nil {
		return ossCli
	}
	// 创建OSSClient实例。
	ossCli, err := oss.New(cfg.OssEndpoint, cfg.OssAccess, cfg.OssSecret)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	return ossCli
}

//Bucket oss桶
func Bucket() *oss.Bucket {
	// 获取存储空间。
	cli := OssClient()
	bucket, err := cli.Bucket(cfg.OssBucket)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}
	return bucket
}

//DownloadUrl 临时下载url
func DownloadUrl(objName string) string {
	bucket := Bucket()
	signedURL, err := bucket.SignURL(objName, oss.HTTPGet, 3600)
	if err != nil {
		fmt.Println("err:" + err.Error())
		return ""
	}
	return signedURL
}

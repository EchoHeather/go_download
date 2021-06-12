package main

import (
	"bufio"
	"encoding/json"
	"goWork/config"
	dblayer "goWork/db"
	"goWork/mq"
	"goWork/store/oss"
	"log"
	"os"
)

func ProcessTransfer(msg []byte) bool {
	//解析msg
	pubDta := mq.TransferData{}
	err := json.Unmarshal(msg, pubDta)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	//获取临时存储路径
	filed, err := os.Open(pubDta.CurLocation)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	//写入oss
	err = oss.Bucket().PutObject(
		pubDta.DestLocation,
		bufio.NewReader(filed),
	)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	//更新文件表信息
	suc := dblayer.UpdateFileLocation(pubDta.FileHash, pubDta.DestLocation)
	if !suc {
		return false
	}

	return true
}

func main() {
	mq.StartConsumer(config.TransQueueErrName, config.TransRoutingkey, ProcessTransfer)
}

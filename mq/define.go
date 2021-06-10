package mq

import (
	"goWork/common"
)

//TransferData 转移消息载体的结构格式
type TransferData struct {
	//文件hash
	FileHash string
	//临时存储目录
	CurLocation string
	//目标地址(例oss)
	DestLocation string
	//文件转移类型(oss、ceph)
	DestStoreType common.StoreType
}

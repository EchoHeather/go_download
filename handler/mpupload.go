package handler

import (
	"fmt"
	"goWork/cache/redis"
	"goWork/util"
	"math"
	"net/http"
	"strconv"
	"time"
)

type MultipartUploadInfo struct {
	FileHash   string
	FileSize   int
	UploadID   string
	ChunkSize  int
	ChunkCount int
}

func InitialMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	//获取参数
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "params invalid", nil))
		return
	}

	//获取redis链接
	rConn := redis.RedisPool().Get()
	defer rConn.Close()

	//生成分块上传信息
	upInfo := MultipartUploadInfo{
		FileHash:   filehash,
		FileSize:   filesize,
		UploadID:   username + fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunkSize:  5 * 1024 * 1024, //5MB
		ChunkCount: int(math.Ceil(float64(filesize) / (5 * 1024 * 1024))),
	}

	//写入redis
	args := []interface{}{"MP_" + upInfo.UploadID}
	kvs := map[string]interface{"chunkcount" : upInfo.ChunkCount, "filehash" : upInfo.FileHash, "filesize" : upInfo.FileSize}

	//返还给客户端信息
}

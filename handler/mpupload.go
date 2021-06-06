package handler

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	rPool "goWork/cache/redis"
	dblayer "goWork/db"
	"goWork/util"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

/**
   上传步骤
		1.客户端请求InitialMultipartUploadHandler初始化信息获取返回结果
		2.客户端根据初始化接口返回的块size进行分块上传
		3.上传完成后，请求分快完成接口
		4.服务端收到信息后对文件进行合并、入库、返回成功信息
	模拟客户端分块上传
		cd test
		go run test_mpupload.go
*/

type MultipartUploadInfo struct {
	FileHash   string
	FileSize   int
	UploadID   string
	ChunkSize  int
	ChunkCount int
}

//InitialMultipartUploadHandler 初始化分块上传
func InitialMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	//获取参数
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "params invalid", nil).JSONBytes())
		return
	}

	//获取redis链接
	rConn := rPool.RedisPool().Get()
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
	fmt.Println(upInfo.UploadID)
	args := []interface{}{"MP_" + upInfo.UploadID}
	kvs := map[string]interface{}{
		"chunkcount": upInfo.ChunkCount,
		"filehash":   upInfo.FileHash,
		"filesize":   upInfo.FileSize,
	}
	for key, value := range kvs {
		args = append(args, key, value)
	}
	_, err = rConn.Do("HMSET", args...)
	if err != nil {
		fmt.Printf("Failed to add redis hash key, err : %s", err.Error())
		return
	}
	//返还给客户端信息
	w.Write(util.NewRespMsg(0, "ok", upInfo).JSONBytes())
	return
}

//UploadPartHandler 上传文件分块
func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	//获取客户端信息
	r.ParseForm()
	uploadID := r.Form.Get("uploadid")
	chunkIndex := r.Form.Get("index")

	//获取redis连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	//获取文件句柄并存储
	fpath := "/data/" + uploadID + "/" + chunkIndex
	os.MkdirAll(path.Dir(fpath), 0744)
	fd, err := os.Create(fpath)
	if err != nil {
		w.Write(util.NewRespMsg(-1, "Upload pard failed", nil).JSONBytes())
		return
	}
	defer fd.Close()
	buf := make([]byte, 1024*1024)
	for {
		//存入到变量buf中,返回n<len(buf)
		n, err := r.Body.Read(buf)
		fd.Write(buf[:n])
		if err != nil {
			break
		}
	}

	//更新redis
	rConn.Do("HSET", "MP_"+uploadID, "chkidx_"+chunkIndex, 1)

	//返回结果
	w.Write(util.NewRespMsg(0, "ok", nil).JSONBytes())
}

//CompleteUploadHandler 查看文件上传状态
func CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	//获取客户端参数
	r.ParseForm()
	username := r.Form.Get("username")
	upid := r.Form.Get("uploadid")
	filehash := r.Form.Get("filehash")
	filesize := r.Form.Get("filesize")
	filename := r.Form.Get("filename")

	//获取redis连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	//通过uploadID判断所有文件是否上传完毕
	data, err := redis.Values(rConn.Do("HGETALL", "MP_"+upid))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "complete upload failed", nil).JSONBytes())
		return
	}
	totalCount := 0
	ChunkCount := 0
	for i := 0; i < len(data); i += 2 {
		k := string(data[i].([]byte))
		v := string(data[i+1].([]byte))
		if k == "chunkcount" {
			totalCount, _ = strconv.Atoi(v)
		} else if strings.HasPrefix(k, "chkidx_") && v == "1" {
			ChunkCount += 1
		}
	}
	//判断分块文件总数和上传完成分块数量是否一致
	if totalCount != ChunkCount {
		w.Write(util.NewRespMsg(-1, "invalid request", nil).JSONBytes())
		return
	}

	//调用shell进行合并
	//pass()

	//更新mysql
	fsize, _ := strconv.Atoi(filesize)
	dblayer.OnFileUploadFinished(filehash, filename, int64(fsize), "")
	dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(fsize))

	//返回结果
	w.Write(util.NewRespMsg(0, "ok", nil).JSONBytes())
}

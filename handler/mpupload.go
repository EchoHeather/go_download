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

const (
	ChunkDir          = "/data/chunks/" //上传分块目录
	MergeDire         = "/data/merge/"  //合并后的目录
	ChunkKeyPrefix    = "MP_"           //分块信息对应的redis key前缀
	HashUpIDKeyPrefix = "HASH_UPID_"    //文件hash映射的uploadid对应的redis key前缀
)

func init() {
	if err := os.MkdirAll(ChunkDir, 0744); err != nil {
		fmt.Println("创建上传分块目录失败: " + ChunkDir)
		os.Exit(1)
	}
	if err := os.MkdirAll(MergeDire, 0744); err != nil {
		fmt.Println("创建合并分块目录失败: " + MergeDire)
		os.Exit(1)
	}
}

type MultipartUploadInfo struct {
	FileHash    string
	FileSize    int
	UploadID    string
	ChunkSize   int
	ChunkCount  int
	ChunkExists []int //已上传完成的分块索引列表
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

	uploadID := ""

	//判断是否是断点续传
	keyExists, _ := redis.Bool(rConn.Do("EXISTS", HashUpIDKeyPrefix+filehash))
	if keyExists {
		uploadID, err = redis.String(rConn.Do("GET", HashUpIDKeyPrefix+filehash))
		if err != nil {
			w.Write(util.NewRespMsg(-1, "Upload part failed", err.Error()).JSONBytes())
			return
		}
	}

	//首次上传则创建信息，断点续传则获取信息
	ChunkExists := []int{}
	if uploadID == "" {
		uploadID = username + fmt.Sprintf("%x", time.Now().UnixNano())
	} else {
		chunks, err := redis.Values(rConn.Do("HGETALL", ChunkKeyPrefix+uploadID))
		if err != nil {
			w.Write(util.NewRespMsg(-1, "complete upload failed", nil).JSONBytes())
			return
		}
		for i := 0; i < len(chunks); i += 2 {
			k := string(chunks[i].([]byte))
			v := string(chunks[i+1].([]byte))
			if strings.HasPrefix(k, "chkidx_") && v == "1" {
				chkidx, _ := strconv.Atoi(k[7:len(k)])
				ChunkExists = append(ChunkExists, chkidx)
			}
		}
	}

	//生成分块上传信息
	upInfo := MultipartUploadInfo{
		FileHash:    filehash,
		FileSize:    filesize,
		UploadID:    uploadID,
		ChunkSize:   5 * 1024 * 1024, //5MB
		ChunkCount:  int(math.Ceil(float64(filesize) / (5 * 1024 * 1024))),
		ChunkExists: ChunkExists,
	}

	//写入redis
	if len(upInfo.ChunkExists) <= 0 {
		hkey := ChunkKeyPrefix + upInfo.UploadID
		args := []interface{}{hkey}
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
		rConn.Do("EXPIRE", hkey, 86400)

		//添加key，下次可判断是否是断点续传
		_, err = rConn.Do("SET", HashUpIDKeyPrefix+filehash, upInfo.UploadID, "EX", 86400)
		if err != nil {
			fmt.Printf("Failed to add redis hash key, err : %s", err.Error())
			return
		}
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

	//获取redis判断是否是正在上传的状态
	val, err := redis.Int(rConn.Do("HGET", ChunkKeyPrefix+uploadID, "chkidx_"+chunkIndex))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "Upload part failed", nil).JSONBytes())
		return
	}
	if val == 2 {
		w.Write(util.NewRespMsg(-1, "Upload part is writing", nil).JSONBytes())
		return
	}

	//获取文件句柄并存储
	fpath := "/data/" + uploadID + "/" + chunkIndex
	os.MkdirAll(path.Dir(fpath), 0744)
	fd, err := os.Create(fpath)
	if err != nil {
		w.Write(util.NewRespMsg(-1, "Upload part failed", nil).JSONBytes())
		return
	}
	defer fd.Close()

	//设置redis为上传中
	rConn.Do("HSET", ChunkKeyPrefix+uploadID, "chkidx_"+chunkIndex, 2)

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
	rConn.Do("HSET", ChunkKeyPrefix+uploadID, "chkidx_"+chunkIndex, 1)

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
	data, err := redis.Values(rConn.Do("HGETALL", ChunkKeyPrefix+upid))
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
	//合并后应判断是否有相同hash文件，有则删除此文件(优化)
	//pass()

	//更新mysql
	fsize, _ := strconv.Atoi(filesize)
	dblayer.OnFileUploadFinished(filehash, filename, int64(fsize), "")
	dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(fsize))

	//返回结果
	w.Write(util.NewRespMsg(0, "ok", nil).JSONBytes())
}

//CancelUploadHandler 取消上传接口
func CancelUploadHandler(w http.ResponseWriter, r *http.Request) {
	//解析参数
	r.ParseForm()
	filehash := r.Form.Get("filehash")

	//获取redis连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	//检查uploadID是否存在，存在即删除
	uploadID, err := redis.String(rConn.Do("GET", HashUpIDKeyPrefix+filehash))
	if err != nil || uploadID == "" {
		w.Write(util.NewRespMsg(-1, "Cancel upload part failed", nil).JSONBytes())
		return
	}
	_, delHashErr := rConn.Do("DEL", HashUpIDKeyPrefix+filehash)
	_, delUploadInfoErr := rConn.Do("DEL", ChunkKeyPrefix+uploadID)
	if delHashErr != nil || delUploadInfoErr != nil {
		w.Write(util.NewRespMsg(-2, "Cancel upload part failed", nil).JSONBytes())
		return
	}

	//删除已上传分块文件
	delChunkRes := util.RemovePathByShell(ChunkDir + uploadID)
	if !delChunkRes {
		fmt.Printf("Failed to delete chunks as upload canceled, uploadID:%s\n", uploadID)
	}

	//返回结果
	w.Write(util.NewRespMsg(0, "ok", nil).JSONBytes())
	return
}

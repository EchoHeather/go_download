package meta

//Filemeta 文件元信息结构
type Filemeta struct {
	FileSha1 string
	FileName string
	FileSize int64
	Location string
	UploadAt string
}

var fileMetas map[string]Filemeta

//init 	fileMetas初始化
func init() {
	fileMetas = make(map[string]Filemeta)
}

//UpdateFileMeta 更新到tree上
func UpdateFileMeta(fmeta Filemeta) {
	fileMetas[fmeta.FileSha1] = fmeta
}

//GetFileMeta 获取tree内指定的Filemeta
func GetFileMeta(filesha1 string) Filemeta {
	return fileMetas[filesha1]
}

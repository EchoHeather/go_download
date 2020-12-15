package meta

import "sort"

//FileMeta 文件元信息结构
type FileMeta struct {
	FileSha1 string
	FileName string
	FileSize int64
	Location string
	UploadAt string
}

var fileMetas map[string]FileMeta

//init 	fileMetas初始化
func init() {
	fileMetas = make(map[string]FileMeta)
}

//UpdateFileMeta 更新到tree上
func UpdateFileMeta(fmeta FileMeta) {
	fileMetas[fmeta.FileSha1] = fmeta
}

//GetFileMeta 获取tree内指定的Filemeta
func GetFileMeta(filesha1 string) FileMeta {
	return fileMetas[filesha1]
}

//DeleteFileMeta 删除tree内指定的Filemeta
func RemoveFileMeta(filesha1 string) {
	delete(fileMetas, filesha1)
}

//GetLastFileMetas 获取批量的文件元信息列表
func GetLastFileMetas(count int) []FileMeta {
	var fMetaArray []FileMeta
	for _, v := range fileMetas {
		fMetaArray = append(fMetaArray, v)
	}
	sort.Sort(ByUploadTime(fMetaArray))
	if count > len(fMetaArray) {
		return fMetaArray
	}
	return fMetaArray[0:count]
}

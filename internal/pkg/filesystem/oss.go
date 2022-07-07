package filesystem

import (
	"bytes"
	"fmt"
	"mime/multipart"

	"go-chat/config"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type OssFilesystem struct {
	conf   *config.Config
	client *oss.Client
}

func NewOssFilesystem(conf *config.Config) *OssFilesystem {

	client, _ := oss.New(conf.Filesystem.Oss.Endpoint, conf.Filesystem.Oss.AccessID, conf.Filesystem.Oss.AccessSecret)

	return &OssFilesystem{conf, client}
}

//上传文件流
func (o *OssFilesystem) UploadByte(fileName string, fs []byte) error {
	// 填写存储空间名称，例如examplebucket。
	bucket, err := o.client.Bucket(o.conf.Filesystem.Oss.Bucket)
	if err != nil {
		fmt.Println("上传文件出错:", err.Error())
		return err
	}
	err = bucket.PutObject(fileName, bytes.NewReader(fs))
	if err != nil {
		return err
	}
	return nil
}

// Write 文件写入
func (c *OssFilesystem) Write(data []byte, filePath string) error {

	return nil
}

// WriteLocal 本地文件上传
func (c *OssFilesystem) WriteLocal(localFile string, filePath string) error {
	return nil
}

func (c *OssFilesystem) WriteFromFile(file *multipart.FileHeader, filePath string) error {
	return nil
}

// Copy 文件拷贝
func (c *OssFilesystem) Copy(srcPath, filePath string) error {

	return nil
}

// Delete 删除一个文件或空文件夹
func (c *OssFilesystem) Delete(filePath string) error {
	return nil
}

// DeleteDir 删除文件夹
func (c *OssFilesystem) DeleteDir(path string) error {
	return nil
}

// CreateDir 递归创建文件夹
func (c *OssFilesystem) CreateDir(path string) error {

	return nil
}

// Stat 文件信息
func (c *OssFilesystem) Stat(filePath string) (*FileStat, error) {
	return nil, nil
}

func (c *OssFilesystem) Append(filePath string) {

}

//获取公共地址
func (o *OssFilesystem) PublicUrl(filePath string) string {
	return fmt.Sprintf("https://11zb.%s/%s", o.conf.Filesystem.Oss.Endpoint, filePath)
}

func (c *OssFilesystem) PrivateUrl(filePath string, timeout int) string {

	return filePath
}

// ReadStream 读取文件流信息
func (c *OssFilesystem) ReadStream(filePath string) ([]byte, error) {
	return nil, nil
}

//初始化分片上传
func (o *OssFilesystem) InitiateMultipartUpload(filePath string, fileName string) (string, error) {
	bucket, _ := o.client.Bucket(o.conf.Filesystem.Oss.Bucket)
	resp, err := bucket.InitiateMultipartUpload(fileName)
	if err != nil {
		return "", err
	}
	return resp.UploadID, nil
}

func (c *OssFilesystem) UploadPart(filePath string, uploadID string, num int, stream []byte) (string, error) {
	return "", nil
}

func (c *OssFilesystem) CompleteMultipartUpload(filePath string, uploadID string, opt interface{}) error {
	return nil
}

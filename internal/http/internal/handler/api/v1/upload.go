package v1

import (
	"fmt"
	"go-chat/config"
	"go-chat/internal/entity"
	"go-chat/internal/http/internal/response"
	"go-chat/internal/pkg/filesystem"
	"go-chat/internal/service"
	"path"
	"strconv"
	"time"

	"github.com/GUAIK-ORG/go-snowflake/snowflake"
	"github.com/gin-gonic/gin"
)

type Upload struct {
	config     *config.Config
	filesystem *filesystem.Filesystem
	service    *service.SplitUploadService
}

func NewUploadHandler(
	config *config.Config,
	filesystem *filesystem.Filesystem,
	service *service.SplitUploadService,
) *Upload {
	return &Upload{
		config:     config,
		filesystem: filesystem,
		service:    service,
	}
}

// 头像文件上传
func (u *Upload) Avatar(ctx *gin.Context) {

	file, err := ctx.FormFile("image")
	if err != nil {
		response.InvalidParams(ctx, "文件上传失败！")
		return
	}

	ext := path.Ext(file.Filename)
	fs, _ := filesystem.ReadMultipartStream(file)
	s, _ := snowflake.NewSnowflake(int64(0), int64(0))
	val := s.NextVal()
	fileName := fmt.Sprintf("chat/avatar/%s/%s%s", time.Now().Format("20060102"), strconv.FormatInt(val, 10), ext)
	if err := u.filesystem.Oss.UploadByte(fileName, fs); err != nil {
		response.BusinessError(ctx, err.Error())
		return
	}
	response.Success(ctx, entity.H{
		"image": u.filesystem.Oss.PublicUrl(fileName),
	})
}

// 其他文件上传
func (u *Upload) File(ctx *gin.Context) {

	file, err := ctx.FormFile("file")
	if err != nil {
		response.InvalidParams(ctx, "文件上传失败！")
		return
	}
	ext := path.Ext(file.Filename)
	fs, _ := filesystem.ReadMultipartStream(file)
	s, _ := snowflake.NewSnowflake(int64(0), int64(0))
	val := s.NextVal()
	fileName := fmt.Sprintf("chat/file/%s/%s%s", time.Now().Format("20060102"), strconv.FormatInt(val, 10), ext)

	if err := u.filesystem.Oss.UploadByte(fileName, fs); err != nil {
		response.BusinessError(ctx, err.Error())
		return
	}

	response.Success(ctx, entity.H{
		"file": u.filesystem.Oss.PublicUrl(fileName),
	})
}

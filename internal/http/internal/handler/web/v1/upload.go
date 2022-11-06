package v1

import (
	"fmt"
	"path"
	"strconv"
	"time"

	"go-chat/config"
	"go-chat/internal/entity"
	"go-chat/internal/http/internal/dto/web"
	"go-chat/internal/pkg/filesystem"
	"go-chat/internal/pkg/ichat"
	"go-chat/internal/service"

	"github.com/GUAIK-ORG/go-snowflake/snowflake"
)

type Upload struct {
	config     *config.Config
	filesystem *filesystem.Filesystem
	service    *service.SplitUploadService
}

func NewUpload(config *config.Config, filesystem *filesystem.Filesystem, service *service.SplitUploadService) *Upload {
	return &Upload{config: config, filesystem: filesystem, service: service}
}

// 头像文件上传
func (u *Upload) Avatar(ctx *ichat.Context) error {

	file, err := ctx.Context.FormFile("image")
	if err != nil {
		return ctx.InvalidParams("参数[image]为必填")

	}

	ext := path.Ext(file.Filename)
	fs, _ := filesystem.ReadMultipartStream(file)
	s, _ := snowflake.NewSnowflake(int64(0), int64(0))
	val := s.NextVal()
	fileName := fmt.Sprintf("chat/avatar/%s/%s%s", time.Now().Format("20060102"), strconv.FormatInt(val, 10), ext)
	if err := u.filesystem.Oss.UploadByte(fileName, fs); err != nil {
		return ctx.BusinessError(err.Error())

	}
	return ctx.Success(entity.H{
		"image": u.filesystem.Oss.PublicUrl(fileName),
	})
}

// 其他文件上传
func (u *Upload) File(ctx *ichat.Context) error {

	file, err := ctx.Context.FormFile("file")
	if err != nil {
		return ctx.InvalidParams("文件上传失败")

	}
	ext := path.Ext(file.Filename)
	fs, _ := filesystem.ReadMultipartStream(file)
	s, _ := snowflake.NewSnowflake(int64(0), int64(0))
	val := s.NextVal()
	fileName := fmt.Sprintf("chat/file/%s/%s%s", time.Now().Format("20060102"), strconv.FormatInt(val, 10), ext)

	if err := u.filesystem.Oss.UploadByte(fileName, fs); err != nil {
		return ctx.BusinessError(err.Error())

	}

	return ctx.Success(entity.H{
		"file": u.filesystem.Oss.PublicUrl(fileName),
	})
}

// InitiateMultipart 批量上传初始化
func (u *Upload) InitiateMultipart(ctx *ichat.Context) error {

	params := &web.UploadInitiateMultipartRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	info, err := u.service.InitiateMultipartUpload(ctx.RequestCtx(), &service.MultipartInitiateOpts{
		Name:   params.FileName,
		Size:   params.FileSize,
		UserId: ctx.UserId(),
	})
	if err != nil {
		return ctx.BusinessError(err.Error())
	}

	return ctx.Success(entity.H{
		"upload_id":  info.UploadId,
		"split_size": 2 << 20,
	})
}

// MultipartUpload 批量分片上传
func (u *Upload) MultipartUpload(ctx *ichat.Context) error {

	params := &web.UploadMultipartRequest{}
	if err := ctx.Context.ShouldBind(params); err != nil {
		return ctx.InvalidParams(err)
	}

	file, err := ctx.Context.FormFile("file")
	if err != nil {
		return ctx.InvalidParams("文件上传失败")
	}

	err = u.service.MultipartUpload(ctx.RequestCtx(), &service.MultipartUploadOpts{
		UserId:     ctx.UserId(),
		UploadId:   params.UploadId,
		SplitIndex: params.SplitIndex,
		SplitNum:   params.SplitNum,
		File:       file,
	})
	if err != nil {
		return ctx.BusinessError(err.Error())
	}

	if params.SplitIndex != params.SplitNum-1 {
		return ctx.Success(entity.H{"is_merge": false})
	}

	return ctx.Success(entity.H{"is_merge": true, "upload_id": params.UploadId})
}

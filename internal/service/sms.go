package service

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"go-chat/internal/pkg/strutil"
	"go-chat/internal/repository/cache"
)

type SmsService struct {
	*BaseService
	smsCodeCache *cache.SmsCodeCache
}

func NewSmsService(baseService *BaseService, codeCache *cache.SmsCodeCache) *SmsService {
	return &SmsService{BaseService: baseService, smsCodeCache: codeCache}
}

// CheckSmsCode 验证短信验证码是否正确
func (s *SmsService) CheckSmsCode(ctx context.Context, channel string, mobile string, code string) bool {
	value, err := s.smsCodeCache.Get(ctx, channel, mobile)

	return err == nil && value == code
}

// DeleteSmsCode 删除短信验证码记录
func (s *SmsService) DeleteSmsCode(ctx context.Context, channel string, mobile string) {
	_ = s.smsCodeCache.Del(ctx, channel, mobile)
}

// SendSmsCode 发送短信
func (s *SmsService) SendSmsCode(ctx context.Context, channel string, mobile string) (string, error) {
	// todo 需要做防止短信攻击处理

	code := strutil.GenValidateCode(6)

	// ... 请求第三方短信接口
	// fmt.Println("正在发送短信验证码：", code)
	// responseCode := map[string]string{
	// 	"0":  "短信发送成功",
	// 	"-1": "参数不全",
	// 	"-2": "服务器空间不支持,请确认支持curl或者fsocket，联系您的空间商解决或者更换空间！",
	// 	"30": "密码错误",
	// 	"40": "账号不存在",
	// 	"41": "余额不足",
	// 	"42": "帐户已过期",
	// 	"43": "IP地址限制",
	// 	"50": "内容含有敏感词",
	// }

	smsapi := "http://api.smsbao.com/" //短信平台帐号
	account := "zhiboakak"             //短信平台密码
	password := md5.Sum([]byte("zhibo999"))
	smscontent := ""
	if s.config.GetEnv() == "alone" {
		smscontent = "【liaoqiu】您的本次验证码为:" + string(code) + ",该验证码5分钟有效"
	} else {
		smscontent = "【11zb】您的本次验证码为:" + string(code) + ",该验证码5分钟有效"
	}
	ss := fmt.Sprintf("%x", password)
	sendurl := smsapi + "sms?u=" + account + "&p=" + ss + "&m=" + mobile + "&c=" + url.QueryEscape(smscontent)
	res, err := http.Get(sendurl)

	if err != nil {
		defer res.Body.Close()
		return "", err
	} else {
		defer res.Body.Close()
	}
	body, _ := ioutil.ReadAll(res.Body)

	if string(body) == "0" {
		// 添加发送记录
		if err := s.smsCodeCache.Set(ctx, channel, mobile, code, 60*5); err != nil {
			return "", err
		}
		return code, nil
	} else {
		return "", errors.New("发送短信错误")
	}
}

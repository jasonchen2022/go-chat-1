package push

import (
	"context"
	"go-chat/config"
	"log"
	"strconv"
	"time"

	"github.com/dacker-soul/getui/auth"
	"github.com/dacker-soul/getui/publics"
	"github.com/dacker-soul/getui/push/list"
	"github.com/dacker-soul/getui/push/single"
	"github.com/go-redis/redis/v8"
)

type GeTuiService struct {
	conf *config.Config
	rds  *redis.Client
}

func NewGeTuiService(conf *config.Config, rds *redis.Client) *GeTuiService {
	return &GeTuiService{conf: conf, rds: rds}
}

//获取个推token,23个小时缓存
func (s *GeTuiService) getGeTuiToken(ctx context.Context) string {
	key := "getGetGeTuiToken"
	result := s.rds.Get(ctx, key).Val()
	if result == "" {
		confLocal := publics.GeTuiConfig{
			AppId:        s.conf.GeTui.AppId, // 个推提供的id，密码等
			AppSecret:    s.conf.GeTui.AppSecret,
			AppKey:       s.conf.GeTui.AppKey,
			MasterSecret: s.conf.GeTui.MasterSecret,
		}
		// 1.获取token，注意这个token过期时间为一天+1秒，每分钟调用量为100次，每天最大调用量为10万次
		data, err := auth.GetToken(ctx, confLocal)
		if err != nil {
			log.Fatalln("error:", err)
			return result
		}
		result = data.Data.Token
		s.rds.Set(ctx, key, result, time.Duration(23)*time.Hour)
	}
	return result

}

func (s *GeTuiService) getConfig() publics.GeTuiConfig {
	confLocal := publics.GeTuiConfig{
		AppId:        s.conf.GeTui.AppId, // 个推提供的id，密码等
		AppSecret:    s.conf.GeTui.AppSecret,
		AppKey:       s.conf.GeTui.AppKey,
		MasterSecret: s.conf.GeTui.MasterSecret,
	}
	return confLocal
}

func (s *GeTuiService) PushSingleByCid(ctx context.Context, cid string, title string, body string) {
	// 2.1ios厂商通道的参数
	iosChannel := publics.IosChannel{
		Type: "",
		Aps: &publics.Aps{
			Alert: &publics.Alert{
				Title: title,
				Body:  body,
			},
			ContentAvailable: 0,
		},
		AutoBadge:      "+1",
		PayLoad:        "",
		Multimedia:     nil,
		ApnsCollapseId: "",
	}

	// 2.2安卓通道和个推通道的普通推送参数
	notification := publics.Notification{
		Title:       title,
		Body:        body,
		ClickType:   "startapp", // 打开应用首页
		BadgeAddNum: 1,
	}
	singleParam := single.PushSingleParam{
		RequestId: strconv.FormatInt(time.Now().UnixNano(), 10), // 请求唯一标识号
		Audience: &publics.Audience{ // 目标用户
			Cid:           []string{cid}, // cid推送数组
			Alias:         nil,           // 别名送数组
			Tag:           nil,           // 推送条件
			FastCustomTag: "",            // 使用用户标签筛选目标用户
		},
		Settings: &publics.Settings{ // 推送条件设置
			TTL: 3600000, // 默认一小时，消息离线时间设置，单位毫秒
			Strategy: &publics.Strategy{ // 厂商通道策略，具体看public_struct.go
				Default: 1,
				Ios:     4,
				St:      4,
				Hw:      4,
				Xm:      4,
				Vv:      4,
				Mz:      4,
				Op:      4,
			},
			Speed:        100, // 推送速度，设置100表示：100条/秒左右，0表示不限速
			ScheduleTime: 0,   // 定时推送时间，必须是7天内的时间，格式：毫秒时间戳
		},
		PushMessage: &publics.PushMessage{
			Duration:     "", // 手机端通知展示时间段
			Notification: &notification,
			Transmission: "",
			Revoke:       nil,
		},
		PushChannel: &publics.PushChannel{
			Ios: &iosChannel,
			Android: &publics.AndroidChannel{Ups: &publics.Ups{
				Notification: &notification,
				TransMission: "", // 透传消息内容，与notification 二选一
			}},
		},
	}
	config := s.getConfig()
	token := s.getGeTuiToken(ctx)
	log.Println("token:", token)
	// 3.执行单推
	singleResult, err := single.PushSingleByCid(ctx, config, token, &singleParam)
	if err != nil {
		log.Fatalln("error:", err)
	}
	log.Println("result:", singleResult)

}

func (s *GeTuiService) PushSingleByCids(ctx context.Context, cid []string, title string, body string) {
	// 2.1ios厂商通道的参数
	iosChannel := publics.IosChannel{
		Type: "",
		Aps: &publics.Aps{
			Alert: &publics.Alert{
				Title: title,
				Body:  body,
			},
			ContentAvailable: 0,
		},
		AutoBadge:      "+1",
		PayLoad:        "",
		Multimedia:     nil,
		ApnsCollapseId: "",
	}

	// 2.2安卓通道和个推通道的普通推送参数
	notification := publics.Notification{
		Title:       title,
		Body:        body,
		ClickType:   "startapp", // 打开应用首页
		BadgeAddNum: 1,
	}
	config := s.getConfig()
	token := s.getGeTuiToken(ctx)

	listMessageResult, err := list.PushListMessage(ctx, config, token, &list.PushListMessageParam{
		RequestId: strconv.FormatInt(time.Now().UnixNano(), 10), // 请求唯一标识号
		Settings: &publics.Settings{ // 推送条件设置
			TTL: 3600000, // 默认一小时，消息离线时间设置，单位毫秒
			Strategy: &publics.Strategy{ // 厂商通道策略，具体看public_struct.go
				Default: 1,
				Ios:     4,
				St:      4,
				Hw:      4,
				Xm:      4,
				Vv:      4,
				Mz:      4,
				Op:      4,
			},
			Speed:        100, // 推送速度，设置100表示：100条/秒左右，0表示不限速
			ScheduleTime: 0,   // 定时推送时间，必须是7天内的时间，格式：毫秒时间戳
		},
		PushMessage: &publics.PushMessage{
			Duration:     "", // 手机端通知展示时间段
			Notification: &notification,
			Transmission: "",
			Revoke:       nil,
		},
		PushChannel: &publics.PushChannel{
			Ios: &iosChannel,
			Android: &publics.AndroidChannel{Ups: &publics.Ups{
				Notification: &notification,
				TransMission: "", // 透传消息内容，与notification 二选一
			}},
		},
	})
	if err != nil {
		log.Fatalln("PushListMessage:", err)
	}

	singleParam := list.PushListCidParam{
		IsAsync: false,
		TaskId:  listMessageResult.Data["taskid"],
		Audience: &publics.Audience{ // 目标用户
			Cid:           cid, // cid推送数组
			Alias:         nil, // 别名送数组
			Tag:           nil, // 推送条件
			FastCustomTag: "",  // 使用用户标签筛选目标用户
		},
	}

	// 3.执行批量推
	listResult, err := list.PushListCid(ctx, config, token, &singleParam)
	if err != nil {
		log.Fatalln("PushListCid:", err)
	}
	log.Println("result:", listResult)

}

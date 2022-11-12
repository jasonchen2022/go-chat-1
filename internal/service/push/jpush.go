package push

import (
	"go-chat/config"

	"github.com/go-redis/redis/v8"
	"github.com/zwczou/jpush"
)

type JpushService struct {
	conf *config.Config
	rds  *redis.Client
}

func NewJpushService(conf *config.Config, rds *redis.Client) *JpushService {
	return &JpushService{conf: conf, rds: rds}
}

func (p *JpushService) PushMessageByCid(cid string, title string, body string) (string, error) {
	client := jpush.New(p.conf.JPush.AppKey, p.conf.JPush.AppSecret)
	payload := &jpush.Payload{
		Platform: jpush.NewPlatform().All(),
		Audience: jpush.NewAudience().SetRegistrationId(cid),
		Notification: &jpush.Notification{
			Alert: title,
			Android: &jpush.AndroidNotification{
				Alert: body,
				Title: title,
			},
			Ios: &jpush.IosNotification{
				Alert: map[string]string{
					"title": title,
					"body":  body,
				},
				Badge: "+1",
			},
		},
		Options: &jpush.Options{
			TimeLive:       60,
			ApnsProduction: false,
		},
	}
	return client.Push(payload)
}

package config

// Redis Redis配置信息
type JPush struct {
	AppKey    string `json:"app_key" yaml:"app_key"`
	AppSecret string `json:"app_secret" yaml:"app_secret"`
}

package config

// Redis Redis配置信息
type GeTui struct {
	AppId        string `json:"app_id" yaml:"app_id"`
	AppKey       string `json:"app_key" yaml:"app_key"`
	AppSecret    string `json:"app_secret" yaml:"app_secret"`
	MasterSecret string `json:"master_secret" yaml:"master_secret"`
}

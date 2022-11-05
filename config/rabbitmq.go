package config

// Redis Redis配置信息
type RabbitMQ struct {
	Host         string `json:"host" yaml:"host"`         // 服务器IP地址
	Port         int    `json:"port" yaml:"port"`         // 服务器端口号
	UserName     string `json:"username" yaml:"username"` // 服务器端口号
	Password     string `json:"password" yaml:"password"` // 密码
	ExchangeName string `json:"exchange_name" yaml:"exchange_name"`
}

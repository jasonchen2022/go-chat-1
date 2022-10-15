package config

type App struct {
	AppName string `json:"app_name"`
	Port    int    `json:"port"`
	Debug   bool   `json:"debug"`
	Welcome string `json:"welcome" yaml:"welcome"`
	JuheKey string `json:"juhe_key" yaml:"juhe_key"`
}

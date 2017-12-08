package config

import (
	"github.com/koding/multiconfig"
)

type (
	Config struct {
		title string
		Mongo struct {
			Host     string `default:"127.0.0.1"`
			Database string
			Post     string `default:"27017"`
			UserName string
			PassWord string
		}
		HttpClient struct {
			ProxyDisable  bool    `default:false`
			ProxyHost     string  `default:"nil"`
			ProxyPort     string  `default:1080`
			ProxyUser     string  `default:"user"`
			ProxyPassword string  `default:"pass"`
			Timeout       float64 `default:30`
		}
		Redis struct {
			Host string `default:"localhost:6379"`
			Db   int
		}

		FilePath string `default:"/data/"`
	}
)

func Get() *Config {
	appConfig := new(Config)
	m := multiconfig.NewWithPath("config.json")
	m.MustLoad(appConfig)

	return appConfig
}

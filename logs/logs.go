package logs

import (
	"centipede/config"

	"github.com/sirupsen/logrus"
	// "github.com/weekface/mgorus"
)

func New() *logrus.Logger {
	var log = logrus.New()

	log.Level = logrus.ErrorLevel

	appConfig := config.Get()

	// hooker, err := mgorus.NewHooker(appConfig.Mongo.Host+":"+appConfig.Mongo.Post, appConfig.Mongo.Database, "log")

	// if err == nil {
	// 	log.Hooks.Add(hooker)
	// }

	return log
}

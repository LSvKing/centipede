package logs

import (
	"centipede/config"

	"github.com/sirupsen/logrus"

	"github.com/x-cray/logrus-prefixed-formatter"
)

func New() *logrus.Logger {
	var log = logrus.New()
	log.Formatter = new(prefixed.TextFormatter)
	log.Level = logrus.DebugLevel

	appConfig := config.Get()

	var hooker logrus.Hook
	var err error

	if appConfig.Mongo.UserName == "" {
		hooker, err = NewHooker(appConfig.Mongo.Host+":"+appConfig.Mongo.Port, appConfig.Mongo.Database, "log")

	} else {
		hooker, err = NewHookerWithAuthDb(appConfig.Mongo.Host+":"+appConfig.Mongo.Port, appConfig.Mongo.Database, appConfig.Mongo.Database, "log", appConfig.Mongo.UserName, appConfig.Mongo.PassWord)

	}

	if err == nil {
		log.Hooks.Add(hooker)
	}

	return log
}

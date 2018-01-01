package common

import (
	"context"
	"strings"
	"time"

	"centipede/config"

	"github.com/go-redis/redis"
	"github.com/imroc/req"
	"golang.org/x/time/rate"
)

func GetProxy() string {

	appConfig := config.Get()
	client := redis.NewClient(&redis.Options{
		Addr: appConfig.Redis.Host,
		DB:   1,
	})

	client.Expire("proxy", 30*time.Second)

	ctx, _ := context.WithCancel(context.Background())

	limiter := rate.NewLimiter(1, 1)

	for {
		limiter.Wait(ctx)

	ReGoto:
		r, err := req.Get("http://H196AR4J9408XN6D:F766CDA5666E4627@http-dyn.abuyun.com:9020")
		if err != nil {
			goto ReGoto
		}

		client.Set("proxy:"+strings.TrimSpace(r.String()), "http://"+strings.TrimSpace(r.String()), 30*time.Second)
	}
}

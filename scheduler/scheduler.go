package scheduler

import (
	"encoding/json"
	"sync"

	"github.com/LSvKing/centipede/config"
	"github.com/LSvKing/centipede/logs"
	"github.com/LSvKing/centipede/request"
	"github.com/go-redis/redis"
)

type Scheduler struct {
	locker *sync.Mutex
	client *redis.Client
}

var log = logs.New()

func New() *Scheduler {
	appConfig := config.Get()

	locker := new(sync.Mutex)
	client := redis.NewClient(&redis.Options{
		Addr: appConfig.Redis.Host,
		DB:   appConfig.Redis.Db,
	})

	return &Scheduler{locker, client}
}

func (this *Scheduler) Push(r *request.Request) {
	this.locker.Lock()

	jsonReq, _ := json.Marshal(r)

	log.Debug(r)

	this.client.RPush("scheuler", jsonReq)

	this.locker.Unlock()
}

func (this *Scheduler) Poll() *request.Request {
	this.locker.Lock()

	jsonReq, err := this.client.LPop("scheuler").Bytes()

	if err != nil {
		log.Errorf(err.Error())
		this.locker.Unlock()
		return nil
	}

	var req *request.Request

	json.Unmarshal(jsonReq, &req)

	this.locker.Unlock()
	return req
}

func (this *Scheduler) Count() int {
	this.locker.Lock()
	len := this.client.LLen("scheuler").Val()
	this.locker.Unlock()
	return int(len)
}

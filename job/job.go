package cronjob

import (
	"fmt"
	"encoding/json"

	"github.com/LSvKing/cron"
	"github.com/nanobox-io/golang-scribble"
	"github.com/LSvKing/centipede/centipede"
)

var (
	cronCore *cron.Cron
	db       *scribble.Driver
	dbErr    error
)

type (
	Job struct {
		Name string
		Spec string
		Job  JobFunc
		Disable bool
	}

	JobFunc func()
)

func init()  {
	cronCore = cron.New()

	db, dbErr = scribble.New("data", nil)

	if dbErr != nil {
		fmt.Println("Error", dbErr)
	}
}

func New() {

	cronCore = cron.New()

	db, dbErr = scribble.New("data", nil)

	if dbErr != nil {
		fmt.Println("Error", dbErr)
	}
}

func Run() {
	j := Job{}

	p,_ := db.ReadAll("jobs")

	for _,v := range p{
		json.Unmarshal([]byte(v),&j)

		if j.Disable{
			j.Job = func() {
				centipede.PushCrawler(j.Name)
			}

			cronCore.AddJob(j.Spec,j.Job,j.Name)
		}
	}

	cronCore.Start()
}

func Add(job Job) {
	err := db.Write("jobs", job.Name, job)
	if err != nil {
		fmt.Println(err)
	}

	job.Job = func() {
		centipede.PushCrawler(job.Name)
	}

	cronCore.AddJob(job.Spec, job.Job, job.Name)
}

func Stop() {
	cronCore.Stop()
}

func Del(name string) {
	err := db.Delete("jobs", name)

	if err != nil {
		fmt.Println("Error", dbErr)
	}

	cronCore.RemoveJob(name)

}

func (jobFunc JobFunc) Run() {
	jobFunc()
}


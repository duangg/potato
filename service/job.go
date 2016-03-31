package service

import (
	"log"
	"github.com/robfig/cron"
	"github.com/SpruceX/potato/models"
)

const (
	INSERT = iota
	DELETE
	UPDATE
	SELECT
	ENABLE
	DISABLE
)

const (
	START = iota
	RUNNING
	STOP
)

// for RunStatus flags
const (
	QUIET = iota
	ACTIVE
	FAILURE
)

type Job struct {
	Id        string
	Sched     interface{}
	State     int
	notice    chan string
	SchedItem *models.SchedItem
	Dispatch  *Dispatch
}

func NewJob(id string, state int, sched interface{}, item *models.SchedItem, dispatch *Dispatch) *Job {
	return &Job{
		Id:        id,
		State:     state,
		Sched:     sched,
		notice:    make(chan string),
		SchedItem: item,
		Dispatch:  dispatch,
	}
}

func (job *Job) Run() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("find panic", err)
		}
	}()

	crond := cron.New()
	excuteCommand := job.makeCommand
	crond.Schedule(job.Sched.(cron.Schedule), cron.FuncJob(excuteCommand))
	crond.Start()
	job.State = RUNNING

	select {
	case <-job.notice:
		crond.Stop()
		job.State = STOP
	}
}

func (job *Job) Stop() {
	job.notice <- job.Id
}

func TimerCheckFormat(timer string) (sched cron.Schedule, err error) {
	schedule, err := cron.Parse(timer)
	if err != nil {
		return nil, err
	}

	return schedule, nil
}

func (job *Job) makeCommand() {
	action := job.Dispatch
	action.Send()
}

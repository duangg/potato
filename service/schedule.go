package service

import (
	"errors"
	"time"
	"log"

	"gopkg.in/mgo.v2/bson"

	"github.com/SpruceX/potato/models"
	"github.com/SpruceX/potato/store"
)

type Scheduler struct {
	crontab map[string]*Job
	Crons *store.CrontabStore
}

func NewSched() *Scheduler {
	return &Scheduler{
		crontab: make(map[string]*Job),
		Crons: store.Store.Crontab,
	}
}

func NewCron(id bson.ObjectId, status int, types int, timer string, host string, st, rt time.Time) *models.SchedItem {
	return &models.SchedItem{
		Id:          id,
		Status:      status,
		RunStatus:   QUIET,
		Type:        types,
		Timer:       timer,
		Host:        host,
		StartTime:   st,
		RefreshTime: rt,
	}
}

func (sched *Scheduler) Start() {
	set, err := sched.Crons.Traversal()
	if err != nil {
		log.Printf("mongoDB traversal null: %s\n", err.Error())
		return
	}

	for _, item := range set {
		id := item.Id.Hex()
		item.RunStatus = QUIET
		sched.Crons.Update(&item)

		if item.Status == DISABLE {
			continue
		}

		schedIf, err := TimerCheckFormat(item.Timer)
		if err != nil {
			log.Printf("timer format error: %s\n", err.Error())
			continue
		}
		info := NewDispatch(id, item.Host, item.Type)
		job := NewJob(id, START, schedIf, &item, info)

		sched.crontab[id] = job
		go job.Run()
	}
}

func (sched *Scheduler) Insert(job *Job) error {
	id := job.Id
	err := sched.Crons.Insert(job.SchedItem)
	if err != nil {
		return err
	}

	sched.crontab[id] = job
	go job.Run()

	return nil
}

func (sched *Scheduler) doDelete(id string) error {
	if job, ok := sched.crontab[id]; ok {
		err := sched.Crons.Delete(job.Id)
		if err != nil {
			return err
		}
		delete(sched.crontab, job.Id)

		if job.State == STOP {
			return nil
		}
		job.Stop()
	} else {
		return sched.Crons.Delete(id)
	}

	return nil
}

func (sched *Scheduler) Delete(id string) error {
	item, err := sched.Crons.Search(id)
	if err != nil {
		return err
	}

	if item.RunStatus == ACTIVE {
		return errors.New("The job is running")
	} else {
		return sched.doDelete(id)
	}

	return nil
}

func (sched *Scheduler) Update(job *Job) error {
	if oldJob, ok := sched.crontab[job.Id]; ok {
		if oldJob.State == RUNNING {
			oldJob.Stop()
		}
		job.SchedItem.Status = oldJob.SchedItem.Status
		job.SchedItem.RunStatus = oldJob.SchedItem.RunStatus
		job.SchedItem.StartTime = oldJob.SchedItem.StartTime

		err := sched.Crons.Update(job.SchedItem)
		if err != nil {
			return err
		}

		sched.crontab[job.Id] = job
		if job.SchedItem.Status != DISABLE {
			go job.Run()
		}
		return nil
	}

	item, err := sched.Crons.Search(job.Id)
	if err != nil {
		return err
	}
	item.RefreshTime = time.Now()
	item.Timer = job.SchedItem.Timer
	item.Host = job.SchedItem.Host
	item.Type = job.SchedItem.Type
	job.SchedItem = &item

	err = sched.Crons.Update(&item)
	if err != nil {
		return err
	}

	sched.crontab[job.Id] = job
	if job.SchedItem.Status != DISABLE {
		go job.Run()
	}

	return nil
}

func (sched *Scheduler) Search(id string) (models.SchedItem, error) {
	if ok := bson.IsObjectIdHex(id); !ok {
		return models.SchedItem{}, errors.New("id format error")
	}
	return sched.Crons.Search(id)
}

func (sched *Scheduler) Enable(id string) error {
	item, err := sched.Crons.Search(id)
	if err != nil {
		return err
	}

	item.Status = ENABLE
	item.StartTime = time.Now()
	err = sched.Crons.Update(&item)
	if err != nil {
		return err
	}

	if job, ok := sched.crontab[id]; ok {
		if job.State == RUNNING {
			return errors.New("This job is running")
		}
		job.SchedItem = &item
		go job.Run()
	} else {
		schedIf, err := TimerCheckFormat(item.Timer)
		if err != nil {
			return err
		}
		info := NewDispatch(item.Id.Hex(), item.Host, item.Type)
		job := NewJob(item.Id.Hex(), START, schedIf, &item, info)
		sched.crontab[id] = job
		go job.Run()
	}

	return nil
}

func (sched *Scheduler) Disable(id string) error {
	if job, ok := sched.crontab[id]; ok {
		item, err := sched.Crons.Search(id)
		if err != nil {
			return err
		}

		if item.RunStatus == ACTIVE {
			return errors.New("The job is running")
		}

		item.RefreshTime = time.Now()
		item.Status = DISABLE
		err = sched.Crons.Update(&item)
		if err != nil {
			return err
		}
		job.SchedItem = &item

		if job.State != STOP {
			job.Stop()
		}
	} else {
		return errors.New("not found id from crontab")
	}

	return nil
}

func (sched *Scheduler) UpdateJobStatus(id string, code int) error {
	item, err := sched.Search(id)
	if err != nil {
		return err
	}

	if item.Status == DISABLE {
		item.RunStatus = QUIET //Field define in the job.go
	} else {
		item.RunStatus = code
	}

	return sched.Crons.Update(&item)
}

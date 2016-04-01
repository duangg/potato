package service

import (
	"errors"
	"time"
	"log"

	"gopkg.in/mgo.v2/bson"

	"github.com/SpruceX/potato/models"
	"github.com/SpruceX/potato/store"
)

type Schedule struct {
	crontab map[string]*Job
	MongoDB *store.CrontabStore
}

func NewSched() *Schedule {
	return &Schedule{
		crontab: make(map[string]*Job),
		MongoDB: store.Store.Crontab,
	}
}

func NewItem(id bson.ObjectId, status int, types int, timer string, host string, st, rt time.Time) *models.SchedItem {
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

func InitSched() (s *Schedule) {
	sched := NewSched()
	sched.Startup()
	return sched
}

func (sched *Schedule) Startup() {
	set, err := sched.MongoDB.Traversal()
	if err != nil {
		log.Printf("mongoDB traversal null: %s\n", err.Error())
		return
	}

	for _, item := range set {
		id := item.Id.Hex()
		item.RunStatus = QUIET
		sched.MongoDB.Update(&item)

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

func (sched *Schedule) Insert(job *Job) error {
	id := job.Id
	err := sched.MongoDB.Insert(job.SchedItem)
	if err != nil {
		return err
	}

	sched.crontab[id] = job
	go job.Run()

	return nil
}

func (sched *Schedule) doDelete(id string) error {
	if job, ok := sched.crontab[id]; ok {
		err := sched.MongoDB.Delete(job.Id)
		if err != nil {
			return err
		}
		delete(sched.crontab, job.Id)

		if job.State == STOP {
			return nil
		}
		job.Stop()
	} else {
		return sched.MongoDB.Delete(id)
	}

	return nil
}

func (sched *Schedule) Delete(id string) error {
	item, err := sched.MongoDB.Search(id)
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

func (sched *Schedule) Update(job *Job) error {
	if oldJob, ok := sched.crontab[job.Id]; ok {
		if oldJob.State == RUNNING {
			oldJob.Stop()
		}
		job.SchedItem.Status = oldJob.SchedItem.Status
		job.SchedItem.RunStatus = oldJob.SchedItem.RunStatus
		job.SchedItem.StartTime = oldJob.SchedItem.StartTime

		err := sched.MongoDB.Update(job.SchedItem)
		if err != nil {
			return err
		}

		sched.crontab[job.Id] = job
		if job.SchedItem.Status != DISABLE {
			go job.Run()
		}
		return nil
	}

	item, err := sched.MongoDB.Search(job.Id)
	if err != nil {
		return err
	}
	item.RefreshTime = time.Now()
	item.Timer = job.SchedItem.Timer
	item.Host = job.SchedItem.Host
	item.Type = job.SchedItem.Type
	job.SchedItem = &item

	err = sched.MongoDB.Update(&item)
	if err != nil {
		return err
	}

	sched.crontab[job.Id] = job
	if job.SchedItem.Status != DISABLE {
		go job.Run()
	}

	return nil
}

func (sched *Schedule) Search(id string) (models.SchedItem, error) {
	if ok := bson.IsObjectIdHex(id); !ok {
		return models.SchedItem{}, errors.New("id format error")
	}
	return sched.MongoDB.Search(id)
}

func (sched *Schedule) Enable(id string) error {
	item, err := sched.MongoDB.Search(id)
	if err != nil {
		return err
	}

	item.Status = ENABLE
	item.StartTime = time.Now()
	err = sched.MongoDB.Update(&item)
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

func (sched *Schedule) Disable(id string) error {
	if job, ok := sched.crontab[id]; ok {
		item, err := sched.MongoDB.Search(id)
		if err != nil {
			return err
		}

		if item.RunStatus == ACTIVE {
			return errors.New("The job is running")
		}

		item.RefreshTime = time.Now()
		item.Status = DISABLE
		err = sched.MongoDB.Update(&item)
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

func (sched *Schedule) UpdateJobStatus(id string, code int) error {
	item, err := sched.Search(id)
	if err != nil {
		return err
	}

	if item.Status == DISABLE {
		item.RunStatus = QUIET //Field define in the job.go
	} else {
		item.RunStatus = code
	}

	return sched.MongoDB.Update(&item)
}

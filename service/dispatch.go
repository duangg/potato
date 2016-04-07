package service

import (
	"log"
)

//TODO(runyang) get rid of duplicate enums
const (
	FULL = iota
	INCR
	COMPRESS
)

type Dispatch struct {
	Host string
	Id   string
	Type int
}

func NewDispatch(id, host string, typei int) *Dispatch {
	return &Dispatch{
		Id:   id,
		Host: host,
		Type: typei,
	}
}

func (dispatch *Dispatch) Send() {
	types := dispatch.Type
//	id := dispatch.Id
	host := dispatch.Host

	switch types {
	case FULL:
		if ok := AllService.BackupService.IsInBackup(host); ok {
			return
		}
		AllService.BackupService.Execute(dispatch)
	case INCR:
		if ok := AllService.BackupService.IsInBackup(host); ok {
			return
		}
		AllService.BackupService.Execute(dispatch)
	case COMPRESS:
		AllService.BackupService.Execute(dispatch)
	default:
		log.Println("dispatch type error")
	}
}

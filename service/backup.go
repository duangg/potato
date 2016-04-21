package service

import (
	"fmt"
//	"os/exec"
//	"strings"
	"time"
	"log"

	"gopkg.in/mgo.v2/bson"
	"github.com/hypersleep/easyssh"
	"github.com/SpruceX/potato/models"
	"github.com/SpruceX/potato/store"
	"github.com/SpruceX/potato/utils"
	"github.com/SpruceX/potato/command"
)

type AsyncJobService struct {
	mongoStore *store.MongoStore
}

var ServerStatus map[string]int

func makeSshClient(b *models.Host) *easyssh.MakeConfig {
	return &easyssh.MakeConfig{
		User:     b.UserName,
		Password: "",
		Server:   b.IP,
		Port:     b.SshPort,
		Key:      utils.Cfg.SshPrivateKey,
	}
}

func GetTime() time.Time {
	return time.Now()
}

func (s AsyncJobService) isInBackup(serverName string) bool {
	if status, ok := ServerStatus[serverName]; ok {
		if status == models.JobInProgress {
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}

func (s AsyncJobService) doUpdateStatus(hostName, jobId, err string, startTime time.Time, backupLog string, runStatus, backupType int) {
	endTime := GetTime()
	ServerStatus[hostName] = runStatus
	AllService.Sched.UpdateJobStatus(jobId, runStatus)
	s.mongoStore.JobResult.SaveJobResult(hostName, jobId, err, startTime, endTime, backupLog, runStatus, backupType)
}

func (s AsyncJobService) doExecute(dispatch *HostInfo) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("find panic", err)
		}
	}()

	if s.isInBackup(dispatch.Host) {
		log.Printf("there is a running job in %s", dispatch.Host)
		return
	}

	startTime := GetTime()
	if ok := bson.IsObjectIdHex(dispatch.Id); !ok {
		str := fmt.Sprintf("wrong job id style,id:%s", dispatch.Id)
		s.doUpdateStatus(dispatch.Host, dispatch.Id, str, startTime, "", models.JobFailed, dispatch.Type)
		return
	}
	host, err := store.Store.Hosts.FindHostByName(dispatch.Host)
	if err != nil {
		str := fmt.Sprintf("can not find host : %s", dispatch.Host)
		s.doUpdateStatus(dispatch.Host, dispatch.Id, str, startTime, "", models.JobFailed, dispatch.Type)
		return
	}

	var asyncJob command.AsyncSSHJob

	switch dispatch.Type {
	case models.JobTypeFullBackup, models.JobTypeIncBackup:
		asyncJob = &command.BackupCmd{
			Target : makeSshClient(&host),
			Host : &host,
			BackType : dispatch.Type,
			JobID : dispatch.Id,
			StartTime : startTime,
		}
		break
	case models.JobTypeCompress:
		asyncJob = &command.CompressCmd{
			Target : makeSshClient(&host),
			Host : &host,
			BackType : dispatch.Type,
			JobID : dispatch.Id,
			StartTime : startTime,
		}
		break
	}

	s.doUpdateStatus(host.Name, dispatch.Id, "", startTime, "", models.JobInProgress, dispatch.Type)
	log.Printf("start job, id:%s\n", dispatch.Id)
	err = command.SSHExecutor.Execute(asyncJob)
	if err != nil {
		log.Printf(err.Error())
		s.doUpdateStatus(host.Name, dispatch.Id, err.Error(), startTime, "", models.JobFailed, dispatch.Type)
	} else {
		s.doUpdateStatus(host.Name, dispatch.Id, "", startTime, "", models.JobSucceeded, dispatch.Type)
	}
}

func (s AsyncJobService) Execute(dispatch *HostInfo) {
	go s.doExecute(dispatch)
}

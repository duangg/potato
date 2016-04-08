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

type ManualBackupService struct {
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

func (s ManualBackupService) IsInBackup(serverName string) bool {
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

func (s ManualBackupService) doUpdateStatus(hostName, jobId, err string, startTime time.Time, backupLog string, runStatus, backupType int) {
	endTime := GetTime()
	ServerStatus[hostName] = runStatus
	AllService.Sched.UpdateJobStatus(jobId, runStatus)
	s.mongoStore.JobResult.SaveJobResult(hostName, jobId, err, startTime, endTime, backupLog, runStatus, backupType)
}

func (s ManualBackupService) doExecute(dispatch *HostInfo) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("find panic", err)
		}
	}()

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

	switch dispatch.Type {
	case models.JobTypeFullBackup, models.JobTypeIncBackup:
		backupTypeStr := ""
		if dispatch.Type ==  models.JobTypeFullBackup {
			backupTypeStr = "full"
		} else {
			backupTypeStr = "inc"
		}

		backup := &command.BackupCmd{
			Target : makeSshClient(&host),
			Host : &host,
			BackType : dispatch.Type,
			JobID : dispatch.Id,
			BackupTypeStr : backupTypeStr,
			StartTime : startTime,
		}
		s.doUpdateStatus(host.Name, backup.JobID, "", startTime, "", models.JobInProgress, backup.BackType)
		log.Printf("start %s backup,job id:%s\n", backupTypeStr, backup.JobID)
		err := command.SSHExecutor.Execute(backup)
		if err != nil {
			log.Printf(err.Error())
			s.doUpdateStatus(host.Name, backup.JobID, err.Error(), startTime, "", models.JobFailed, backup.BackType)
		} else {
			s.doUpdateStatus(host.Name, backup.JobID, "", startTime, "", models.JobSucceeded, backup.BackType)
		}
		break

	case models.JobTypeCompress:
		backupTypeStr := "compress"
		compress := &command.CompressCmd{
			Target : makeSshClient(&host),
			Host : &host,
			BackType : dispatch.Type,
			JobID : dispatch.Id,
			BackupTypeStr : backupTypeStr,
			StartTime : startTime,
		}
		s.doUpdateStatus(host.Name, compress.JobID, "", startTime, "", models.JobInProgress, compress.BackType)
		log.Printf("start %s compress,job id:%s\n", backupTypeStr, compress.JobID)
		err := command.SSHExecutor.Execute(compress)
		if err != nil {
			log.Printf(err.Error())
			s.doUpdateStatus(host.Name, compress.JobID, err.Error(), startTime, "", models.JobFailed, compress.BackType)
		} else {
			s.doUpdateStatus(host.Name, compress.JobID, "", startTime, "", models.JobSucceeded, compress.BackType)
		}
		break
	}
}

func (s ManualBackupService) Execute(dispatch *HostInfo) {
	go s.doExecute(dispatch)
}

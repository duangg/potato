package service

import (
	"github.com/SpruceX/potato/models"
	"github.com/SpruceX/potato/store"
)

type Service struct {
	FileService
	BackupService
	Sched *Schedule
	User  *ApiUserService
	Topology  *TopologyService
}

type UserService interface {
	Login(username string, password string) (string, error)
	Logout(sid string) error
}

type FileService interface {
	GetFiles(b *models.Host, folder string) (error)
}

type BackupService interface {
	FullBackup(jobId, serverName string)
	IncBackup(jobId, serverName string)
	IsInBackup(serverName string) bool
	Compress(jobId, serverName string)
}

var AllService *Service

func Init() {
	AllService = NewService()
	if backup, ok := AllService.BackupService.(ManualBackupService); ok {
		backup.CheckAllRunningJobResult()
	} else {
		log.Println("failed  to convert type of service.BackupService")
	}
	AllService.Topology.TimeRefreshTopology()
}

func NewService() *Service {
	service := &Service{}
	service.BackupService = ManualBackupService{store.Store}
	ServerStatus = make(map[string]int)

	service.Sched = InitSched()
	service.User = InitUser()

	service.Topology = &TopologyService{store.Store}
	return service
}

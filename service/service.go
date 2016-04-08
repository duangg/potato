package service

import (
//	"log"

	"github.com/SpruceX/potato/models"
	"github.com/SpruceX/potato/store"
)

type Service struct {
	FileService
	BackupService
	Sched *Scheduler
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
	Execute(dispatch *HostInfo)
}

var AllService *Service

func Init() {
	AllService = NewService()
	AllService.Topology.TimeRefreshTopology()
}

func NewService() *Service {
	service := &Service{}
	service.BackupService = AsyncJobService{store.Store}
	ServerStatus = make(map[string]int)

	service.Sched = NewSched()
	service.User = InitUser()

	service.Topology = &TopologyService{store.Store}
	return service
}

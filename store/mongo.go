package store

import (
	"log"

	"gopkg.in/mgo.v2"
	"github.com/SpruceX/potato/utils"
)

type MongoStore struct {
	*mgo.Session
	Hosts            *HostStore
	Crontab          *CrontabStore
	User             *UserStore
	JobResult        *JobResultStore
	BackupFileResult *BackupFileStore
	Topology         *TopologyStore
}

const DBName = "ops-center"

var Store *MongoStore

func Init() {
	Store = New()
}

func New() *MongoStore {
	session, err := mgo.Dial(utils.Cfg.Mongo)
	if err != nil {
		log.Fatalf("couldn't connec to mongo at %s, error is %s", utils.Cfg.Mongo, err.Error())
	}
	log.Printf("connected to mongo server at %s", utils.Cfg.Mongo)
	store := &MongoStore{Session: session}
	store.Hosts = NewHostStore(store)
	store.Hosts.EnsureIndex()
	store.JobResult = NewJobResultStore(store)
	store.BackupFileResult = NewBackupFileStore(store)

	store.Crontab = NewCrontabStore(store)
	store.User = NewUserStore(store)
	store.User.UserEnsureIndex()

	store.Topology = NewTopologyStore(store)
	return store
}

func (s MongoStore) GetSession() *mgo.Session {
	return s.Copy()
}

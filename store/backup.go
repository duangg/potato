package store

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/SpruceX/potato/models"
)

type BackupFileStore struct {
	*MongoStore
}

const (
	ColBackupFile = "backupfile"
)

func NewBackupFileStore(mongo *MongoStore) *BackupFileStore {
	return &BackupFileStore{mongo}
}

func getBackupFileCol(session *mgo.Session) *mgo.Collection {
	return session.DB(DBName).C(ColBackupFile)
}

func (j *BackupFileStore) SaveBackupFileResult(servername, date string, backuptype int, jobId, filepath, size string, createtime time.Time) error {
	session := j.GetSession()
	defer session.Close()

	backupFile := models.BackupFile{
		Id:         bson.NewObjectId(),
		JobId:      bson.ObjectIdHex(jobId),
		ServerName: servername,
		Date:       date,
		BackupType: backuptype,
		FilePath:   filepath,
		Size:       size,
		Modified:   createtime,
	}

	return getBackupFileCol(session).Insert(backupFile)
}

func (j *BackupFileStore) FindBackupFileByServerName(name string) ([]models.BackupFile, error) {
	session := j.GetSession()
	defer session.Close()
	var backupFiles []models.BackupFile
	err := getBackupFileCol(session).Find(bson.M{"servername": name}).Sort("servername", "-date", "backuptype", "time").All(&backupFiles)
	return backupFiles, err
}

func (j *BackupFileStore) FindBackupFileByServerId(id string) ([]models.BackupFile, error) {
	session := j.GetSession()
	defer session.Close()
	host, err := Store.Hosts.FindHostById(id)
	if err != nil {
		return nil, err
	}
	var backupFiles []models.BackupFile
	err = getBackupFileCol(session).Find(bson.M{"servername": host.Name}).Sort("servername", "-date", "backuptype", "time").All(&backupFiles)
	return backupFiles, err
}

func (j *BackupFileStore) Find(name string, pNum, pSize int) ([]models.BackupFile, error) {
	var backupFiles []models.BackupFile
	session := j.GetSession()
	defer session.Close()
	err := getBackupFileCol(session).Find(bson.M{"servername": name}).Sort("-date", "backuptype", "time").Skip(pNum * pSize).Limit(pSize).All(&backupFiles)
	return backupFiles, err
}

func (j *BackupFileStore) GetAllBackupFile() ([]models.BackupFile, error) {
	var backupFiles []models.BackupFile
	session := j.GetSession()
	defer session.Close()
	err := getBackupFileCol(session).Find(nil).Sort("-date", "backuptype", "time").All(&backupFiles)
	return backupFiles, err
}

func (j *BackupFileStore) UpdateBackupFilePath(hostname, oldpath, newpath string) error {
	var backupFile models.BackupFile
	session := j.GetSession()
	defer session.Close()
	err := getBackupFileCol(session).Find(bson.M{"servername": hostname, "filepath": oldpath}).One(&backupFile)
	if err != nil {
		return err
	} else {
		backupFile.FilePath = newpath
		return getBackupFileCol(session).Update(bson.M{"servername": hostname, "filepath": oldpath}, backupFile)
	}
}

func (j *BackupFileStore) GetAllBackupFileCount(hostname string) (int, error) {
	session := j.GetSession()
	defer session.Close()
	num, err := getBackupFileCol(session).Find(bson.M{"servername": hostname}).Count()
	return num, err
}

package models

import (
	"gopkg.in/mgo.v2/bson"
	"time"
)

type JobResult struct {
	Id         bson.ObjectId `bson:"_id" json:"_id"`
	JobId      bson.ObjectId `bson:"job_id" json:"job_id"`
	ServerName string        `bson:"servername" json:"servername"`
	Status     int           `bson:"status" json:"status"` //1-正在备份/还原/压缩 2-备份/还原/压缩失败 3-备份/还原/压缩成功
	ErrInfo    string        `bson:"errinfo" json:"errinfo"`
	StartTime  time.Time     `bson:"starttime" json:"starttime"`
	EndTime    time.Time     `bson:"endtime" json:"endtime"`
	Output     string        `bson:"output" json:"output"`
	Type       int           `bson:"type" json:"type"` //0-全库备份 1-增量备份 2-压缩 3-还原
	Dismiss    bool          `bson:"dismiss" json:"dismiss"`
}

const (
	JobInProgress = iota + 1
	JobFailed
	JobSucceeded
)

const (
	BackupComplete = iota
	BackupError
	XtrabackupNotFind
	BackupErrorSSH
)

const (
	JobTypeFullBackup = iota
	JobTypeIncBackup
	JobTypeCompress
	JobTypeRestore
)
package models

import (
	"time"
	"gopkg.in/mgo.v2/bson"
)

type BackupFile struct {
	Id         bson.ObjectId `bson:"_id" json:"id"`
	JobId      bson.ObjectId `bson:"job_id" json:"job_id"`
	ServerName string        `bson:"servername" json:"servername"`
	Date       string        `bson:"date" json:"date"`
	BackupType int           `bson:"backuptype" json:"backuptype"`
	FilePath   string        `bson:"filepath" json:"filepath"`
	Size       string        `bson:"size" json:"size"`
	Modified   time.Time     `bson:"time" json:"time"`
}

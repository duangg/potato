package models

import (
	"time"
	"gopkg.in/mgo.v2/bson"
)

type SchedItem struct {
	Id          bson.ObjectId `bson:"_id" json:"id"`
	Status      int           `bson:"status" json:"status"`
	RunStatus   int           `bson:"runStatus" json:"runStatus"`
	Type        int           `bson:"type" json:"type"`
	Timer       string        `bson:"timer" json:"timer"`
	Host        string        `bson:"host" json:"host"`
	StartTime   time.Time     `bson:"startTime" json:"startTime"`
	RefreshTime time.Time     `bson:"refreshTime" json:"refreshTime"`
}

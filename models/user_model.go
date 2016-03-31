package models

import (
	"time"
	"gopkg.in/mgo.v2/bson"
)

type UserTable struct {
	Id       bson.ObjectId `bson:"_id" json:"id"`
	Status   int           `bson:"status" json:"status"`
	Auth     int           `bson:"auth" json:"auth"`
	Name     string        `bson:"name" json:"name"`
	Pwd      string        `bson:"pwd" json:"pwd"`
	Session  string        `bson:"session" json:"session"`
	DateTime time.Time     `bson:"DateTime" json:"DateTime"`
}

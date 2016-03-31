package models

import (
	"gopkg.in/mgo.v2/bson"
)

type HostDBLink struct {
	Id   bson.ObjectId 			`bson:"_id" json:"id"`
	SourceHostId bson.ObjectId  `bson:"sourcehostid" json:"sourcehostid"`
	SourceHostName string       `bson:"sourcehostname" json:"sourcehostname"`
	TargetHostId bson.ObjectId  `bson:"targethostid" json:"targethostid"`
	TargetHostName string       `bson:"targethostname" json:"targethostname"`
	Type string 				`bson:"type" json:"type"`
}

type Topology struct {
	Nodes   []Host 				`bson:"nodes" json:"nodes"`
	Links   []HostDBLink     	`bson:"links" json:"links"`
}

type Database struct {
	Name string					`bson:"name" json:"name"`
}

type DatabaseTable struct {
	Name string					`bson:"name" json:"name"`
	Descs []DatabaseTableDesc   `bson:"desc" json:"desc"`
}

type DatabaseTableDesc struct {
	Field string				`bson:"field" json:"field"`
	Type string					`bson:"type" json:"type"`
	IsNull string				`bson:"isnull" json:"isnull"`
	Key string					`bson:"key" json:"key"`
	Default string				`bson:"default" json:"default"`
	Comment string				`bson:"comment" json:"comment"`
}

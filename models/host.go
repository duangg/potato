package models

import "gopkg.in/mgo.v2/bson"

type Host struct {
	Id   bson.ObjectId `bson:"_id" json:"id"`
	Name string        `bson:"name" json:"name"`
	IP   string        `bson:"ip" json:"ip"`
	//	Port       string    `bson:"port" json:"port"`
	UserName string `bson:"username" json:"username"`
	//	Password   string    `bson:"password" json:"password"`
	DBUser     string `bson:"dbuser" json:"dbuser"`
	DBPassword string `bson:"dbpassword" json:"dbpassword"`
	DBPort     string `bson:"dbport" json:"dbport"`
	SshPort    string `bson:"ssh_port" json:"ssh_port"`
	BackupPath string `bson:"backuppath" json:"backuppath"`
	DBHost     string `bson:"dbhost" json:"dbhost"`
	DBSocket   string `bson:"dbsocket" json:"dbsocket"`
	DBMyCnf    string `bson:"dbmycnf" json:"dbmycnf"`
}

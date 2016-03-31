package store

import (
	"reflect"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/SpruceX/potato/models"
)

type CrontabStore struct {
	m *MongoStore
}

const ColCrontabs = "crontabs"

func type2map(h *models.SchedItem) map[string]interface{} {
	t := reflect.TypeOf(*h)
	v := reflect.ValueOf(*h)
	data := make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		data[t.Field(i).Tag.Get("bson")] = v.Field(i).Interface()
	}

	return data
}

func NewCrontabStore(mongo *MongoStore) *CrontabStore {
	return &CrontabStore{mongo}
}

func getCrontabCol(session *mgo.Session) *mgo.Collection {
	return session.DB(DBName).C(ColCrontabs)
}

func (s CrontabStore) Insert(h *models.SchedItem) error {
	session := s.m.GetSession()
	defer session.Close()
	return getCrontabCol(session).Insert(h)
}

func (s CrontabStore) Delete(id string) error {
	session := s.m.GetSession()
	defer session.Close()
	_, err := getCrontabCol(session).RemoveAll(bson.M{"_id": bson.ObjectIdHex(id)})
	return err
}

func (s CrontabStore) Update(h *models.SchedItem) error {
	session := s.m.GetSession()
	defer session.Close()
	return getCrontabCol(session).Update(bson.M{"_id": h.Id}, bson.M{"$set": type2map(h)})
}

func (s CrontabStore) Search(id string) (models.SchedItem, error) {
	session := s.m.GetSession()
	defer session.Close()
	var result models.SchedItem
	err := getCrontabCol(session).Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&result)
	return result, err
}

func (s CrontabStore) Find(pNum, pSize int) ([]models.SchedItem, error) {
	var result []models.SchedItem
	session := s.m.GetSession()
	defer session.Close()
	err := getCrontabCol(session).Find(nil).Skip(pNum * pSize).Limit(pSize).All(&result)
	return result, err
}

func (s CrontabStore) Traversal() ([]models.SchedItem, error) {
	var result []models.SchedItem
	session := s.m.GetSession()
	defer session.Close()
	err := getCrontabCol(session).Find(nil).All(&result)
	return result, err
}

func (s CrontabStore) GetCrontabsCount() (int, error) {
	session := s.m.GetSession()
	defer session.Close()
	num, err := getCrontabCol(session).Count()
	return num, err
}

//func (s CrontabStore) EnsureIndex() {
//	getCrontabCol(s.m.Session).EnsureIndex(mgo.Index{Key: []string{"name"}, Unique: true})
//}

package store

import (
	"reflect"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/SpruceX/potato/models"
)

type UserStore struct {
	m *MongoStore
}

const ColsUser = "user"

func UserType2Map(h *models.UserTable) map[string]interface{} {
	t := reflect.TypeOf(*h)
	v := reflect.ValueOf(*h)
	data := make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		data[t.Field(i).Tag.Get("bson")] = v.Field(i).Interface()
	}

	return data
}

func NewUserStore(mongo *MongoStore) *UserStore {
	return &UserStore{mongo}
}

func getCol(session *mgo.Session) *mgo.Collection {
	return session.DB(DBName).C(ColsUser)
}

func (s UserStore) UserInsert(h *models.UserTable) error {
	session := s.m.GetSession()
	defer session.Close()
	return getCol(session).Insert(h)
}

func (s UserStore) UserDelete(id string) error {
	session := s.m.GetSession()
	defer session.Close()
	_, err := getCol(session).RemoveAll(bson.M{"_id": bson.ObjectIdHex(id)})
	return err
}

func (s UserStore) DeleteByName(name string) error {
	session := s.m.GetSession()
	defer session.Close()
	_, err := getCol(session).RemoveAll(bson.M{"name": name})
	return err
}

func (s UserStore) UserUpdate(h *models.UserTable) error {
	session := s.m.GetSession()
	defer session.Close()
	return getCol(session).Update(bson.M{"_id": h.Id}, bson.M{"$set": UserType2Map(h)})
}

func (s UserStore) UserFindUserByName(name string) (models.UserTable, error) {
	var result models.UserTable
	session := s.m.GetSession()
	defer session.Close()
	err := getCol(session).Find(bson.M{"name": name}).One(&result)
	return result, err
}

func (s UserStore) UserSearch(id string) (models.UserTable, error) {
	session := s.m.GetSession()
	defer session.Close()
	var result models.UserTable
	err := getCol(session).Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&result)
	return result, err
}

func (s UserStore) UserFind(pNum, pSize int) ([]models.UserTable, error) {
	var result []models.UserTable
	session := s.m.GetSession()
	defer session.Close()
	err := getCol(session).Find(nil).Skip(pNum * pSize).Limit(pSize).All(&result)
	return result, err
}

func (s UserStore) UserTraversal() ([]models.UserTable, error) {
	var result []models.UserTable
	session := s.m.GetSession()
	defer session.Close()
	err := getCol(session).Find(nil).All(&result)
	return result, err
}

func (s UserStore) GetUsersCount() (int, error) {
	session := s.m.GetSession()
	defer session.Close()
	num, err := getCol(session).Count()
	return num, err
}

func (s UserStore) UserEnsureIndex() {
	getCol(s.m.Session).EnsureIndex(mgo.Index{Key: []string{"name"}, Unique: true})
}

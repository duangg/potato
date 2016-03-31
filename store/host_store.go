package store

import (
	"reflect"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/SpruceX/potato/models"
)

type HostStore struct {
	m *MongoStore
}

const ColHosts = "hosts"

func NewHostStore(mongo *MongoStore) *HostStore {
	return &HostStore{mongo}
}

func getHostCol(session *mgo.Session) *mgo.Collection {
	return session.DB(DBName).C(ColHosts)
}

func (s HostStore) EnsureIndex() {
	getHostCol(s.m.Session).EnsureIndex(mgo.Index{Key: []string{"name"}, Unique: true})
}

func (s HostStore) Add(h *models.Host) error {
	session := s.m.GetSession()
	defer session.Close()
	return getHostCol(session).Insert(h)
}

func (s HostStore) Find(pNum, pSize int) ([]models.Host, error) {
	var result []models.Host
	session := s.m.GetSession()
	defer session.Close()
	err := getHostCol(session).Find(nil).Skip(pNum * pSize).Limit(pSize).All(&result)
	return result, err
}

func (s HostStore) GetAllHosts() ([]models.Host, error) {
	var result []models.Host
	session := s.m.GetSession()
	defer session.Close()
	err := getHostCol(session).Find(nil).All(&result)
	return result, err
}

func (s HostStore) DeleteById(id string) error {
	session := s.m.GetSession()
	defer session.Close()
	_, err := getHostCol(session).RemoveAll(bson.M{"_id": bson.ObjectIdHex(id)})
	return err
}

func (s HostStore) DeleteByName(name string) error {
	session := s.m.GetSession()
	defer session.Close()
	_, err := getHostCol(session).RemoveAll(bson.M{"name": name})
	return err
}

func host2map(h *models.Host) map[string]interface{} {
	t := reflect.TypeOf(*h)
	v := reflect.ValueOf(*h)
	data := make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		data[t.Field(i).Tag.Get("bson")] = v.Field(i).Interface()
	}
	return data
}

func (s HostStore) Update(h *models.Host) error {
	session := s.m.GetSession()
	defer session.Close()
	//	fmt.Println(host2map(h))
	return getHostCol(session).Update(bson.M{"_id": h.Id}, bson.M{"$set": host2map(h)})
}

func (s HostStore) FindHostByName(name string) (models.Host, error) {
	var result models.Host
	session := s.m.GetSession()
	defer session.Close()
	err := getHostCol(session).Find(bson.M{"name": name}).One(&result)
	return result, err
}

func (s HostStore) FindHostByIp(ip string) (models.Host, error) {
	var result models.Host
	session := s.m.GetSession()
	defer session.Close()
	err := getHostCol(session).Find(bson.M{"ip": ip}).One(&result)
	return result, err
}

func (s HostStore) FindHostById(id string) (models.Host, error) {
	session := s.m.GetSession()
	defer session.Close()
	var result models.Host
	err := getHostCol(session).Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&result)
	return result, err
}

func (s HostStore) GetHostsCount() (int, error) {
	session := s.m.GetSession()
	defer session.Close()
	num, err := getHostCol(session).Count()
	return num, err
}

func (s HostStore) GetAllIPs() (ips []struct {
	IP string `bson:"ip"`
}, err error) {
	session := s.m.GetSession()
	defer session.Close()
	err = getHostCol(session).Find(nil).Select(bson.M{"ip": 1}).All(&ips)

	return
}

func (s HostStore) GetAllDbs() (dbs []struct {
	Name string `bson:"name" json:"name"`
	Ip   string `bson:"ip" json:"ip"`
	Port string `bson:"port" json:"port"`
}, err error) {

	session := s.m.GetSession()
	defer session.Close()
	err = getHostCol(session).Find(nil).Select(bson.M{"name": 1, "ip": 1, "port": 1}).All(&dbs)

	return
}

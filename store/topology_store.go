package store

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/SpruceX/potato/models"
)

type TopologyStore struct {
	m *MongoStore
}

const ColTopology = "topology"

func NewTopologyStore(mongo *MongoStore) *TopologyStore {
	return &TopologyStore{mongo}
}

func getTopologyCol(session *mgo.Session) *mgo.Collection {
	return session.DB(DBName).C(ColTopology)
}

func (s TopologyStore) EnsureIndex() {
}

func (s TopologyStore) GetHostDBLink() ([]models.HostDBLink, error) {
	var result []models.HostDBLink
	session := s.m.GetSession()
	defer session.Close()
	err := getTopologyCol(session).Find(nil).All(&result)
	return result, err
}

func (s TopologyStore) RemoveAllHostDBLink() error {
	session := s.m.GetSession()
	defer session.Close()
	_, err := getTopologyCol(session).RemoveAll(nil)
	return err
}

func (s TopologyStore) InsertDBLink(source, target *models.Host, nowtype string) error {
	session := s.m.GetSession()
	defer session.Close()

	var link models.HostDBLink
	err := getTopologyCol(session).Find(bson.M{"sourcehostname": source.Name, "targethostname": target.Name}).One(&link)
	if err != nil {
		hostdblink := models.HostDBLink{
			Id: bson.NewObjectId(),
			SourceHostId: source.Id,
			SourceHostName: source.Name,
			TargetHostId: target.Id,
			TargetHostName: target.Name,
			Type:nowtype,
		}
		return getTopologyCol(session).Insert(hostdblink)
	} else {
		link.Type = nowtype
		return getTopologyCol(session).Update(bson.M{"sourcehostname": source.Name, "targethostname": target.Name}, link)
	}
}

func (s TopologyStore) GetTopology() (models.Topology, error) {
	session := s.m.GetSession()
	defer session.Close()

	var topology models.Topology
	var links []models.HostDBLink

	err := getTopologyCol(session).Find(nil).All(&links)
	if err != nil {
		return topology, err
	}

	hosts, err := Store.Hosts.GetAllHosts()
	if err != nil {
		return topology, err
	}
	if links == nil {
		topology.Links = make([]models.HostDBLink, 0)
	} else {
		topology.Links = links
	}
	topology.Nodes = hosts
	return topology, err
}

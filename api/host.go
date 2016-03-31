package api

import (
	"net/http"
	"reflect"
	"strconv"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"

	"github.com/SpruceX/potato/models"
	"github.com/SpruceX/potato/utils"
)

func initHost(r *mux.Router) {
	sr := r.PathPrefix("/hosts").Subrouter()
	sr.Handle("/", ApiAppHandler(getHostsByPage)).Methods("GET")
	sr.Handle("/getallhosts", ApiAppHandler(getAllHosts)).Methods("GET")
	sr.Handle("/", ApiAppHandler(addHost)).Methods("POST")
	sr.Handle("/{id}", ApiAppHandler(deleteHost)).Methods("DELETE")
	sr.Handle("/", ApiAppHandler(updateHost)).Methods("PUT")
	sr.Handle("/hostsinfo", ApiAppHandler(getAllHostsInfo)).Methods("GET")
	sr.Handle("/ips", ApiAppHandler(getAllIPs)).Methods("GET")
	sr.Handle("/dbs", ApiAppHandler(getAllDbs)).Methods("GET")
	sr.Handle("/{name}", ApiAppHandler(getHost)).Methods("GET")
}

const (
	PageSize = 10
)

var EmptyArray []int = make([]int, 0)

func getHostsByPage(c *Context, w http.ResponseWriter, r *http.Request) {
	pageNumStr := r.FormValue("pn")
	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil {
		c.Err = NewAppError("gethosts parsing parameter", err.Error(), 400)
		return
	}
	find, err := Srv.Store.Hosts.Find(pageNum, PageSize)
	if err != nil {
		c.Err = NewAppError("gethosts find hosts", err.Error(), 500)
		return
	}
	if find == nil {
		utils.WriteJson(w, EmptyArray)
	} else {
		utils.WriteJson(w, find)
	}
}

func getAllHosts(c *Context, w http.ResponseWriter, r *http.Request) {
	find, err := Srv.Store.Hosts.GetAllHosts()
	if err != nil {
		c.Err = NewAppError("get all hosts", err.Error(), 500)
		return
	}
	if find == nil {
		utils.WriteJson(w, EmptyArray)
	} else {
		utils.WriteJson(w, find)
	}
}

func updateHost(c *Context, w http.ResponseWriter, r *http.Request) {
	var host models.Host
	err := r.ParseForm()
	if err != nil {
		c.Err = NewAppError("parse form:", err.Error(), 400)
		return
	}

	t := reflect.TypeOf(host)
	v := reflect.ValueOf(&host).Elem()
	for i := 0; i < t.NumField(); i++ {
		v.FieldByName(t.Field(i).Name).SetString(r.PostForm.Get(t.Field(i).Tag.Get("json")))
	}
	host.Id = bson.ObjectIdHex(r.PostForm.Get("id"))
	err = Srv.Store.Hosts.Update(&host)
	if err != nil {
		c.Err = NewAppError("addhost add to store:", err.Error(), 500)
		return
	}
	utils.WriteJson(w, host.Id)
}

func getHost(c *Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	result, err := Srv.Store.Hosts.FindHostByName(name)
	if err != nil {
		c.Err = NewAppError("find host by name:", err.Error(), 500)
		return
	}
	utils.WriteJson(w, result)
}

func addHost(c *Context, w http.ResponseWriter, r *http.Request) {
	var host models.Host
	err := r.ParseForm()
	if err != nil {
		c.Err = NewAppError("parse form:", err.Error(), 400)
		return
	}
	t := reflect.TypeOf(host)
	v := reflect.ValueOf(&host).Elem()
	for i := 0; i < t.NumField(); i++ {
		v.FieldByName(t.Field(i).Name).SetString(r.PostForm.Get(t.Field(i).Tag.Get("json")))
	}
	host.Id = bson.NewObjectId()
	err = Srv.Store.Hosts.Add(&host)
	if err != nil {
		c.Err = NewAppError("addhost add to store:", err.Error(), 500)
		return
	}
	utils.WriteJson(w, host.Id)
}

func deleteHost(c *Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	Id := vars["id"]
	err := Srv.Store.Hosts.DeleteById(Id)
	if err != nil {
		c.Err = NewAppError("delete host from store:", err.Error(), 500)
		return
	}
	utils.WriteJson(w, Id)
}

func getAllHostsInfo(c *Context, w http.ResponseWriter, r *http.Request) {
	sum, err := Srv.Store.Hosts.GetHostsCount()
	if err != nil {
		c.Err = NewAppError("get all hosts number", err.Error(), 500)
		return
	}

	var infoArray []int = make([]int, 2)
	infoArray[0] = sum
	infoArray[1] = PageSize
	utils.WriteJson(w, infoArray)
}

func getAllIPs(c *Context, w http.ResponseWriter, r *http.Request) {
	ips, err := Srv.Store.Hosts.GetAllIPs()
	if err != nil {
		c.Err = NewAppError("get all ips", err.Error(), 500)
		return
	}
	var ipList []string
	for _, ip := range ips {
		ipList = append(ipList, ip.IP)
	}
	utils.WriteJson(w, ipList)
}

func getAllDbs(c *Context, w http.ResponseWriter, r *http.Request) {
	dbs, err := Srv.Store.Hosts.GetAllDbs()
	if err != nil {
		c.Err = NewAppError("get all dbs", err.Error(), 500)
		return
	}
	utils.WriteJson(w, dbs)
}

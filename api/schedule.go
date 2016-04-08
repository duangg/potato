package api

import (
	"errors"
	"net/http"
	"time"
	"strconv"

	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"

	"github.com/SpruceX/potato/service"
	"github.com/SpruceX/potato/utils"
)

func initSched(r *mux.Router) {
	local := r.PathPrefix("/crons").Subrouter()
	local.Handle("/", ApiAppHandler(schedSearchAll)).Methods("GET")
	local.Handle("/", ApiAppHandler(schedInsert)).Methods("POST")
	local.Handle("/{id}", ApiAppHandler(schedDelete)).Methods("DELETE")
	local.Handle("/{id}", ApiAppHandler(schedUpdate)).Methods("PUT")
	local.Handle("/{id}", ApiAppHandler(schedSearch)).Methods("GET")
	local.Handle("/{id}/disable", ApiAppHandler(schedDisable)).Methods("GET")
	local.Handle("/{id}/enable", ApiAppHandler(schedEnable)).Methods("GET")
}

func schedInsert(c *Context, w http.ResponseWriter, r *http.Request) {
	timer := r.PostFormValue("cron")
	host := r.PostFormValue("host")
	types := r.PostFormValue("type")

	if timer == "" || host == "" || types == "" {
		c.Err = NewAppError("Param error", "Parameters cannot be empty", 500)
		return
	}

	sched, err := service.TimerCheckFormat(timer)
	if err != nil {
		c.Err = NewAppError("Timer format error", err.Error(), 500)
		return
	}

	schedType, err := strconv.Atoi(types)
	if err != nil {
		c.Err = NewAppError("Type Atoi converse error", err.Error(), 500)
		return
	}

	id := bson.NewObjectId()
	timeStamp := time.Now()
	schedItem := service.NewCron(id, service.ENABLE, schedType, timer, host, timeStamp, timeStamp)
	info := &service.HostInfo{
				Id:id.Hex(),
				Host:host,
				Type:schedType,
			}

	job := service.NewJob(id.Hex(), service.START, sched, schedItem, info)

	err = Srv.Service.Sched.Insert(job)
	if err != nil {
		c.Err = NewAppError("Insert job error", err.Error(), 500)
		return
	}
	utils.WriteJson(w, "Insert Ok")
}

func schedDelete(c *Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		c.Err = NewAppError("Param error", "Parameters cannot be empty", 500)
		return
	}

	err := Srv.Service.Sched.Delete(id)
	if err != nil {
		c.Err = NewAppError("Delete job error", err.Error(), 500)
		return
	}
	utils.WriteJson(w, "Delete Ok")
}

func schedUpdate(c *Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	timer := r.PostFormValue("cron")
	host := r.PostFormValue("host")
	types := r.PostFormValue("type")
	if id == "" || timer == "" || host == "" || types == "" {
		c.Err = NewAppError("Param error", "Parameters cannot be empty", 500)
		return
	}

	sched, err := service.TimerCheckFormat(timer)
	if err != nil {
		c.Err = NewAppError("timer format error", err.Error(), 500)
		return
	}

	schedType, err := strconv.Atoi(types)
	if err != nil {
		c.Err = NewAppError("Type Atoi converse error", err.Error(), 500)
		return
	}

	timeStamp := time.Now()
	schedItem := service.NewCron(bson.ObjectIdHex(id), -1, schedType, timer, host, timeStamp, timeStamp)
	info := &service.HostInfo{
				Id:id,
				Host:host,
				Type:schedType,
			}

	job := service.NewJob(id, service.START, sched, schedItem, info)
	err = Srv.Service.Sched.Update(job)
	if err != nil {
		c.Err = NewAppError("Update job error", err.Error(), 500)
		return
	}
	utils.WriteJson(w, "Update Ok")
}

func schedSearch(c *Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		c.Err = NewAppError("Param error", "Parameters cannot be empty", 500)
		return
	}

	item, err := Srv.Service.Sched.Search(id)
	if err != nil {
		c.Err = NewAppError("sched search err:", err.Error(), 500)
	} else {
		utils.WriteJson(w, item)
	}
}

func schedSearchAll(c *Context, w http.ResponseWriter, r *http.Request) {
	item, err := Srv.Service.Sched.Crons.Traversal()
	if err != nil {
		c.Err = NewAppError("sched search err:", err.Error(), 500)
		return
	} else {
		utils.WriteJson(w, item)
	}
}

func schedDisable(c *Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		c.Err = NewAppError("Param error", "Parameters cannot be empty", 500)
		return
	}

	err := Srv.Service.Sched.Disable(id)
	if err != nil {
		c.Err = NewAppError("Disable job error", err.Error(), 500)
		return
	}
	utils.WriteJson(w, "Set disable ok")
}

func schedEnable(c *Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		c.Err = NewAppError("Param error", "Parameters cannot be empty", 500)
		return
	}

	err := Srv.Service.Sched.Enable(id)
	if err != nil {
		c.Err = NewAppError("Enable job error", err.Error(), 500)
		return
	}
	utils.WriteJson(w, "Set enable ok")
}

func schedGetType(types string) (int, error) {
	switch types {
	case "0":
		return 0, nil //Full backup
	case "1":
		return 1, nil //Incremental backup
	case "2":
		return 2, nil //compress
	default:
		return -1, errors.New("Type out of range")
	}
}

package api

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/SpruceX/potato/utils"
)

func initTopology(r *mux.Router) {
	sr := r.PathPrefix("/dbdata").Subrouter()
	sr.Handle("/topology", ApiAppHandler(findTopologyResults)).Methods("GET")
	sr.Handle("/db", AnonymousApiHandler(findDBResults)).Methods("GET")
	sr.Handle("/dbtable", AnonymousApiHandler(findDBTableResults)).Methods("GET")
	sr.Handle("/dbtabledesc", AnonymousApiHandler(findDBTableDescResults)).Methods("GET")
}

func findTopologyResults(context *Context, w http.ResponseWriter, req *http.Request) {
	if results, err := Srv.Store.Topology.GetTopology(); err != nil {
		context.Err = NewAppError("find topology results", err.Error(), 500)
		return
	} else {
		utils.WriteJson(w, results)
	}
}

func findDBResults(context *Context, w http.ResponseWriter, req *http.Request) {
	hostname := req.FormValue("hostname")
	hostname = strings.TrimSpace(hostname)
	if hostname == ""{
		context.Err = NewAppError("Param error", "host name can not be empty", 500)
		return
	}
	dbs, err := Srv.Service.Topology.GetDatabase(hostname)
	if err != nil {
		context.Err = NewAppError("find database error", err.Error(), 500)
	} else {
		utils.WriteJson(w, dbs)
	}
}

func findDBTableResults(context *Context, w http.ResponseWriter, req *http.Request) {
	hostname := req.FormValue("hostname")
	hostname = strings.TrimSpace(hostname)
	if hostname == ""{
		context.Err = NewAppError("Param error", "host name can not be empty", 500)
		return
	}
	dbname := req.FormValue("dbname")
	dbname = strings.TrimSpace(dbname)
	if hostname == ""{
		context.Err = NewAppError("Param error", "database name can not be empty", 500)
		return
	}
	dbs, err := Srv.Service.Topology.GetDatabaseTable(hostname, dbname)
	if err != nil {
		context.Err = NewAppError("find database table error", err.Error(), 500)
	} else {
		utils.WriteJson(w, dbs)
	}
}

func findDBTableDescResults(context *Context, w http.ResponseWriter, req *http.Request) {
	hostname := req.FormValue("hostname")
	hostname = strings.TrimSpace(hostname)
	if hostname == ""{
		context.Err = NewAppError("Param error", "host name can not be empty", 500)
		return
	}
	dbname := req.FormValue("dbname")
	dbname = strings.TrimSpace(dbname)
	if hostname == ""{
		context.Err = NewAppError("Param error", "database name can not be empty", 500)
		return
	}
	tablename := req.FormValue("tablename")
	tablename = strings.TrimSpace(tablename)
	if hostname == ""{
		context.Err = NewAppError("Param error", "database name can not be empty", 500)
		return
	}
	dbs, err := Srv.Service.Topology.GetDatabaseTableDesc(hostname, dbname, tablename)
	if err != nil {
		context.Err = NewAppError("find database table desc error", err.Error(), 500)
	} else {
		utils.WriteJson(w, dbs)
	}
}

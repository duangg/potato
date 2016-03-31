package api

import (
	"net/http"
	"strconv"
	"strings"


	"gopkg.in/mgo.v2/bson"
	"github.com/gorilla/mux"
	"github.com/SpruceX/potato/utils"
)

func initJobResults(r *mux.Router) {
	sr := r.PathPrefix("/jobresults").Subrouter()
	sr.Handle("/", ApiAppHandler(findJobResults)).Methods("GET")
	sr.Handle("/last", ApiAppHandler(findLastJobResults)).Methods("GET")
	sr.Handle("/page", ApiAppHandler(findPageJobResults)).Methods("GET")
	sr.Handle("/error", ApiAppHandler(findErrorJobResults)).Methods("GET")
	sr.Handle("/run", ApiAppHandler(findRunJobResults)).Methods("GET")
	sr.Handle("/{id}/dismiss", ApiAppHandler(dismissErrorJobResult)).Methods("POST")
	sr.Handle("/abort", ApiAppHandler(abortRunningJobResult)).Methods("POST")
}

func findJobResults(context *Context, w http.ResponseWriter, req *http.Request) {
	cronId := req.FormValue("cronid")
	if cronId == "" {
		context.Err = NewAppError("find job results", "missing parameter cronid", 400)
		return
	}
	if ok := bson.IsObjectIdHex(cronId); !ok {
		context.Err = NewAppError("find job results", "wrong cron id style", 500)
		return
	}
	if results, err := Srv.Store.JobResult.FindJobResultByCronId(cronId); err != nil {
		context.Err = NewAppError("find job results", err.Error(), 500)
		return
	} else {
		if results != nil {
			utils.WriteJson(w, results)
		} else {
			utils.WriteJson(w, EmptyArray)
		}
	}
}

func findPageJobResults(context *Context, w http.ResponseWriter, req *http.Request) {
	cronId := req.FormValue("cronid")
	if cronId == "" {
		context.Err = NewAppError("find page result", "missing parameter cronid", 400)
		return
	}
	if ok := bson.IsObjectIdHex(cronId); !ok {
		context.Err = NewAppError("find page results", "wrong cron id style", 500)
		return
	}
	pageNumStr := req.FormValue("pn")
	if pageNumStr == "" {
		context.Err = NewAppError("find page result", "missing parameter page number", 400)
		return
	}
	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil {
		context.Err = NewAppError("find page result parsing parameter", err.Error(), 400)
		return
	}
	if result, err := Srv.Store.JobResult.FindPageJobResultByCronId(cronId, pageNum, PageSize); err != nil {
		context.Err = NewAppError("find last result", err.Error(), 500)
		return
	} else {
		utils.WriteJson(w, result)
	}
}

func findLastJobResults(context *Context, w http.ResponseWriter, req *http.Request) {
	cronId := req.FormValue("cronid")
	if cronId == "" {
		context.Err = NewAppError("find last result", "missing parameter cronid", 400)
		return
	}
	if ok := bson.IsObjectIdHex(cronId); !ok {
		context.Err = NewAppError("find last results", "wrong cron id style", 500)
		return
	}
	numStr := req.FormValue("number")
	if numStr == "" {
		context.Err = NewAppError("find last result", "missing parameter page number", 400)
		return
	}
	num, err := strconv.Atoi(numStr)
	if err != nil {
		context.Err = NewAppError("find last result parsing parameter", err.Error(), 400)
		return
	}
	if result, err := Srv.Store.JobResult.FindLastJobResultByCronId(cronId, num); err != nil {
		context.Err = NewAppError("find last result", err.Error(), 500)
		return
	} else {
		utils.WriteJson(w, result)
	}
}

func findErrorJobResults(context *Context, w http.ResponseWriter, req *http.Request) {
	if result, err := Srv.Store.JobResult.FindAllBackupErrorResult(); err != nil {
		context.Err = NewAppError("find all error result", err.Error(), 500)
		return
	} else {
		utils.WriteJson(w, result)
	}
}

func findRunJobResults(context *Context, w http.ResponseWriter, req *http.Request) {
	if result, err := Srv.Store.JobResult.FindAllBackupRunningResult(); err != nil {
		context.Err = NewAppError("find all running result", err.Error(), 500)
		return
	} else {
		utils.WriteJson(w, result)
	}
}

func dismissErrorJobResult(context *Context, w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	id := vars["id"]
	if id == "" {
		context.Err = NewAppError("Param error", "Parameters id cannot be empty", 500)
		return
	}
	if ok := bson.IsObjectIdHex(id); !ok {
		context.Err = NewAppError("Param error", "wrong job result id style", 500)
		return
	}
	err := Srv.Store.JobResult.DismissBackupErrorResult(id)
	if err != nil {
		context.Err = NewAppError("dismiss error job result", err.Error(), 500)
		return
	}
	utils.WriteJson(w, "Dismiss ok")
}

func abortRunningJobResult(context *Context, w http.ResponseWriter, req *http.Request) {
	id := req.FormValue("id")
	if id == "" {
		context.Err = NewAppError("Param error", "Parameters id cannot be empty", 500)
		return
	}
	if ok := bson.IsObjectIdHex(id); !ok {
		context.Err = NewAppError("Param error", "wrong job result id style", 500)
		return
	}
	errValue := req.FormValue("error")
	errValue = strings.TrimSpace(errValue)
	if errValue == "" {
		context.Err = NewAppError("Param error", "the error info is empyt", 500)
		return
	}
	err, jobid := Srv.Store.JobResult.AbortRunningBackupResult(id, errValue)
	if err != nil {
		context.Err = NewAppError("abort job result", err.Error(), 500)
		return
	}
	Srv.Service.Sched.UpdateJobStatus(jobid, 2)
	utils.WriteJson(w, "abort ok")
}

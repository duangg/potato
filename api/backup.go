package api

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/SpruceX/potato/utils"
)

func initBackup(r *mux.Router) {
	sr := r.PathPrefix("/backup").Subrouter()
	sr.Handle("/", ApiAppHandler(getBackupFilePage)).Methods("GET")
	sr.Handle("/getallbackup", ApiAppHandler(getAllBackupFile)).Methods("GET")
	sr.Handle("/info", ApiAppHandler(getAllBackupInfo)).Methods("GET")
	sr.Handle("/{name}", ApiAppHandler(getBackupFile)).Methods("GET")
}

func getBackupFilePage(c *Context, w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	pageNumStr := r.FormValue("pn")
	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil {
		c.Err = NewAppError("gethosts parsing parameter", err.Error(), 400)
		return
	}
	find, err := Srv.Store.BackupFileResult.Find(name, pageNum, PageSize)
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

func getBackupFile(c *Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	result, err := Srv.Store.BackupFileResult.FindBackupFileByServerName(name)
	if err != nil {
		c.Err = NewAppError("find backup file by server name:", err.Error(), 500)
		return
	}
	utils.WriteJson(w, result)
}

func getAllBackupFile(c *Context, w http.ResponseWriter, r *http.Request) {
	find, err := Srv.Store.BackupFileResult.GetAllBackupFile()
	if err != nil {
		c.Err = NewAppError("get all backup file", err.Error(), 500)
		return
	}
	if find == nil {
		utils.WriteJson(w, EmptyArray)
	} else {
		utils.WriteJson(w, find)
	}
}

func getAllBackupInfo(c *Context, w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	sum, err := Srv.Store.BackupFileResult.GetAllBackupFileCount(name)
	if err != nil {
		c.Err = NewAppError("get all backup file number", err.Error(), 500)
		return
	}
	var infoArray []int = make([]int, 2)
	infoArray[0] = sum
	infoArray[1] = PageSize
	utils.WriteJson(w, infoArray)
}

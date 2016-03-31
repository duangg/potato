package api

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/SpruceX/potato/utils"
)

func initConfig(r *mux.Router) {
	sr := r.PathPrefix("/config").Subrouter()
	sr.Handle("/", ApiAppHandler(getConfig))
}

func getConfig(c *Context, w http.ResponseWriter, r *http.Request) {
	err := utils.WriteJson(w, utils.Cfg)
	if err != nil {
		c.Err = NewAppError("getConfig", err.Error(), 10001)
	}
}

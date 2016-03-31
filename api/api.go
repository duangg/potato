package api

func InitApi() {
	r := Srv.Router.PathPrefix("/api/v1").Subrouter()
	initAdmin(r)
	initUser(r)
	initConfig(r)
	initBackup(r)
	initHost(r)
	initSched(r)
	initJobResults(r)
	initTopology(r)
}

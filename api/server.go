package api

import (
	"net/http"
	"log"

	"github.com/braintree/manners"
	"github.com/gorilla/mux"

	"github.com/SpruceX/potato/service"
	"github.com/SpruceX/potato/store"
	"github.com/SpruceX/potato/utils"
)

type Server struct {
	Router  *mux.Router
	Service *service.Service
	//       Sched   *schedule.Schedule
	Store *store.MongoStore
}

var Srv *Server

func NewServer() {
	log.Printf("Server is initializing...")
	Srv = &Server{}
	Srv.Store = store.Store
	Srv.Service = service.AllService
	Srv.Router = mux.NewRouter()
	Srv.Router.NotFoundHandler = http.NotFoundHandler()
}

func StartServer() {
	log.Printf("Server is starting...")
	log.Printf("Server is listening on " + utils.Cfg.IP + ":" + utils.Cfg.Port)
	listening := utils.Cfg.IP + ":" + utils.Cfg.Port

	go func() {
		manners.ListenAndServe(listening, Srv.Router)
	}()
}

func StopServer() {
	log.Printf("Server is stopping...")
	manners.Close()
	log.Printf("Server is stopped")
}

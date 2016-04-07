package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"log"

	"github.com/SpruceX/potato/api"
	"github.com/SpruceX/potato/service"
	"github.com/SpruceX/potato/store"
	"github.com/SpruceX/potato/utils"
	"github.com/SpruceX/potato/web"
)

var (
	configFolder *string = flag.String("config-folder", "./config", "folder where to find config files")
)

const (
	APP_VERSION = "0.3.1"
)

func main() {
	log.Printf("Application version %s", APP_VERSION)
	flag.Parse()
	utils.LoadConfig(*configFolder)
	store.Init()
	service.Init()
	service.AllService.Sched.Start()
	api.NewServer()
	api.InitApi()
	web.InitWeb()
	api.StartServer()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-c
	log.Printf("Stopping signal recieved, stopping server")
	api.StopServer()
}

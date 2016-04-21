package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	ServiceConfigFileName = "service.json"
)

type ServiceSettings struct {
	IP            string `json:"ip"`
	Port          string `json:"port"`
	SiteName      string `json:"site_name"`
	Mongo         string `json:"mongo"`
	SshPrivateKey string `json:"sshprivatekey"`
	SsoVerifyUrl  string `json:"ssoverifyurl"`
	RedirectUrl   string `json:"redirecturl"`
	SystemID	  string `json:"systemid"`
}

type Config struct {
	*ServiceSettings
}

var Cfg *Config

func LoadConfig(configFolderPath string) {
	if _, err := os.Stat(configFolderPath); err != nil {
		panic("Config folder not exists " + err.Error())
	}
	serviceConfigFilePath := filepath.Join(configFolderPath, ServiceConfigFileName)
	if _, err := os.Stat(serviceConfigFilePath); err != nil {
		panic("Service config file " + serviceConfigFilePath + " does not exist")
	}
	serviceConfigFile, err := os.Open(serviceConfigFilePath)
	if err != nil {
		panic("error opening file " + serviceConfigFilePath + ", error is " + err.Error())
	}
	defer serviceConfigFile.Close()

	Cfg = &Config{}
	var serviceSettings ServiceSettings
	err = json.NewDecoder(serviceConfigFile).Decode(&serviceSettings)
	if err != nil {
		panic("error parsing config file " + serviceConfigFilePath + " error is " + err.Error())
	}
	Cfg.ServiceSettings = &serviceSettings
}

func FindDir(path string) string {
	return path
}

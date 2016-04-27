package service

import (
	"fmt"
	"os/exec"
	"time"
	"strings"
	"log"

	"github.com/SpruceX/easyssh"
	"github.com/SpruceX/potato/models"
	"github.com/SpruceX/potato/store"
)

const (
	ShellFile = "topology.sh"
	IsSlave = "sh %s isslave %s \"%s\" %s %s %s %s %s %s "
	GetSlaveMaster = "sh %s getslavemaster %s \"%s\" %s %s %s %s %s %s "
	CheckSlaveStatus = "sh %s checkslavestatus %s \"%s\" %s %s %s %s %s %s "
	GetMasterSlave = "sh %s getmasterslave %s \"%s\" %s %s %s %s %s %s "
	GetDatabase = "sh %s getdatabase %s \"%s\" %s %s %s %s %s %s "
	GetDBTable = "sh %s getdbtable %s \"%s\" %s %s %s %s %s %s "
	GetDBTableDesc = "sh %s getdbtabledesc %s \"%s\" %s %s %s %s %s %s "
	FindFile = "sudo find %s -name %s"
)
const (
	RefreshTopology = 20
)

const (
	MasterSlaveOK = "connect"
	MasterSlaveError="disconnect"
)
type TopologyService struct {
	mongoStore *store.MongoStore
}

func (s TopologyService) isslave(ssh *easyssh.MakeConfig, host *models.Host) (bool, error) {
	cmdPara := fmt.Sprintf(IsSlave, host.BackupPath+"/"+ShellFile, host.DBUser, host.DBPassword,
		host.DBPort, host.DBHost, host.DBSocket, host.DBMyCnf, "", "")
	result, err := ssh.Run(cmdPara)
	if err != nil {
		return false, err
	}
	result = strings.TrimSpace(result)
	result = strings.Trim(result, "\n")
	var ret bool
	if result == "1" {
		ret = true
	} else {
		ret = false
	}
	return ret, err
}

func (s TopologyService) checkslavestatus(ssh *easyssh.MakeConfig, host *models.Host) (string, error) {
	cmdPara := fmt.Sprintf(CheckSlaveStatus, host.BackupPath+"/"+ShellFile, host.DBUser, host.DBPassword,
		host.DBPort, host.DBHost, host.DBSocket, host.DBMyCnf, "", "")
	result, err := ssh.Run(cmdPara)
	if err != nil {
		return MasterSlaveError, err
	}
	result = strings.TrimSpace(result)
	result = strings.Trim(result, "\n")
	var ret string
	if result == "1" {
		ret = MasterSlaveOK
	} else {
		ret = MasterSlaveError
	}
	return ret, err
}

func (s TopologyService) getmaster(ssh *easyssh.MakeConfig, host *models.Host) (string, error) {
	cmdPara := fmt.Sprintf(GetSlaveMaster, host.BackupPath+"/"+ShellFile, host.DBUser, host.DBPassword,
		host.DBPort, host.DBHost, host.DBSocket, host.DBMyCnf, "", "")
	result, err := ssh.Run(cmdPara)
	if err != nil {
		return "", err
	}
	result = strings.TrimSpace(result)
	result = strings.Trim(result, "\n")
	return result, err
}

func (s TopologyService) gethost(ip string) (*models.Host, error) {
	host, err := s.mongoStore.Hosts.FindHostByIp(ip)
	return &host, err
}

func (s TopologyService) insertdblink(master, host *models.Host, nowtype string) error {
	return s.mongoStore.Topology.InsertDBLink(master, host, nowtype)
}

func (s TopologyService) CheckTopology(host *models.Host) {
	cmd := exec.Command("/bin/sh", "-c", "ssh-add")
	cmd.Output()

	ssh := makeSshClient(host)
	err := s.doScpShellFile(host, ssh)
	if err != nil {
		log.Println(fmt.Sprintf("failed to scp file to %s in check topology job, error info-%s", host.Name, err.Error()))
		return
	}

	result, err := s.isslave(ssh, host)
	if err != nil {
		log.Println(fmt.Sprintf("failed to check slave for host %s, error info-%s", host.Name, err.Error()))
		return
	}
	if(result) {
		masterip, err := s.getmaster(ssh, host)
		if err != nil {
			log.Println(fmt.Sprintf("failed to get slave's master ip for host %s, error info-%s", host.Name, err.Error()))
			return
		}
		if masterip == "" {
			log.Println(fmt.Sprintf("the slave's master ip is empty for host %s", host.Name))
			return
		}
		masterhost, err := s.gethost(masterip)
		if err != nil {
			log.Println(fmt.Sprintf("failed to get host by ip:%s, error info-%s", masterip, err.Error()))
			return
		}
		ok, err := s.checkslavestatus(ssh, host)
		if err != nil {
			log.Println(fmt.Sprintf("failed to get slave status for host %s, error info-%s", host.Name, err.Error()))
			return
		}
		s.insertdblink(masterhost, host, ok)
	} else {
	}
}

func (s TopologyService) doScpShellFile(host *models.Host, ssh *easyssh.MakeConfig) error {
	_, errTest := ssh.Run("ls")
	if errTest != nil {
		return  errTest
	}

	cmdPara := fmt.Sprintf("%s@%s:%s/", host.UserName, host.IP, host.BackupPath)
	cmdParaPort := fmt.Sprintf("-P%s", host.SshPort)
	cmd := exec.Command("scp", cmdParaPort, "./script/" + ShellFile, cmdPara)
	_, err := cmd.Output()
	return err
}

func (s TopologyService) RefreshTopology() {
	hosts, err := s.mongoStore.Hosts.GetAllHosts()
	if err != nil {
		str := fmt.Sprintf("failed to refresh topology, error info-%s", err.Error())
		log.Println(str)
		return
	}
	for _, host := range hosts {
		s.CheckTopology(&host)
	}
}

func (s TopologyService) TimeRefreshTopology() {
	go func () {
		timeRefresh := time.NewTicker(RefreshTopology * time.Second)
		for {
			select {
			case <-timeRefresh.C:
				s.RefreshTopology()
			}
		}
	} ()
}

func (s TopologyService) getdatabase(ssh *easyssh.MakeConfig, host *models.Host) ([]models.Database ,error) {
	cmdPara := fmt.Sprintf(GetDatabase, host.BackupPath+"/"+ShellFile, host.DBUser, host.DBPassword,
		host.DBPort, host.DBHost, host.DBSocket, host.DBMyCnf, "", "")
	result, err := ssh.Run(cmdPara)
	if err != nil {
		return nil, err
	}
	var dbs []models.Database
	result = strings.TrimSpace(result)
	dbnames := strings.Split(result, "\n")
	for _, dbname := range dbnames  {
		var db models.Database
		db.Name = dbname
		dbs = append(dbs, db)
	}
	return dbs, nil
}

func (s TopologyService) getdbtable(ssh *easyssh.MakeConfig, host *models.Host, dbname string) ([]models.DatabaseTable ,error) {
	cmdPara := fmt.Sprintf(GetDBTable, host.BackupPath+"/"+ShellFile, host.DBUser, host.DBPassword,
		host.DBPort, host.DBHost, host.DBSocket, host.DBMyCnf, dbname, "")
	result, err := ssh.Run(cmdPara)
	if err != nil {
		return nil, err
	}

	var tables []models.DatabaseTable
	result = strings.TrimSpace(result)
	tablenames := strings.Split(result, "\n")
	for _, tablename := range tablenames  {
		var table models.DatabaseTable
		table.Name = tablename
		result, err := s.getdbtabledesc(ssh, host, dbname,tablename)
		if err != nil {
			log.Println(fmt.Sprintf("failed to get database table desc for host %s, error info-%s", host.Name, err.Error()))
			return nil, err
		}
		table.Descs = result
		tables = append(tables, table)
	}
	return tables, nil
}

func (s TopologyService) getdbtabledesc(ssh *easyssh.MakeConfig, host *models.Host, dbname, tablename string) ([]models.DatabaseTableDesc ,error) {
	cmdPara := fmt.Sprintf(GetDBTableDesc, host.BackupPath+"/"+ShellFile, host.DBUser, host.DBPassword,
		host.DBPort, host.DBHost, host.DBSocket, host.DBMyCnf, dbname, tablename)
	result, err := ssh.Run(cmdPara)
	if err != nil {
		return nil, err
	}
	var tabledescs []models.DatabaseTableDesc
	result = strings.TrimSpace(result)
	fileds := strings.Split(result, "\n")
	for index, filed := range fileds  {
		if index == 0 {
			continue
		}
		result := strings.Split(filed, "\t")
		var tabledesc models.DatabaseTableDesc
		resultlen := len(result)
		if resultlen >= 1 {
			tabledesc.Field = result[0]
		}
		if resultlen >= 2 {
			tabledesc.Type = result[1]
		}
		if resultlen >= 3 {
			tabledesc.IsNull = result[2]
		}
		if resultlen >= 4 {
			tabledesc.Key = result[3]
		}
		if resultlen >= 5 {
			tabledesc.Default = result[4]
		}
		if resultlen >= 6 {
			tabledesc.Comment = result[5]
		}
		tabledescs = append(tabledescs, tabledesc)
	}
	return tabledescs, nil
}

func (s TopologyService) GetDatabase(hostname string) ([]models.Database, error){
	cmd := exec.Command("/bin/sh", "-c", "ssh-add")
	cmd.Output()
	host, err := s.mongoStore.Hosts.FindHostByName(hostname)
	if err != nil {
		log.Println(fmt.Sprintf("failed to find host by name-%s, error info-%s", hostname, err.Error()))
		return nil, err
	}

	ssh := makeSshClient(&host)
	err = s.doScpShellFile(&host, ssh)
	if err != nil {
		log.Println(fmt.Sprintf("failed to scp file to %s in get database job, error info-%s", host.Name, err.Error()))
		return nil, err
	}

	result, err := s.getdatabase(ssh, &host)
	if err != nil {
		log.Println(fmt.Sprintf("failed to get database for host %s, error info-%s", host.Name, err.Error()))
		return nil, err
	}
	return result,nil
}

func (s TopologyService) GetDatabaseTable(hostname, dbname string) ([]models.DatabaseTable, error){
	cmd := exec.Command("/bin/sh", "-c", "ssh-add")
	cmd.Output()
	host, err := s.mongoStore.Hosts.FindHostByName(hostname)
	if err != nil {
		log.Println(fmt.Sprintf("failed to find host by name-%s, error info-%s", hostname, err.Error()))
		return nil, err
	}

	ssh := makeSshClient(&host)
	err = s.doScpShellFile(&host, ssh)
	if err != nil {
		log.Println(fmt.Sprintf("failed to scp file to %s in get database table job, error info-%s", host.Name, err.Error()))
		return nil, err
	}

	result, err := s.getdbtable(ssh, &host, dbname)
	if err != nil {
		log.Println(fmt.Sprintf("failed to get database tables for host %s, error info-%s", host.Name, err.Error()))
		return nil, err
	}
	return result,nil
}

func (s TopologyService) GetDatabaseTableDesc(hostname, dbname, tablename string) ([]models.DatabaseTableDesc, error){
	cmd := exec.Command("/bin/sh", "-c", "ssh-add")
	cmd.Output()
	host, err := s.mongoStore.Hosts.FindHostByName(hostname)
	if err != nil {
		log.Println(fmt.Sprintf("failed to find host by name-%s, error info-%s", hostname, err.Error()))
		return nil, err
	}

	ssh := makeSshClient(&host)
	err = s.doScpShellFile(&host, ssh)
	if err != nil {
		log.Println(fmt.Sprintf("failed to scp file to %s in get database table desc job, error info-%s", host.Name, err.Error()))
		return nil, err
	}

	result, err := s.getdbtabledesc(ssh, &host, dbname,tablename)
	if err != nil {
		log.Println(fmt.Sprintf("failed to get database table desc for host %s, error info-%s", host.Name, err.Error()))
		return nil, err
	}
	return result,nil
}

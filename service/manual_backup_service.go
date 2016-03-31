package service

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"log"

	"gopkg.in/mgo.v2/bson"
	"github.com/hypersleep/easyssh"
	"github.com/SpruceX/potato/models"
	"github.com/SpruceX/potato/store"
	"github.com/SpruceX/potato/utils"
)

type ManualBackupService struct {
	mongoStore *store.MongoStore
}

const (
	CreateBackup_Directory = "sudo mkdir -p %s"
	//nohup sh /BackupPath/xx.sh full_backup DBUser DBPassword DBPort BackupPath DBHost DBSocket DBMyCnf
	ShellFullBackupContent = "sudo nohup sh %s full_backup %s \"%s\" %s %s %s %s %s & "
	ShellIncBackupContent  = "sudo nohup sh %s inc_backup %s \"%s\" %s %s %s %s %s & "
	//cat /BackupPath/log
	GetResultOK      = "sudo cat %s/log | grep -c \"completed OK\" "
	GetResultError   = "sudo cat %s/log | grep -c \"Error\" "
	GetResultErrInfo = "sudo cat %s/log | grep  \"Error\" | awk -F \"Error\" '{print $2}' | head -n 10"
	GetBackupFile    = "sudo cat %s/log | grep \"Backup created in directory\" |  awk -F \"'\" '{print $2}'"
	TimeStyle        = "2006-01-02 12:04:05"
	//du -s 备份结果完整路径
	GetBackupFileSize      = "sudo du -s %s"
	GetBackupLog           = "sudo tail -n 10 %s/log"
	ShellExecBackupShell   = "exbackup.sh"
	ClearBackupLog         = "sudo echo \"\" > %s/log"
	//sh /BackupPath/xx.sh full_backup DBUser DBPassword DBPort BackupPath DBHost DBSocket DBMyCnf
	CompressFile           = "sudo sh %s scp_lftp %s \"%s\" %s %s %s %s %s"
	GetCompressResult      = "sudo cat %s/compresslog"
	GetCompressResultError = "sudo cat %s/compresslog | grep \"ERROR\" | head -n 1 | awk -F\":\" '{print $2}'"
	XtrabackupNotFind = "sudo cat %s/log | grep \"innobackupex: not found\" "
)

const (
	Idle = iota
	Running
	Error
	Finish
)

var ServerStatus map[string]int

func makeSshClient(b *models.Host) *easyssh.MakeConfig {
	return &easyssh.MakeConfig{
		User:     b.UserName,
		Password: "",
		Server:   b.IP,
		Port:     b.SshPort,
		Key:      utils.Cfg.SshPrivateKey,
	}
}

func GetTime() time.Time {
	return time.Now()
}

func (s ManualBackupService) getBackupLog(ssh *easyssh.MakeConfig, host *models.Host) (string, error) {
	output, err := ssh.Run(fmt.Sprintf(GetBackupLog, host.BackupPath))
	return output, err
}

func (s ManualBackupService) doCheckBackupResult(ssh *easyssh.MakeConfig, jobId string, host *models.Host, startTime time.Time, backupType int) {
	backupTypeStr := ""
	if backupType == FULL {
		backupTypeStr = "full"
	} else {
		backupTypeStr = "inc"
	}

	for {
		output, logErr := s.getBackupLog(ssh, host)
		if logErr != nil {

		}
		s.mongoStore.JobResult.SaveJobResult(host.Name, jobId, "", startTime, startTime, output, Running, backupType)

		resultGetResultOK, errGetResultOK := ssh.Run(fmt.Sprintf(GetResultOK, host.BackupPath))
		if errGetResultOK != nil {

		} else {
			res, _ := strconv.Atoi(strings.Trim(resultGetResultOK, "\n"))
			if res >= 2 {
				//file full path  /dbdata/20151225/ful/2015-12-25_16-56-32
				getBackupFile, errGetBackupFile := ssh.Run(fmt.Sprintf(GetBackupFile, host.BackupPath))
				if errGetBackupFile == nil {
					getBackupFile = strings.Trim(getBackupFile, "\n")
					if getBackupFile == "" {
						continue
					}
					pos := strings.LastIndex(getBackupFile, "/")
					filename := getBackupFile[pos+1:]
					filedate := strings.Split(filename, "_")

					//get backup file size
					output, _ := ssh.Run(fmt.Sprintf(GetBackupFileSize, getBackupFile))
					filesizeinfo := strings.Split(output, "\t")

					//wait sometime and get backup log content
					time.Sleep(10)
					backuplog, _ := s.getBackupLog(ssh, host)

					//get end time
					endttime := GetTime()

					//write backup file db and job result db
					s.mongoStore.BackupFileResult.SaveBackupFileResult(host.Name, filedate[0], backupType, jobId, getBackupFile, filesizeinfo[0], endttime)
					s.mongoStore.JobResult.SaveJobResult(host.Name, jobId, "", startTime, endttime, backuplog, Finish, backupType)
					log.Printf("%s backup succeful for %s in job %s, backup file:%s", backupTypeStr, host.Name, jobId, getBackupFile)

					//update server status and notify crons status
					ServerStatus[host.Name] = Idle
					AllService.Sched.UpdateJobStatus(jobId, Idle)
				} else {
					log.Printf("failed to get %s backup's file for %s in job %s, error info-%s", backupTypeStr, host.Name, jobId, errGetBackupFile.Error())
					ServerStatus[host.Name] = Error
					AllService.Sched.UpdateJobStatus(jobId, Error)
				}
				break
			}
		}

		_, errNotFindError := ssh.Run(fmt.Sprintf(XtrabackupNotFind, host.BackupPath))
		if errNotFindError != nil {

		} else {
			strErr := "innobackupex is not found"
			errStr := fmt.Sprintf("failed to %s backup for %s in job %s, error info-%s", backupTypeStr, host.Name, jobId, strErr)
			s.mongoStore.JobResult.SaveJobResult(host.Name, jobId, errStr, startTime, GetTime(), strErr, Error, backupType)
			log.Printf(errStr)
			ServerStatus[host.Name] = Error
			AllService.Sched.UpdateJobStatus(jobId, Error)
			break
		}

		resultGetResultError, errGetResultError := ssh.Run(fmt.Sprintf(GetResultError, host.BackupPath))
		if errGetResultError != nil {

		} else {
			resError, _ := strconv.Atoi(strings.Trim(resultGetResultError, "\n"))
			if resError >= 1 {
				//get error information
				getErrInfo, errGetErrInfo := ssh.Run(fmt.Sprintf(GetResultErrInfo, host.BackupPath))
				if errGetErrInfo == nil {
					//wait sometime and get backup log content
					time.Sleep(10)
					backuplog, _ := s.getBackupLog(ssh, host)

					errStr := fmt.Sprintf("failed to %s backup for %s in job %s, error info-%s", backupTypeStr, host.Name, jobId, getErrInfo)
					s.mongoStore.JobResult.SaveJobResult(host.Name, jobId, getErrInfo, startTime, GetTime(), backuplog, Error, backupType)
					log.Printf(errStr)
					ServerStatus[host.Name] = Error
					AllService.Sched.UpdateJobStatus(jobId, Error)
				}
				break
			}
		}

		time.Sleep(time.Second)
	}
}

func (s ManualBackupService) doScpShellFile(jobId string, host *models.Host, exectype int, starttime time.Time, ssh *easyssh.MakeConfig) error {
	_, errTest := ssh.Run("ls")
	if errTest != nil {
		return  errTest
	}

	cmdPara := fmt.Sprintf("%s@%s:%s/", host.UserName, host.IP, host.BackupPath)
	cmdParaPort := fmt.Sprintf("-P%s", host.SshPort)
	cmd := exec.Command("scp", cmdParaPort, "./script/"+utils.Cfg.Shellname, cmdPara)
	_, err := cmd.Output()
	return err
}

func (s ManualBackupService) doExecShellBackup(ssh *easyssh.MakeConfig, jobId string, host *models.Host, backupType int) error {
	shellStr := ""
	if backupType == FULL {
		shellStr = ShellFullBackupContent
	} else {
		shellStr = ShellIncBackupContent
	}

	serverPara := fmt.Sprintf("%s@%s", host.UserName, host.IP)
	cmdPara := fmt.Sprintf(shellStr, host.BackupPath+"/"+utils.Cfg.Shellname, host.DBUser,
		host.DBPassword, host.DBPort, host.BackupPath, host.DBHost, host.DBSocket, host.DBMyCnf)
	cmdPort := fmt.Sprintf("%s", host.SshPort)
	cmd := exec.Command("sh", "./script/" + ShellExecBackupShell, cmdPort, serverPara, cmdPara)
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("cmdPara-%s \n", cmdPara)
		log.Printf("exec backup shell,result-%s", output)
	}
	return err
}

func (s ManualBackupService) doBackup(jobId string, host *models.Host, backupType int) {
	backupTypeStr := ""
	if backupType == FULL {
		backupTypeStr = "full"
	} else {
		backupTypeStr = "inc"
	}

	defer func() {
		if err := recover(); err != nil {
			log.Printf("full backup find panic", err)
		}
	}()

	startTime := GetTime()
	ServerStatus[host.Name] = Running
	AllService.Sched.UpdateJobStatus(jobId, Running)
	s.mongoStore.JobResult.SaveJobResult(host.Name, jobId, "", startTime, startTime, "", Running, backupType)

	cmd := exec.Command("/bin/sh", "-c", "ssh-add")
	cmd.Output()
	ssh := makeSshClient(host)
	err := s.doScpShellFile(jobId, host, backupType, startTime, ssh)
	if err != nil {
		ServerStatus[host.Name] = Error
		AllService.Sched.UpdateJobStatus(jobId, Error)
		str := fmt.Sprintf("failed to scp file to %s in %s backup job %s, error info-%s", backupTypeStr, host.Name, jobId, err.Error())
		s.mongoStore.JobResult.SaveJobResult(host.Name, jobId, str, startTime, GetTime(), "", Error, backupType)
		log.Printf(str)
		return
	}

	//clear backup log file
	_, err = ssh.Run(fmt.Sprintf(ClearBackupLog, host.BackupPath))
	if err != nil {
		str := fmt.Sprintf("failed to clear %s backup's log for %s in job %s, error info-%s", backupTypeStr, host.Name, jobId, err.Error())
		log.Printf(str)
		ServerStatus[host.Name] = Error
		s.mongoStore.JobResult.SaveJobResult(host.Name, jobId, str, startTime, GetTime(), "", Error, backupType)
		AllService.Sched.UpdateJobStatus(jobId, Error)
		return
	}

	//do backup and set backupover to true when full backup is over whatever succeed or fail
	err = s.doExecShellBackup(ssh, jobId, host, backupType)
	if err != nil {
		ServerStatus[host.Name] = Error
		AllService.Sched.UpdateJobStatus(jobId, Error)
		str := fmt.Sprintf("failed to exec %s backup's shell file for %s in job %s, error info-%s", backupTypeStr, host.Name, jobId, err.Error())
		s.mongoStore.JobResult.SaveJobResult(host.Name, jobId, str, startTime, GetTime(), "", Error, backupType)
		log.Printf(str)
		return
	}

	time.Sleep(2 * time.Second)
	//check result log file to determine whether backup succeed or fail
	s.doCheckBackupResult(ssh, jobId, host, startTime, backupType)
}

func (s ManualBackupService) FullBackup(jobId, serverName string) {
	if ok := bson.IsObjectIdHex(jobId); !ok {
		s.mongoStore.JobResult.SaveJobResult(serverName, jobId, fmt.Sprintf("wrong job id style,id:%s", jobId), GetTime(), GetTime(), "", Error, FULL)
		AllService.Sched.UpdateJobStatus(jobId, Error)
		return
	}
	host, err := s.mongoStore.Hosts.FindHostByName(serverName)
	if err != nil {
		s.mongoStore.JobResult.SaveJobResult(serverName, jobId, fmt.Sprintf("can not find host : %s", serverName), GetTime(), GetTime(), "", Error, FULL)
		AllService.Sched.UpdateJobStatus(jobId, Error)
		return
	}

	go s.doBackup(jobId, &host, FULL)
}

func (s ManualBackupService) IncBackup(jobId, serverName string) {
	if ok := bson.IsObjectIdHex(jobId); !ok {
		s.mongoStore.JobResult.SaveJobResult(serverName, jobId, fmt.Sprintf("wrong job id style,id:%s", jobId), GetTime(), GetTime(), "", Error, INCR)
		AllService.Sched.UpdateJobStatus(jobId, Error)
		return
	}
	host, err := s.mongoStore.Hosts.FindHostByName(serverName)
	if err != nil {
		s.mongoStore.JobResult.SaveJobResult(serverName, jobId, fmt.Sprintf("can not find host : %s", serverName), GetTime(), GetTime(), "", Error, INCR)
		AllService.Sched.UpdateJobStatus(jobId, Error)
		return
	}

	go s.doBackup(jobId, &host, INCR)
}

func (s ManualBackupService) IsInBackup(serverName string) bool {
	if status, ok := ServerStatus[serverName]; ok {
		if status == Running {
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}

func (s ManualBackupService) doCompress(jobId string, host *models.Host) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("compress find panic", err)
		}
	}()

	starttime := GetTime()
	AllService.Sched.UpdateJobStatus(jobId, Running)
	s.mongoStore.JobResult.SaveJobResult(host.Name, jobId, "", starttime, starttime, "", Running, COMPRESS)

	cmd := exec.Command("/bin/sh", "-c", "ssh-add")
	cmd.Output()
	ssh := makeSshClient(host)
	err := s.doScpShellFile(jobId, host, COMPRESS, starttime, ssh)
	if err != nil {
		AllService.Sched.UpdateJobStatus(jobId, Error)
		str := fmt.Sprintf("failed to scp file to %s in compress files job %s, error info-%s", host.Name, jobId, err.Error())
		s.mongoStore.JobResult.SaveJobResult(host.Name, jobId, str, starttime, GetTime(), "", Error, COMPRESS)
		log.Printf(str)
		return
	}

	cmdPara := fmt.Sprintf(CompressFile, host.BackupPath+"/"+utils.Cfg.Shellname, host.DBUser,
		host.DBPassword, host.DBPort, host.BackupPath, host.DBHost, host.DBSocket, host.DBMyCnf)
	_, err = ssh.Run(cmdPara)
	if err != nil {
		str := fmt.Sprintf("failed to compress backup file for %s in job %s, error info-%s", host.Name, jobId, err.Error())
		log.Printf(str)
		s.mongoStore.JobResult.SaveJobResult(host.Name, jobId, str, starttime, GetTime(), "", Error, COMPRESS)
		AllService.Sched.UpdateJobStatus(jobId, Error)
		return
	}

	//modify backup file's name in mongodb
	outputError, err := ssh.Run(fmt.Sprintf(GetCompressResultError, host.BackupPath))
	if err != nil {
		str := fmt.Sprintf("failed to get compress error log for %s in job %s, error info-%s", host.Name, jobId, err.Error())
		log.Printf(str)
		s.mongoStore.JobResult.SaveJobResult(host.Name, jobId, str, starttime, GetTime(), "", Error, COMPRESS)
		AllService.Sched.UpdateJobStatus(jobId, Error)
		return
	}
	if outputError != "" {
		str := fmt.Sprintf("failed to compress backup file for %s in job %s, error info-%s", host.Name, jobId, outputError)
		log.Printf(str)
		s.mongoStore.JobResult.SaveJobResult(host.Name, jobId, str, starttime, GetTime(), "", Error, COMPRESS)
		AllService.Sched.UpdateJobStatus(jobId, Error)
		return
	}

	output, err := ssh.Run(fmt.Sprintf(GetCompressResult, host.BackupPath))
	if err != nil {
		str := fmt.Sprintf("failed to get compress log for %s in job %s, error info-%s", host.Name, jobId, err.Error())
		log.Printf(str)
		s.mongoStore.JobResult.SaveJobResult(host.Name, jobId, str, starttime, GetTime(), "", Error, COMPRESS)
		AllService.Sched.UpdateJobStatus(jobId, Error)
		return
	}
	if output != "" {
		files := strings.Split(output, "\n")
		for _, file := range files {
			newpath := file + ".tar.gz"
			store.Store.BackupFileResult.UpdateBackupFilePath(host.Name, file, newpath)
		}
	}
	AllService.Sched.UpdateJobStatus(jobId, Idle)
	log.Printf("compress succeful for %s in job %s", host.Name, jobId)
	s.mongoStore.JobResult.SaveJobResult(host.Name, jobId, "", starttime, GetTime(), "", Finish, COMPRESS)
}

func (s ManualBackupService) Compress(jobId, serverName string) {
	if ok := bson.IsObjectIdHex(jobId); !ok {
		s.mongoStore.JobResult.SaveJobResult(serverName, jobId, fmt.Sprintf("wrong job id style,id:%s", jobId), GetTime(), GetTime(), "", Error, COMPRESS)
		AllService.Sched.UpdateJobStatus(jobId, Error)
		return
	}
	host, err := s.mongoStore.Hosts.FindHostByName(serverName)
	if err != nil {
		s.mongoStore.JobResult.SaveJobResult(serverName, jobId, fmt.Sprintf("can not find host : %s", serverName), GetTime(), GetTime(), "", Error, COMPRESS)
		AllService.Sched.UpdateJobStatus(jobId, Error)
		return
	}

	go s.doCompress(jobId, &host)
}

func (s ManualBackupService) doCheckAllRunningJobResult(jobresult *models.JobResult, host *models.Host) {
	ssh := makeSshClient(host)
	cmd := exec.Command("/bin/sh", "-c", "ssh-add")
	cmd.Output()

	s.doCheckBackupResult(ssh, jobresult.JobId.Hex(), host, jobresult.StartTime, jobresult.Type)
}

func (s ManualBackupService) CheckAllRunningJobResult() {
	jobresults, err := s.mongoStore.JobResult.FindAllBackupRunningResult()
	if err != nil {
		log.Printf("Failed to check all running job's result")
		return
	}
	for _, jobresult := range jobresults {
		host, err := s.mongoStore.Hosts.FindHostByName(jobresult.ServerName)
		if err != nil {
			log.Printf("Failed to find host:%s", jobresult.ServerName)
			return
		}
		go s.doCheckAllRunningJobResult(&jobresult, &host)
	}
}

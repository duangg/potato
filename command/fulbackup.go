package command
import (
	"fmt"
	"time"
	"log"
	"os/exec"
	"errors"
	"strconv"
	"strings"

	"github.com/hypersleep/easyssh"

	"github.com/SpruceX/potato/models"
	"github.com/SpruceX/potato/utils"
	"github.com/SpruceX/potato/store"
)

type BackupCmd struct {
	Target *easyssh.MakeConfig
	Host *models.Host
	BackType int
	BackupTypeStr string
	JobID string
	StartTime time.Time
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
	XtrabackupNotFind = "sudo cat %s/log | grep \"innobackupex: not found\" "
)

func (b *BackupCmd) UploadScript() error {
	log.Printf("upload script to host:%s in %s backup job:%s\n", b.Host.Name, b.BackupTypeStr, b.JobID)
	_, errTest := b.Target.Run("ls")
	if errTest != nil {
		return errors.New(fmt.Sprintf("failed to scp file to %s in %s backup job %s, error info-%s",
			b.BackupTypeStr, b.Host.Name, b.JobID, errTest.Error()))
	}

	cmdPara := fmt.Sprintf("%s@%s:%s/", b.Host.UserName, b.Host.IP, b.Host.BackupPath)
	cmdParaPort := fmt.Sprintf("-P%s", b.Host.SshPort)
	cmd := exec.Command("scp", cmdParaPort, "./script/"+utils.Cfg.Shellname, cmdPara)
	_, err := cmd.Output()
	if err != nil {
		return errors.New(fmt.Sprintf("failed to scp file to %s in %s backup job %s, error info-%s",
			b.BackupTypeStr, b.Host.Name, b.JobID, err.Error()))
	}
	return nil
}

func (b *BackupCmd) CleanUp() error {
	log.Printf("clean up in host:%s in %s backup job:%s\n", b.Host.Name, b.BackupTypeStr, b.JobID)
	_, err := b.Target.Run(fmt.Sprintf(ClearBackupLog, b.Host.BackupPath))
	if err != nil {
		fmt.Printf("CleanUp error-%s\n ", err.Error())
		return errors.New(fmt.Sprintf("failed to clear %s backup's log for %s in job %s, error info-%s",
			b.BackupTypeStr, b.Host.Name, b.JobID, err.Error()))
	}
	return nil
}

func (b *BackupCmd) Execute() error {
	log.Printf("execute script in host:%s in %s backup job:%s\n", b.Host.Name, b.BackupTypeStr, b.JobID)
	shellStr := ""
	if b.BackType == models.JobTypeFullBackup {
		shellStr = ShellFullBackupContent
	} else {
		shellStr = ShellIncBackupContent
	}

	serverPara := fmt.Sprintf("%s@%s", b.Host.UserName, b.Host.IP)
	cmdPara := fmt.Sprintf(shellStr, b.Host.BackupPath+"/"+utils.Cfg.Shellname, b.Host.DBUser,
		b.Host.DBPassword, b.Host.DBPort, b.Host.BackupPath, b.Host.DBHost, b.Host.DBSocket, b.Host.DBMyCnf)
	cmdPort := fmt.Sprintf("%s", b.Host.SshPort)
	cmd := exec.Command("sh", "./script/" + ShellExecBackupShell, cmdPort, serverPara, cmdPara)
	_, err := cmd.Output()
	if err != nil {
		return errors.New(fmt.Sprintf("failed to exec %s backup's shell file for %s in job %s, error info-%s",
			b.BackupTypeStr, b.Host.Name, b.JobID, err.Error()))
	}
	return nil
}

func (b *BackupCmd) updateBackupLog() {
	output, err := b.Target.Run(fmt.Sprintf(GetBackupLog, b.Host.BackupPath))
	if err != nil {
	} else {
		if store.Store == nil {
			fmt.Println("store.Store is null")
		}
		store.Store.JobResult.SaveJobResult(b.Host.Name, b.JobID, "", b.StartTime, time.Now(), output, models.JobInProgress, b.BackType)
	}
}

func (b *BackupCmd) getBackupErrInfo( ) error {
	getErrInfo, errGetErrInfo := b.Target.Run(fmt.Sprintf(GetResultErrInfo, b.Host.BackupPath))
	fmt.Printf("getErrInfo-%s,errGetErrInfo-%s \n", getErrInfo, errGetErrInfo.Error())
	if errGetErrInfo == nil {
		return errors.New(fmt.Sprintf("failed to %s backup for %s in job %s, error info-%s",
			b.BackupTypeStr, b.Host.Name, b.JobID, getErrInfo))
	} else {

	}
	return nil
}

func (b *BackupCmd) CheckResult() error {
	log.Printf("check backup result in host:%s in %s backup job:%s\n", b.Host.Name, b.BackupTypeStr, b.JobID)
	ret := models.BackupComplete
	result := ""
	for {
		b.updateBackupLog()
		resultGetResultOK, errGetResultOK := b.Target.Run(fmt.Sprintf(GetResultOK, b.Host.BackupPath))
		if errGetResultOK != nil {
			result = fmt.Sprintf("failed to get %s backup complete flag for %s in job %s, error info-%s",
				b.BackupTypeStr, b.Host.Name, b.JobID, errGetResultOK.Error())
			ret = models.BackupErrorSSH
		} else {
			res, _ := strconv.Atoi(strings.Trim(resultGetResultOK, "\n"))
			if res >= 2 {
				ret = models.BackupComplete
				break
			}
		}

		_, errNotFindError := b.Target.Run(fmt.Sprintf(XtrabackupNotFind, b.Host.BackupPath))
		if errNotFindError != nil {
			result = fmt.Sprintf("failed to check xtraback exist in %s backup for %s in job %s, error info-%s",
				b.BackupTypeStr, b.Host.Name, b.JobID, errNotFindError.Error())
			ret = models.BackupErrorSSH
		} else {
			ret = models.XtrabackupNotFind
			break
		}
		resultGetResultError, errGetResultError := b.Target.Run(fmt.Sprintf(GetResultError, b.Host.BackupPath))
		if errGetResultError != nil {
			result = fmt.Sprintf("failed to get error info in %s backup for %s in job %s, error info-%s",
				b.BackupTypeStr, b.Host.Name, b.JobID, errGetResultError.Error())
			ret = models.BackupErrorSSH
		} else {
			resError, _ := strconv.Atoi(strings.Trim(resultGetResultError, "\n"))
			if resError >= 1 {
				ret = models.BackupError
				break
			}
		}
		time.Sleep(time.Second)
	}
	b.updateBackupLog()

	switch ret {
	case models.BackupComplete:
		return b.getBackupSuccInfo()
	case models.BackupError:
		return b.getBackupErrInfo()
	case models.BackupErrorSSH:
		return errors.New(result)
	case models.XtrabackupNotFind:
		strErr := "innobackupex is not found"
		return errors.New(fmt.Sprintf("failed to %s backup for %s in job %s, error info-%s", b.BackupTypeStr, b.Host.Name, b.JobID, strErr))
	}
	return nil
}

func (b *BackupCmd) getBackupSuccInfo() error {
	//file full path  /dbdata/20151225/ful/2015-12-25_16-56-32
	getBackupFile, errGetBackupFile := b.Target.Run(fmt.Sprintf(GetBackupFile, b.Host.BackupPath))
	if errGetBackupFile == nil {
		getBackupFile = strings.Trim(getBackupFile, "\n")
		if getBackupFile == "" {
			return nil
		}
		pos := strings.LastIndex(getBackupFile, "/")
		filename := getBackupFile[pos+1:]
		filedate := strings.Split(filename, "_")

		//get backup file size
		output, err := b.Target.Run(fmt.Sprintf(GetBackupFileSize, getBackupFile))
		if err != nil {
			return errors.New(fmt.Sprintf("failed to get %s backup file size for %s in job %s, error info-%s",
				b.BackupTypeStr, b.Host.Name, b.JobID, err.Error()))
		}
		//write backup file db and job result db
		filesizeinfo := strings.Split(output, "\t")
		store.Store.BackupFileResult.SaveBackupFileResult(b.Host.Name, filedate[0], b.BackType, b.JobID, getBackupFile, filesizeinfo[0], time.Now())
		log.Printf("%s backup succeful for %s in job %s, backup file:%s", b.BackupTypeStr, b.Host.Name, b.JobID, getBackupFile)
	} else {

	}
	return nil
}

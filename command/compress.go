package command
import (
	"time"
	"log"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/hypersleep/easyssh"

	"github.com/SpruceX/potato/models"
	"github.com/SpruceX/potato/utils"
	"github.com/SpruceX/potato/store"
)

const (
	//sh /BackupPath/xx.sh full_backup DBUser DBPassword DBPort BackupPath DBHost DBSocket DBMyCnf
	CompressFile           = "sudo sh %s scp_lftp %s \"%s\" %s %s %s %s %s"
	GetCompressResult      = "sudo cat %s/compresslog"
	GetCompressResultError = "sudo cat %s/compresslog | grep \"ERROR\" | head -n 1 | awk -F\":\" '{print $2}'"
)

type CompressCmd struct {
	Target *easyssh.MakeConfig
	Host *models.Host
	BackType int
	BackupTypeStr string
	JobID string
	StartTime time.Time
}

func (b *CompressCmd) UploadScript() error {
	log.Printf("upload script to host:%s in %s job:%s\n", b.Host.Name, b.BackupTypeStr, b.JobID)
	_, errTest := b.Target.Run("ls")
	if errTest != nil {
		return errors.New(fmt.Sprintf("failed to scp file to %s in %s job %s, error info-%s",
			b.BackupTypeStr, b.Host.Name, b.JobID, errTest.Error()))
	}

	cmdPara := fmt.Sprintf("%s@%s:%s/", b.Host.UserName, b.Host.IP, b.Host.BackupPath)
	cmdParaPort := fmt.Sprintf("-P%s", b.Host.SshPort)
	cmd := exec.Command("scp", cmdParaPort, "./script/"+utils.Cfg.Shellname, cmdPara)
	_, err := cmd.Output()
	if err != nil {
		return errors.New(fmt.Sprintf("failed to scp file to %s in %s job %s, error info-%s",
			b.BackupTypeStr, b.Host.Name, b.JobID, err.Error()))
	}
	return nil
}

func (b *CompressCmd) CleanUp() error {
	return nil
}

func (b *CompressCmd) Execute() error {
	log.Printf("execute script in host:%s in %s job:%s\n", b.Host.Name, b.BackupTypeStr, b.JobID)
	cmdPara := fmt.Sprintf(CompressFile, b.Host.BackupPath+"/"+utils.Cfg.Shellname, b.Host.DBUser,
		b.Host.DBPassword, b.Host.DBPort, b.Host.BackupPath, b.Host.DBHost, b.Host.DBSocket, b.Host.DBMyCnf)
	_, err := b.Target.Run(cmdPara)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to compress file for %s in job %s, error info-%s", b.Host.Name, b.JobID, err.Error()))
	}
	return nil
}

func (b *CompressCmd) getCompressErrInfo( ) error {
	outputError, err := b.Target.Run(fmt.Sprintf(GetCompressResultError, b.Host.BackupPath))
	if err != nil {
		str := fmt.Sprintf("failed to get compress error log for %s in job %s, error info-%s", b.Host.Name, b.JobID, err.Error())
		return errors.New(str)
	}
	if outputError != "" {
		str := fmt.Sprintf("failed to compress backup file for %s in job %s, error info-%s", b.Host.Name, b.JobID, outputError)
		return errors.New(str)
	}
	return nil
}

func (b *CompressCmd) CheckResult() error {
	log.Printf("check result in host:%s in %s job:%s\n", b.Host.Name, b.BackupTypeStr, b.JobID)
	err := b.getCompressErrInfo()
	if err != nil {
		return err
	}
	return b.getCompressSuccInfo()
}

func (b *CompressCmd) getCompressSuccInfo() error {
	output, err := b.Target.Run(fmt.Sprintf(GetCompressResult, b.Host.BackupPath))
	if err != nil {
		str := fmt.Sprintf("failed to get compress log for %s in job %s, error info-%s", b.Host.Name, b.JobID, err.Error())
		return errors.New(str)
	}
	if output != "" {
		files := strings.Split(output, "\n")
		for _, file := range files {
			newpath := file + ".tar.gz"
			store.Store.BackupFileResult.UpdateBackupFilePath(b.Host.Name, file, newpath)
		}
	}
	log.Printf("compress succeful for %s in job %s", b.Host.Name, b.JobID)
	return nil
}


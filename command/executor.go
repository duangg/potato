package command

var SSHExecutor = sshExecutor{}

type sshExecutor struct {}

func (s sshExecutor) Execute(j AsyncSSHJob) (err error){
	if err = j.UploadScript(); err!=nil{
		return err
	}
	if err = j.CleanUp(); err!=nil{
		return err
	}
	if err = j.Execute(); err!=nil{
		return err
	}
	if err = j.CheckResult(); err!=nil{
		return err
	}

	return nil
}
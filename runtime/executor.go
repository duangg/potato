package runtime

var SSHExecutor = sshExecutor{}

type sshExecutor struct {}

func (s sshExecutor) Execute(j AsyncSSHJob) (err error){
	if err = j.UploadScript(); err!=nil{
		return
	}

	if err = j.Execute(); err!=nil{
		return
	}

	if err = j.CheckResult(); err!=nil{
		return
	}

	return
}
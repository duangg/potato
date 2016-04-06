package runtime

//TODO(runyang) abstract the ssh command interface
type AsyncSSHJob interface {
	UploadScript() error
	Execute() error
	CheckResult() error
	CleanUp() error
}

package runtime
import "github.com/hypersleep/easyssh"

type FullBackupCmd struct {
	target *easyssh.MakeConfig
	script string
	params []string
}

func (b *FullBackupCmd) UploadScript(file string) error {return nil}


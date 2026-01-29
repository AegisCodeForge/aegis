//go:build ignore
package templates

import "fmt"
import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"


func(cfg gitus.GitusConfig, repo model.Repository) string {
	gitSshHostName := cfg.GitSSHHostName()
	sshfn := fmt.Sprintf("%s/%s", repo.Namespace, repo.Name)
	return fmt.Sprintf("%s%s", gitSshHostName, sshfn)
}


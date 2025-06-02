//go:build ignore
package templates

import "fmt"
import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"


func(cfg aegis.AegisConfig, repo model.Repository) string {
	gitSshHostName := cfg.GitSSHHostName()
	sshfn := fmt.Sprintf("%s/%s", repo.Namespace, repo.Name)
	return fmt.Sprintf("%s%s", gitSshHostName, sshfn)
}


//go:build ignore
package templates

import "fmt"
import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"


func(cfg gitus.GitusConfig, repo model.Repository) string {
	httpHostName := cfg.ProperHTTPHostName()
	rfn := repo.FullName()
	return fmt.Sprintf("%s/repo/%s", httpHostName, rfn)
}


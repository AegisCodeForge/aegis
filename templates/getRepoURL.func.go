//go:build ignore
package templates

import "fmt"
import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"


func(cfg aegis.AegisConfig, repo model.Repository) string {
	httpHostName := cfg.ProperHTTPHostName()
	rfn := repo.FullName()
	return fmt.Sprintf("%s/repo/%s", httpHostName, rfn)
}


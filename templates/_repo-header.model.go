//go:build ignore
package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"

type RepoHeaderTemplateModel struct {
	RepoURL string
	RepoSSH string
	TypeStr string
	NodeName string
	Repository *model.Repository
	RepoLabelList []string
	Config *aegis.AegisConfig
}


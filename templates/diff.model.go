//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/gitlib"

type DiffTemplateModel struct {
	Config *aegis.AegisConfig
	RepoHeaderInfo RepoHeaderTemplateModel
	CommitInfo CommitInfoTemplateModel
	Diff string
	LoginInfo *LoginInfoModel
}

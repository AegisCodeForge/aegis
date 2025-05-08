//go:build ignore

package templates

import "github.com/bctnry/gitus/pkg/gitus"
import "github.com/bctnry/gitus/pkg/gitlib"

type DiffTemplateModel struct {
	Config *gitus.GitusConfig
	RepoHeaderInfo RepoHeaderTemplateModel
	CommitInfo CommitInfoTemplateModel
	Diff string
	LoginInfo *LoginInfoModel
}

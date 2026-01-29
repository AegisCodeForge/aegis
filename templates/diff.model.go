//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitlib"

type DiffTemplateModel struct {
	Config *gitus.GitusConfig
	Repository *model.Repository
	RepoHeaderInfo RepoHeaderTemplateModel
	CommitInfo CommitInfoTemplateModel
	Diff *gitlib.Diff
	LoginInfo *LoginInfoModel
}

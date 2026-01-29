//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitlib"

type CommitHistoryModel struct {
	Config *gitus.GitusConfig
	Repository *model.Repository
	RepoHeaderInfo RepoHeaderTemplateModel
	Commit gitlib.CommitObject
	CommitHistory []gitlib.CommitObject
	LoginInfo *LoginInfoModel
	NextPageCommitId string
	EmailUserMapping map[string]string
}


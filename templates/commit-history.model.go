//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/gitlib"

type CommitHistoryModel struct {
	Config *aegis.AegisConfig
	Repository *model.Repository
	RepoHeaderInfo RepoHeaderTemplateModel
	Commit gitlib.CommitObject
	CommitHistory []gitlib.CommitObject
	LoginInfo *LoginInfoModel
	NextPageCommitId string
}


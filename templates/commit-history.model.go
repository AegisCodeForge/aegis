//go:build ignore

package templates

import "github.com/bctnry/gitus/pkg/gitus"
import "github.com/bctnry/gitus/pkg/gitlib"

type CommitHistoryModel struct {
	Config *gitus.GitusConfig
	RepoHeaderInfo RepoHeaderTemplateModel
	Commit gitlib.CommitObject
	CommitHistory []gitlib.CommitObject
	LoginInfo *LoginInfoModel
}


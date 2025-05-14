//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/gitlib"

type CommitHistoryModel struct {
	Config *aegis.AegisConfig
	RepoHeaderInfo RepoHeaderTemplateModel
	Commit gitlib.CommitObject
	CommitHistory []gitlib.CommitObject
	LoginInfo *LoginInfoModel
}


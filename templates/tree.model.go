//go:build ignore
package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"

type TreeTemplateModel struct {
	Config *aegis.AegisConfig
	Repository *model.Repository
	RepoHeaderInfo RepoHeaderTemplateModel
	TreeFileList *TreeFileListTemplateModel
	ParentTreeFileList *TreeFileListTemplateModel
	PermaLink string
	
	ComparisonInfo *gitlib.BranchComparisonInfo
	TreePath *TreePathTemplateModel
	CommitInfo *CommitInfoTemplateModel
	TagInfo *TagInfoTemplateModel

	LoginInfo *LoginInfoModel
}


//go:build ignore
package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/gitlib"

type RepositoryModel struct{
	Config *aegis.AegisConfig
	Repository *model.Repository
	RepoHeaderInfo RepoHeaderTemplateModel
	BranchList map[string]*gitlib.Branch
	TagList map[string]*gitlib.Tag
	LoginInfo *LoginInfoModel

	MajorBranchPermaLink string
	ReadmeString string
	TreeFileList *TreeFileListTemplateModel
	CommitInfo *CommitInfoTemplateModel
	ComparisonInfo *gitlib.BranchComparisonInfo
}

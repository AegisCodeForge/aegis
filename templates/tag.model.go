//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"
import "github.com/bctnry/aegis/pkg/gitlib"

type TagTemplateModel struct {
	Config *aegis.AegisConfig
	Repository *model.Repository
	RepoHeaderInfo RepoHeaderTemplateModel
	Tag *gitlib.TagObject
	TagInfo *TagInfoTemplateModel
	LoginInfo *LoginInfoModel
}


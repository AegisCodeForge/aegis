//go:build ignore

package templates

import "github.com/bctnry/gitus/pkg/gitus"
import "github.com/bctnry/gitus/pkg/gitlib"

type TagTemplateModel struct {
	Config *gitus.GitusConfig
	RepoHeaderInfo RepoHeaderTemplateModel
	Tag *gitlib.TagObject
	TagInfo *TagInfoTemplateModel
	LoginInfo *LoginInfoModel
}


//go:build ignore

package templates

import "github.com/bctnry/gitus/pkg/gitlib"

type TagTemplateModel struct {
	RepoHeaderInfo RepoHeaderTemplateModel
	Tag *gitlib.TagObject
	TagInfo *TagInfoTemplateModel
}


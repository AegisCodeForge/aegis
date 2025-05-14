//go:build ignore

package templates

//go:build ignore

import "github.com/bctnry/aegis/pkg/gitlib"

type TagInfoTemplateModel struct {
	// it should be made sure that when Annotated is true, Tag is nil,
	// and vice versa.
	Annotated bool
	RepoName string
	Tag *gitlib.TagObject
}


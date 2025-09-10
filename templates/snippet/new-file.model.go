//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"
import "github.com/bctnry/aegis/pkg/gitlib"

type SnippetNewFileTemplateModel struct {
	Config *aegis.AegisConfig
	LoginInfo *LoginInfoModel
	BelongingUser string
	Name string
}


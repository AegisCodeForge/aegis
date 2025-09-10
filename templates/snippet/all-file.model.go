//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"
import "github.com/bctnry/aegis/pkg/gitlib"

type SnippetAllFileTemplateModel struct {
	Config *aegis.AegisConfig
	LoginInfo *LoginInfoModel
	Snippet *model.Snippet
	DisplayingFileList map[string]string
}


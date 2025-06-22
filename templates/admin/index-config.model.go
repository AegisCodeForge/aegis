//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"

type AdminIndexConfigTemplateModel struct {
	Config *aegis.AegisConfig
	IndexType string
	IndexNamespace string
	IndexRepository string
	IndexFileContent string
	ErrorMsg string
	LoginInfo *LoginInfoModel
}


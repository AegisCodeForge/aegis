//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"

type NamespaceSettingEditMemberTemplateModel struct {
	Config *aegis.AegisConfig
	Namespace *model.Namespace
	LoginInfo *LoginInfoModel
	ErrorMsg string
	Username string
	ACLTuple *model.ACLTuple
}


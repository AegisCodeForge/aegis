//go:build ignore

package templates

import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"

type NamespaceSettingMemberListTemplateModel struct {
	Config *aegis.AegisConfig
	Namespace *model.Namespace
	LoginInfo *LoginInfoModel
	ErrorMsg string
	ACL map[string]*model.ACLTuple
	PageInfo *PageInfoModel
}


//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type NamespaceSettingMemberListTemplateModel struct {
	Config *gitus.GitusConfig
	Namespace *model.Namespace
	LoginInfo *LoginInfoModel
	ErrorMsg string
	ACL map[string]*model.ACLTuple
	PageInfo *PageInfoModel
}


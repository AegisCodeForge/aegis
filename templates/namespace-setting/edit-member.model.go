//go:build ignore

package templates

import "github.com/GitusCodeForge/Gitus/pkg/gitus"
import "github.com/GitusCodeForge/Gitus/pkg/gitus/model"

type NamespaceSettingEditMemberTemplateModel struct {
	Config *gitus.GitusConfig
	Namespace *model.Namespace
	LoginInfo *LoginInfoModel
	ErrorMsg string
	Username string
	ACLTuple *model.ACLTuple
}


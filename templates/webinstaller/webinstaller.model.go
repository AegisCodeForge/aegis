//go:build ignore

package webinstaller

import "github.com/GitusCodeForge/Gitus/pkg/gitus"

type WebInstallerTemplateModel struct {
	Config *gitus.GitusConfig
	RootSSHKey string
	ConfirmStageReached bool
}


//go:build ignore

package webinstaller

import "github.com/bctnry/aegis/pkg/aegis"

type WebInstallerTemplateModel struct {
	Config *aegis.AegisConfig
	RootSSHKey string
	ConfirmStageReached bool
}


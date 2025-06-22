//go:build ignore

package webinstaller

import "github.com/bctnry/aegis/pkg/aegis"

type WebInstallerTemplateModel struct {
	Config *aegis.AegisConfig
	ConfirmStageReached bool
}


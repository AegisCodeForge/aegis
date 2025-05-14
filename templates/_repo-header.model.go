//go:build ignore
package templates

import "github.com/bctnry/aegis/pkg/aegis"

type RepoHeaderTemplateModel struct {
	NamespaceName string
	RepoName string
	RepoDescription string
	RepoURL string
	RepoSSH string
	TypeStr string
	NodeName string
	RepoLabelList []string
	Config *aegis.AegisConfig
}


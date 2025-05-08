//go:build ignore
package templates

import "github.com/bctnry/gitus/pkg/gitus"

type RepoHeaderTemplateModel struct {
	NamespaceName string
	RepoName string
	RepoDescription string
	RepoURL string
	TypeStr string
	NodeName string
	RepoLabelList []string
	Config *gitus.GitusConfig
}


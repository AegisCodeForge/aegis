//go:build ignore

package templates

import "github.com/bctnry/gitus/pkg/gitus/model"

type NamespaceTemplateModel struct {
	DepotName string
	Namespace *model.Namespace
}


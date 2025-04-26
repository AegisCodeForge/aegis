//go:build ignore

package templates

import "github.com/bctnry/gitus/pkg/gitus/model"

type DepotNamespaceModel struct {
	DepotName string
	NamespaceList map[string]*model.Namespace
}


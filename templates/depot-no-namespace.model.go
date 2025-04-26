//go:build ignore

package templates

type DepotNoNamespaceModel struct {
	RepositoryList []struct{RelPath string; Description string}
	DepotName string
}


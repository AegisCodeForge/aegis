//go:build ignore

package templates

type TreeFileListTemplateModel struct {
	ShouldHaveParentLink bool
	RootPath string
	TreePath string
	FileList []gitlib.TreeObjectItem
}


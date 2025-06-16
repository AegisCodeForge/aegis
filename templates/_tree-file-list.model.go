//go:build ignore

package templates

type TreeFileListTemplateModel struct {
	ShouldHaveParentLink bool
	// the "root path", i.e. the path to the particular branch/commit/tag.
	// no trailing slash.
	RootPath string
	// the relative path of the tree within the branch/commit/tag.
	// does not start with a slash, but should end with a trailing slash.
	TreePath string
	FileList []gitlib.TreeObjectItem
}


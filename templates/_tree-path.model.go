//go:build ignore

package templates

type TreePathTemplateModel struct {
	// the full name of the root. "the root" here represents the
	// branch/commit/tree the path is based upon. for example, the file
	// `src/main.go` in the branch `master` of the repository `myrepo.git`
	// would have a root full name of `myrepo.git@branch:master` and a
	// root path of `/repo/myrepo.git/branch/master`. (the full path to that
	// particular file would then be `/repo/myrepo.git/branch/master/src/main.go`.)
	// things are similar for cases like trees, tags and commits.
	RootFullName string
	// the full path to the root.
	RootPath string
	TreePath string
	TreePathSegmentList []struct{
		Name string
		RelPath string
	}
}


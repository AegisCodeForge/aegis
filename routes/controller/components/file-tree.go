package components

import (
	"fmt"
	"html"
	"slices"
	"strings"

	"github.com/bctnry/gitus/pkg/gitlib"
	"github.com/bctnry/gitus/routes"
)

func renderFileTreeFileItem(r *strings.Builder, ctx *routes.RouterContext, baseUrl string, ft *gitlib.TreeFileNode) error {
	r.WriteString(fmt.Sprintf("<div class=\"file-tree-item file-tree-item-file\"><a href=\"%s%s\">%s</a></div>", html.EscapeString(baseUrl), html.EscapeString(ft.GetPath()), html.EscapeString(ft.GetName())))
	return nil
}
func renderFileTreeDirItem(r *strings.Builder, ctx *routes.RouterContext, baseUrl string, ft *gitlib.TreeDirNode, noWrapper bool) error {
	// because the order of range over map is undetermined so we
	// must arrange them on our own...
	dirs := make([]*gitlib.TreeDirNode, 0)
	files := make([]*gitlib.TreeFileNode, 0)
	for _, item := range ft.Children {
		if item.GetType() == gitlib.FILE {
			files = append(files, item.(*gitlib.TreeFileNode))
		} else if item.GetType() == gitlib.DIRECTORY {
			dirs = append(dirs, item.(*gitlib.TreeDirNode))
		}
	}
	slices.SortFunc(dirs, func(a, b *gitlib.TreeDirNode) int {
		return strings.Compare(a.Name, b.Name)
	})
	slices.SortFunc(files, func(a, b *gitlib.TreeFileNode) int {
		return strings.Compare(a.Name, b.Name)
	})
	if !noWrapper {
		r.WriteString("<details class=\"file-tree-item file-tree-item-dir\">")
		r.WriteString(fmt.Sprintf("<summary>%s</summary>", html.EscapeString(ft.GetName())))
	}
	for _, item := range dirs {
		renderFileTreeDirItem(r, ctx, baseUrl, item, false)
	}
	for _, item := range files {
		renderFileTreeFileItem(r, ctx, baseUrl, item)
	}
	if !noWrapper {
		r.WriteString("</details>")
	}
	return nil
}


func RenderFileTree(ctx *routes.RouterContext, baseUrl string, ft gitlib.TreeNode) (string, error) {
	r := new(strings.Builder)
	switch ft.GetType() {
	case gitlib.FILE:
		renderFileTreeFileItem(r, ctx, baseUrl, ft.(*gitlib.TreeFileNode))
	case gitlib.DIRECTORY:
		// we don't wrap the result in <div> because we expect the wrapper
		// to be in the template.
		renderFileTreeDirItem(r, ctx, baseUrl, ft.(*gitlib.TreeDirNode), true)
	} 
	return r.String(), nil
}


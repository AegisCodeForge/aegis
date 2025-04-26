package controller

import (
	"archive/zip"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/bctnry/gitus/pkg/gitlib"
)

func basicStringEscape(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\\", "\\\\"), "\"", "\\\"")
}

func writeTree(repo *gitlib.LocalGitRepository, w *zip.Writer, pathPrefix string, tree *gitlib.TreeObject) error {
	// `pathPrefix` should be empty or end with slash.
	for _, item := range tree.ObjectList {
		pathname := fmt.Sprintf("%s%s", pathPrefix, item.Name)
		obj, err := repo.ReadObject(item.Hash)
		if err != nil { return err }
		switch item.Mode {
		case gitlib.TREE_NORMAL_FILE: fallthrough
		case gitlib.TREE_EXECUTABLE_FILE: fallthrough
		case gitlib.TREE_SYMBOLIC_LINK:
			// go's zip library in stdlib seems to not have anything
			// that supports ymbolic links. we might get away with not
			// supporting it...
			wr, err := w.Create(pathname)
			if err != nil { return err }
			if obj.Type() != gitlib.BLOB {
				return errors.New(fmt.Sprintf("%s is not a blob object", obj.ObjectId()))
			}
			wr.Write(obj.RawData())
		case gitlib.TREE_TREE_OBJECT:
			tobj, ok := obj.(*gitlib.TreeObject)
			if !ok {
				return errors.New(fmt.Sprintf("%s is not a blob object", obj.ObjectId()))
			}
			writeTree(repo, w, pathname+"/", tobj)
		case gitlib.TREE_SUBMODULE:
			// we don't support submodule at the moment...
			break
		}
	}
	return nil
}

func responseWithTreeZip(repo *gitlib.LocalGitRepository, obj gitlib.GitObject, name string, w http.ResponseWriter, r *http.Request) error {
	// requires:
	// + `name` to be descriptive and without the `.zip` extension name.
	// + `obj` to be a tree object.
	tobj, ok := obj.(*gitlib.TreeObject)
	if !ok {
		return errors.New(fmt.Sprintf(
			"%s is not a tree object",
			obj.ObjectId(),
		))
	}
	filenameStar := url.QueryEscape(fmt.Sprintf("%s.zip", name))
	// it was said that "browsers handle escape sequences
	// differently", but i would assume that most of them would at
	// least handle \" and \\...
	filename := fmt.Sprintf("\"%s.zip\"", basicStringEscape(name))
	w.Header().Add(
		"Content-Disposition",
		fmt.Sprintf("attachment; filename=%s; filename*=UTF-8''%s", filename, filenameStar),
	)
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()
	err := writeTree(repo, zipWriter, "", tobj)
	if err != nil { return err }
	zipWriter.Flush()
	return nil
}


package controller

import (
	"fmt"
	"net/http"
	"os"
	"path"

	. "github.com/bctnry/gitus/routes"
)

// info/refs
// HEAD
// objects/

func bindHttpCloneController(ctx RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/info/{p...}", WithLog(func(w http.ResponseWriter, r *http.Request) {
		repoName := r.PathValue("repoName")
		repo, ok := ctx.GitRepositoryList[repoName]
		if !ok {
			ctx.ReportNotFound(repoName, "Repository", ctx.Config.DepotName, w, r)
			return
		}
		fmt.Println(r.URL.Query())
		p := path.Join(repo.GitDirectoryPath, "info", r.PathValue("p"))
		s, err := os.ReadFile(p)
		if err != nil {
			ctx.ReportInternalError("Fail to read info/refs", w, r)
			return
		}
		w.Write(s)
	}))
	http.HandleFunc("GET /repo/{repoName}/HEAD", WithLog(func(w http.ResponseWriter, r *http.Request) {
		repoName := r.PathValue("repoName")
		repo, ok := ctx.GitRepositoryList[repoName]
		if !ok {
			ctx.ReportNotFound(repoName, "Repository", ctx.Config.DepotName, w, r)
			return
		}
		fmt.Println(r.URL.Query())
		p := path.Join(repo.GitDirectoryPath, "HEAD")
		s, err := os.ReadFile(p)
		if err != nil {
			ctx.ReportInternalError("Fail to read info/refs", w, r)
			return
		}
		w.Write(s)
	}))
	http.HandleFunc("GET /repo/{repoName}/objects/{obj...}", WithLog(func(w http.ResponseWriter, r *http.Request) {
		repoName := r.PathValue("repoName")
		repo, ok := ctx.GitRepositoryList[repoName]
		if !ok {
			ctx.ReportNotFound(repoName, "Repository", ctx.Config.DepotName, w, r)
			return
		}
		obj := r.PathValue("obj")
		p := path.Join(repo.GitDirectoryPath, "objects", obj)
		s, err := os.ReadFile(p)
		if os.IsNotExist(err) {
			ctx.ReportNotFound(repoName, "object", ctx.Config.DepotName, w, r)
			return
		}
		if err != nil {
			ctx.ReportInternalError("Fail to read info/refs", w, r)
			return
		}
		w.Write(s)
	}))
}

package controller

import (
	"net/http"
	"os"
	"path"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/routes"
	. "github.com/bctnry/aegis/routes"
)

// info/refs
// HEAD
// objects/

// NOTE THAT this route handles public http read-only clone, thus it
// will report 404 for all private repositories.

func bindHttpCloneController(ctx *RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/info/{p...}", WithLog(func(w http.ResponseWriter, r *http.Request) {
		if ctx.Config.GlobalVisibility != aegis.GLOBAL_VISIBILITY_PUBLIC {
			ctx.ReportForbidden("", w, r)
			return
		}
		rfn := r.PathValue("repoName")
		_, _, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		if err == routes.ErrNotFound {
			ctx.ReportNotFound(rfn, "Repository", "Depot", w, r)
			return
		}
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		isNamespacePublic := ns.Status == model.NAMESPACE_NORMAL_PUBLIC
		isRepoPublic := repo.Status == model.REPO_NORMAL_PUBLIC
		isRepoArchived := repo.Status == model.REPO_ARCHIVED
		if !isNamespacePublic || !(isRepoPublic || isRepoArchived) {
			w.WriteHeader(404)
			w.Write([]byte("404 Not Found"))
			return
		}
		p := path.Join(repo.Repository.GitDirectoryPath, "info", r.PathValue("p"))
		s, err := os.ReadFile(p)
		if err != nil {
			ctx.ReportInternalError("Fail to read info/refs", w, r)
			return
		}
		w.Write(s)
	}))
	http.HandleFunc("GET /repo/{repoName}/HEAD", WithLog(func(w http.ResponseWriter, r *http.Request) {
		if ctx.Config.GlobalVisibility != aegis.GLOBAL_VISIBILITY_PUBLIC {
			ctx.ReportForbidden("", w, r)
			return
		}
		rfn := r.PathValue("repoName")
		_, _, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		if err == routes.ErrNotFound {
			ctx.ReportNotFound(rfn, "Repository", "Depot", w, r)
			return
		}
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		isNamespacePublic := ns.Status == model.NAMESPACE_NORMAL_PUBLIC
		isRepoPublic := repo.Status == model.REPO_NORMAL_PUBLIC
		isRepoArchived := repo.Status == model.REPO_ARCHIVED
		if !isNamespacePublic || !(isRepoPublic || isRepoArchived) {
			w.WriteHeader(404)
			w.Write([]byte("404 Not Found"))
			return
		}
		p := path.Join(repo.Repository.GitDirectoryPath, "HEAD")
		s, err := os.ReadFile(p)
		if err != nil {
			ctx.ReportInternalError("Fail to read info/refs", w, r)
			return
		}
		w.Write(s)
	}))
	http.HandleFunc("GET /repo/{repoName}/objects/{obj...}", WithLog(func(w http.ResponseWriter, r *http.Request) {
		if ctx.Config.GlobalVisibility != aegis.GLOBAL_VISIBILITY_PUBLIC {
			ctx.ReportForbidden("", w, r)
			return
		}
		rfn := r.PathValue("repoName")
		_, _, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		if err == routes.ErrNotFound {
			ctx.ReportNotFound(rfn, "Repository", "Depot", w, r)
			return
		}
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		isNamespacePublic := ns.Status == model.NAMESPACE_NORMAL_PUBLIC
		isRepoPublic := repo.Status == model.REPO_NORMAL_PUBLIC
		isRepoArchived := repo.Status == model.REPO_ARCHIVED
		if !isNamespacePublic || !(isRepoPublic || isRepoArchived) {
			w.WriteHeader(404)
			w.Write([]byte("404 Not Found"))
			return
		}
		obj := r.PathValue("obj")
		p := path.Join(repo.Repository.GitDirectoryPath, "objects", obj)
		s, err := os.ReadFile(p)
		if os.IsNotExist(err) {
			ctx.ReportNotFound(rfn, "object", ctx.Config.DepotName, w, r)
			return
		}
		if err != nil {
			ctx.ReportInternalError("Fail to read info/refs", w, r)
			return
		}
		w.Write(s)
	}))
}

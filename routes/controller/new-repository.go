package controller

import (
	"fmt"
	"net/http"

	. "github.com/bctnry/gitus/routes"
	"github.com/bctnry/gitus/templates"
)

func bindNewRepositoryController(ctx *RouterContext) {
	http.HandleFunc("GET /new/repo", WithLog(func(w http.ResponseWriter, r *http.Request) {
		if ctx.Config.PlainMode { FoundAt(w, "/"); return }
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !loginInfo.LoggedIn { FoundAt(w, "/"); return }
		l, err := ctx.DatabaseInterface.GetAllNamespaceByOwner(loginInfo.UserName)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		LogTemplateError(ctx.LoadTemplate("new-repository").Execute(w, templates.NewRepositoryTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			NamespaceList: l,
			Selected: r.URL.Query().Get("ns"),
		}))
	}))
	http.HandleFunc("POST /new/repo", WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !loginInfo.LoggedIn { FoundAt(w, "/"); return }
		err = r.ParseForm()
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		userName := loginInfo.UserName
		newRepoNS := r.Form.Get("namespace")
		ns, err := ctx.DatabaseInterface.GetNamespaceByName(newRepoNS)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if ns.Owner != userName {
			ctx.ReportForbidden("Not owner", w, r)
			return
		}
		fmt.Println("reached!")
		newRepoName := r.Form.Get("name")
		newRepoDescription := r.Form.Get("description")
		repo, err := ctx.DatabaseInterface.CreateRepository(newRepoNS, newRepoName)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		repo.Description = newRepoDescription
		// NOTE: we ignore this error since we have the repository already.
		ctx.DatabaseInterface.UpdateRepositoryInfo(newRepoNS, newRepoName, repo)
		FoundAt(w, fmt.Sprintf("/repo/%s:%s", newRepoNS, newRepoName))
	}))
}


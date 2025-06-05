package controller

import (
	"fmt"
	"net/http"

	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

// since we've decided that admin have full power, should we list all
// namespace when the user is an admin? the answer should be no -
// admin should go to the respective page to find the new repository
// link when the namespace does not explicitly have them as a member.
func bindNewRepositoryController(ctx *RouterContext) {
	http.HandleFunc("GET /new/repo", WithLog(func(w http.ResponseWriter, r *http.Request) {
		if ctx.Config.PlainMode { FoundAt(w, "/"); return }
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !loginInfo.LoggedIn { FoundAt(w, "/"); return }
		l, err := ctx.DatabaseInterface.GetAllComprisingNamespace(loginInfo.UserName)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		LogTemplateError(ctx.LoadTemplate("new/repository").Execute(w, templates.NewRepositoryTemplateModel{
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
		if !loginInfo.LoggedIn {
			ctx.ReportRedirect("/", 3, "Not Logged In", "You need to log in before creating a new repository", w, r)
			return
		}
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
		isOwner := ns.Owner == userName
		priv := ns.ACL.GetUserPrivilege(loginInfo.UserName)
		isPrivilegedMember := priv != nil && priv.AddRepository
		if !loginInfo.IsAdmin && !isOwner && !isPrivilegedMember {
			ctx.ReportForbidden("Not enough privilege", w, r)
			return
		}
		newRepoName := r.Form.Get("name")
		newRepoDescription := r.Form.Get("description")
		repo, err := ctx.DatabaseInterface.CreateRepository(newRepoNS, newRepoName, userName)
		fmt.Println(repo)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		repo.Owner = userName
		repo.Description = newRepoDescription
		// NOTE: we ignore this error since we have the repository already.
		ctx.DatabaseInterface.UpdateRepositoryInfo(newRepoNS, newRepoName, repo)
		FoundAt(w, fmt.Sprintf("/repo/%s:%s", newRepoNS, newRepoName))
	}))
}


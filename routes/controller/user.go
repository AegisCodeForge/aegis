package controller

import (
	"net/http"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/db"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)


func bindUserController(ctx *RouterContext) {
	http.HandleFunc("GET /u/{userName}", WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if ctx.Config.PlainMode || !CheckGlobalVisibleToUser(ctx, loginInfo) {
			switch ctx.Config.GlobalVisibility {
			case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
				FoundAt(w, "/maintenance-notice")
				return
			case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
				FoundAt(w, "/shutdown-notice")
				return
			case aegis.GLOBAL_VISIBILITY_PRIVATE:
				if !ctx.Config.PlainMode {
					FoundAt(w, "/login")
				} else {
					FoundAt(w, "/private-notice")
				}
				return
			}
		}
		un := r.PathValue("userName")
		user, err := ctx.DatabaseInterface.GetUserByName(un)
		if err == db.ErrEntityNotFound {
			ctx.ReportNotFound(un, "User", "depot", w, r)
			return
		}
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		nsList, err := ctx.DatabaseInterface.GetAllBelongingNamespace(loginInfo.UserName, un)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		repoList, err := ctx.DatabaseInterface.GetAllBelongingRepository(loginInfo.UserName, un, 0, 0)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		LogTemplateError(ctx.LoadTemplate("user").Execute(w, templates.UserTemplateModel{
			User: user,
			RepositoryList: repoList,
			Config: ctx.Config,
			LoginInfo: loginInfo,
			BelongingNamespaceList: nsList,
		}))
	}))
}

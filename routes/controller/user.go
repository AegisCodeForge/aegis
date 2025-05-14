package controller

import (
	"net/http"

	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
	"github.com/bctnry/aegis/pkg/aegis/db"
)


func bindUserController(ctx *RouterContext) {
	http.HandleFunc("GET /u/{userName}", WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		un := r.PathValue("userName")
		user, err := ctx.DatabaseInterface.GetUserByName(un)
		if db.IsAegisDatabaseError(err) {
			if err.(*db.AegisDatabaseError).ErrorType == db.ENTITY_NOT_FOUND {
				ctx.ReportNotFound(un, "User", "depot", w, r)
				return
			}
		}
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		ls, err := ctx.DatabaseInterface.GetAllRepositoryFromNamespace(un)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		LogTemplateError(ctx.LoadTemplate("user").Execute(w, templates.UserTemplateModel{
			User: user,
			RepositoryList: ls,
			Config: ctx.Config,
			LoginInfo: loginInfo,
		}))
	}))
}

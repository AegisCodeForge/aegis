package controller

import (
	"fmt"
	"net/http"

	. "github.com/bctnry/gitus/routes"
	"github.com/bctnry/gitus/templates"
	"github.com/bctnry/gitus/pkg/gitus/db"
)


func bindUserController(ctx *RouterContext) {
	http.HandleFunc("GET /u/{userName}", WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		un := r.PathValue("userName")
		user, err := ctx.DatabaseInterface.GetUserByName(un)
		fmt.Println(user, err)
		if db.IsGitusDatabaseError(err) {
			if err.(*db.GitusDatabaseError).ErrorType == db.ENTITY_NOT_FOUND {
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

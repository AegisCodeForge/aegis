package admin

import (
	"net/http"

	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindAdminIndexController(ctx *routes.RouterContext) {
	http.HandleFunc("GET /admin", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		u, err := ctx.DatabaseInterface.GetUserByName(loginInfo.UserName)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !(u.Status == model.ADMIN || u.Status == model.SUPER_ADMIN) {
			routes.FoundAt(w, "/")
			return
		}
		routes.LogTemplateError(ctx.LoadTemplate("admin/index").Execute(w, templates.AdminIndexTemplateModel{
			Config: ctx.Config,
		}))
		
	}))
}

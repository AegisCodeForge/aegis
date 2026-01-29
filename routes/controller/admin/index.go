package admin

import (
	"net/http"

	. "github.com/GitusCodeForge/Gitus/routes"
	"github.com/GitusCodeForge/Gitus/templates"
)

func bindAdminIndexController(ctx *RouterContext) {
	http.HandleFunc("GET /admin", UseMiddleware(
		[]Middleware{Logged, LoginRequired, AdminRequired,
			GlobalVisibility, ErrorGuard,
		}, ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			LogTemplateError(rc.LoadTemplate("admin/index").Execute(w, templates.AdminIndexTemplateModel{
				Config: rc.Config,
				LoginInfo: rc.LoginInfo,
			}))
		},
	))
}


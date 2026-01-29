package rrdoc

import (
	"net/http"

	. "github.com/GitusCodeForge/Gitus/routes"
	"github.com/GitusCodeForge/Gitus/templates"
)

func bindAdminRRDocListController(ctx *RouterContext) {
	http.HandleFunc("GET /admin/rrdoc", UseMiddleware(
		[]Middleware{
			Logged, LoginRequired, AdminRequired,
			GlobalVisibility, ErrorGuard,
		}, ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "text/html")
			w.WriteHeader(200)
			LogTemplateError(rc.LoadTemplate("admin/rrdoc/list").Execute(w, &templates.AdminRRDocListTemplateModel{
				Config: rc.Config,
				LoginInfo: rc.LoginInfo,
			}))
		},
	))
}


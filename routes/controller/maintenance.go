package controller

import (
	"net/http"

	. "github.com/GitusCodeForge/Gitus/routes"
	"github.com/GitusCodeForge/Gitus/templates"
)

func bindMaintenanceNoticeController(ctx *RouterContext) {
	http.HandleFunc("GET /maintenance-notice", UseMiddleware(
		[]Middleware{Logged, ErrorGuard}, ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			loginInfo, _ := GenerateLoginInfoModel(ctx, r)
			LogTemplateError(ctx.LoadTemplate("maintenance-notice").Execute(w, &templates.MaintenanceNoticeTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				Message: ctx.Config.MaintenanceMessage,
			}))
		},
	))
}


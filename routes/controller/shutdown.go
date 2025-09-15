package controller

import (
	"net/http"

	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)


func bindShutdownNoticeController(ctx *RouterContext) {
	http.HandleFunc("GET /shutdown-notice", UseMiddleware(
		[]Middleware{Logged, ErrorGuard}, ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			LogTemplateError(rc.LoadTemplate("shutdown-notice").Execute(w, &templates.ShutdownNoticeTemplateModel{
				Config: rc.Config,
				LoginInfo: rc.LoginInfo,
				Message: rc.Config.ShutdownMessage,
			}))
		},
	))
}


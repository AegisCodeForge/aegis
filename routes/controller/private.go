package controller

import (
	"net/http"

	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)


func bindPrivateNoticeController(ctx *RouterContext) {
	http.HandleFunc("GET /private-notice", UseMiddleware(
		[]Middleware{Logged, UseLoginInfo, ErrorGuard}, ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			LogTemplateError(rc.LoadTemplate("private-notice").Execute(w, &templates.PrivateNoticeTemplateModel{
				Config: rc.Config,
				LoginInfo: rc.LoginInfo,
				Message: rc.Config.PrivateNoticeMessage,
			}))
		},
	))
}


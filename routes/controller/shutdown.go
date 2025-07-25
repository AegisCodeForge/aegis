package controller

import (
	"net/http"

	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)


func bindShutdownNoticeController(ctx *RouterContext) {
	http.HandleFunc("GET /shutdown-notice", WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, _ := GenerateLoginInfoModel(ctx, r)
		LogTemplateError(ctx.LoadTemplate("shutdown-notice").Execute(w, &templates.ShutdownNoticeTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			Message: ctx.Config.ShutdownMessage,
		}))
	}))
}


package controller

import (
	"net/http"

	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)


func bindPrivateNoticeController(ctx *RouterContext) {
	http.HandleFunc("GET /private-notice", WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, _ := GenerateLoginInfoModel(ctx, r)
		LogTemplateError(ctx.LoadTemplate("private-notice").Execute(w, &templates.PrivateNoticeTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			Message: ctx.Config.PrivateNoticeMessage,
		}))
	}))
}


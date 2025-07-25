package controller

import (
	"net/http"

	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)


func bindMaintenanceNoticeController(ctx *RouterContext) {
	http.HandleFunc("GET /maintenance-notice", WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, _ := GenerateLoginInfoModel(ctx, r)
		LogTemplateError(ctx.LoadTemplate("maintenance-notice").Execute(w, &templates.MaintenanceNoticeTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			Message: ctx.Config.MaintenanceMessage,
		}))
	}))
}


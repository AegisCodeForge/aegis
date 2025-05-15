package admin

import (
	"fmt"
	"net/http"

	"github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindAdminSessionSettingController(ctx *routes.RouterContext) {
	http.HandleFunc("GET /admin/session-setting", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		routes.LogTemplateError(ctx.LoadTemplate("admin/session-setting").Execute(w, &templates.AdminConfigTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			ErrorMsg: "",
		}))
	}))
	http.HandleFunc("POST /admin/session-setting", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		err = r.ParseForm()
		if err != nil {
			routes.LogTemplateError(ctx.LoadTemplate("admin/session-setting").Execute(w, &templates.AdminConfigTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: fmt.Sprintf("Error while parsing request: %s. Please contact site owner for this...", err.Error()),
			}))
			return
		}
		ctx.Config.SessionType = r.Form.Get("type")
		ctx.Config.SessionPath = r.Form.Get("path")
		ctx.Config.SessionURL = r.Form.Get("url")
		err = ctx.Config.Sync()
		if err != nil {
			routes.LogTemplateError(ctx.LoadTemplate("admin/session-setting").Execute(w, &templates.AdminConfigTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: fmt.Sprintf("Error while saving config: %s. Please contact site owner for this...", err.Error()),
			}))
			return
		}
		routes.LogTemplateError(ctx.LoadTemplate("admin/db-setting").Execute(w, &templates.AdminConfigTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			ErrorMsg: fmt.Sprintf("Updated."),
		}))
	}))
}


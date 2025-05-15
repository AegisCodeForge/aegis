package admin

import (
	"fmt"
	"net/http"

	"github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindAdminDatabaseSettingController(ctx *routes.RouterContext) {
	http.HandleFunc("GET /admin/db-setting", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		routes.LogTemplateError(ctx.LoadTemplate("admin/db-setting").Execute(w, &templates.AdminConfigTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			ErrorMsg: "",
		}))
	}))
	http.HandleFunc("POST /admin/db-setting", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		err = r.ParseForm()
		if err != nil {
			routes.LogTemplateError(ctx.LoadTemplate("admin/db-setting").Execute(w, &templates.AdminConfigTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: fmt.Sprintf("Error while parsing request: %s. Please contact site owner for this...", err.Error()),
			}))
			return
		}
		ctx.Config.DatabaseType = r.Form.Get("type")
		ctx.Config.DatabasePath = r.Form.Get("path")
		ctx.Config.DatabaseURL = r.Form.Get("url")
		ctx.Config.DatabaseName = r.Form.Get("name")
		ctx.Config.DatabaseUser = r.Form.Get("user")
		ctx.Config.DatabasePassword = r.Form.Get("password")
		ctx.Config.DatabaseTablePrefix = r.Form.Get("table-prefix")
		err = ctx.Config.Sync()
		if err != nil {
			routes.LogTemplateError(ctx.LoadTemplate("admin/db-setting").Execute(w, &templates.AdminConfigTemplateModel{
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


package admin

import (
	"fmt"
	"net/http"

	"github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindAdminReceiptSystemSettingController(ctx *routes.RouterContext) {
	http.HandleFunc("GET /admin/rs-setting", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		routes.LogTemplateError(ctx.LoadTemplate("admin/rs-setting").Execute(w, &templates.AdminConfigTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			ErrorMsg: "",
		}))
	}))
	http.HandleFunc("POST /admin/rs-setting", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		err = r.ParseForm()
		if err != nil {
			routes.LogTemplateError(ctx.LoadTemplate("admin/rs-setting").Execute(w, &templates.AdminConfigTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: fmt.Sprintf("Error while parsing request: %s. Please contact site owner for this...", err.Error()),
			}))
			return
		}
		ctx.Config.ReceiptSystem.Type = r.Form.Get("type")
		ctx.Config.ReceiptSystem.Path = r.Form.Get("path")
		ctx.Config.ReceiptSystem.URL = r.Form.Get("url")
		ctx.Config.ReceiptSystem.UserName = r.Form.Get("username")
		ctx.Config.ReceiptSystem.Password = r.Form.Get("password")
		ctx.Config.ReceiptSystem.TablePrefix = r.Form.Get("table-prefix")
		err = ctx.Config.Sync()
		if err != nil {
			ctx.ReportRedirect("/admin/rs-setting", 0, "Internal Error", fmt.Sprintf("Error while saving config: %s. Please contact site owner for this...", err.Error()), w, r)
			return
		}
		ctx.ReportRedirect("/admin/rs-setting", 3, "Updated", "Configuration is updated.", w, r)
	}))
}


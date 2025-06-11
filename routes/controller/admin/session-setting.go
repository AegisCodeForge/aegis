package admin

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

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
		ctx.Config.Session.Type = r.Form.Get("type")
		ctx.Config.Session.Path = r.Form.Get("path")
		ctx.Config.Session.TablePrefix = r.Form.Get("table-prefix")
		ctx.Config.Session.Host = r.Form.Get("host")
		ctx.Config.Session.UserName = r.Form.Get("user-name")
		ctx.Config.Session.Password = r.Form.Get("password")
		dbnStr := strings.TrimSpace(r.Form.Get("database-number"))
		if dbnStr == "" { dbnStr = "0" }
		dbn, err := strconv.Atoi(dbnStr)
		if err != nil {
			ctx.ReportNormalError("Invalid request", w, r)
			return
		}
		ctx.Config.Session.DatabaseNumber = dbn
		err = ctx.Config.Sync()
		if err != nil {
			ctx.ReportRedirect("/admin/session-setting", 0, "Internal Error", fmt.Sprintf("Error while saving config: %s. Please contact site owner for this...", err.Error()), w, r)
			return
		}
		ctx.ReportRedirect("/admin/session-setting", 3, "Updated", "Configuration updated.", w, r)
	}))
}


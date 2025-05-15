package admin

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindAdminMailerSettingController(ctx *routes.RouterContext) {
	http.HandleFunc("GET /admin/mailer-setting", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		routes.LogTemplateError(ctx.LoadTemplate("admin/mailer-setting").Execute(w, &templates.AdminConfigTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			ErrorMsg: "",
		}))
	}))
	http.HandleFunc("POST /admin/mailer-setting", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		err = r.ParseForm()
		if err != nil {
			routes.LogTemplateError(ctx.LoadTemplate("admin/mailer-setting").Execute(w, &templates.AdminConfigTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: fmt.Sprintf("Error while parsing request: %s. Please contact site owner for this...", err.Error()),
			}))
			return
		}
		ctx.Config.Mailer.Type = r.Form.Get("type")
		ctx.Config.Mailer.SMTPServer = r.Form.Get("server")
		i, err := strconv.ParseInt(r.Form.Get("port"), 10, 32)
		if err != nil {
			routes.LogTemplateError(ctx.LoadTemplate("admin/mailer-setting").Execute(w, &templates.AdminConfigTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: fmt.Sprintf("Invalid format for SMTP server port: %s.", err.Error()),
			}))
			return
		}
		ctx.Config.Mailer.SMTPPort = int(i)
		ctx.Config.Mailer.User = r.Form.Get("user")
		ctx.Config.Mailer.Password = r.Form.Get("password")
		err = ctx.Config.Sync()
		if err != nil {
			routes.LogTemplateError(ctx.LoadTemplate("admin/mailer-setting").Execute(w, &templates.AdminConfigTemplateModel{
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


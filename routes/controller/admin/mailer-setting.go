package admin

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/mail"
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
		if r.Form.Get("action") == "Test Mailer" {
			port, err := strconv.ParseInt(r.Form.Get("port"), 10, 64)
			if err != nil {
				routes.LogTemplateError(ctx.LoadTemplate("admin/mailer-setting").Execute(w, &templates.AdminConfigTemplateModel{
					Config: ctx.Config,
					LoginInfo: loginInfo,
					ErrorMsg: fmt.Sprintf("Invalid format for SMTP server port: %s.", err.Error()),
				}))
				return
			}
			mailer, err := mail.CreateMailerFromMailerConfig(&aegis.AegisMailerConfig{
				Type: r.Form.Get("type"),
				SMTPServer: r.Form.Get("server"),
				SMTPPort: int(port),
				User: r.Form.Get("username"),
				Password: r.Form.Get("password"),
			})
			if err != nil {
				routes.LogTemplateError(ctx.LoadTemplate("admin/mailer-setting").Execute(w, &templates.AdminConfigTemplateModel{
					Config: ctx.Config,
					LoginInfo: loginInfo,
					ErrorMsg: fmt.Sprintf("Failed to create mailer: %s.", err.Error()),
				}))
				return
			}
			fmt.Println(r.Form.Get("test-email-target"))
			err = mailer.SendPlainTextMail(r.Form.Get("test-email-target"), "Mailer Configuration Test", fmt.Sprintf(`
This is a test email from %s.

If you can see this message it means the mailer can be used normally.
`, ctx.Config.DepotName))
			if err != nil {
				routes.LogTemplateError(ctx.LoadTemplate("admin/mailer-setting").Execute(w, &templates.AdminConfigTemplateModel{
					Config: ctx.Config,
					LoginInfo: loginInfo,
					ErrorMsg: fmt.Sprintf("Failed to send test email: %s.", err.Error()),
				}))
				return
			}
			routes.LogTemplateError(ctx.LoadTemplate("admin/mailer-setting").Execute(w, &templates.AdminConfigTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: "Test email has been sent.",
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
		routes.LogTemplateError(ctx.LoadTemplate("admin/mailer-setting").Execute(w, &templates.AdminConfigTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			ErrorMsg: "Updated.",
		}))
	}))
}


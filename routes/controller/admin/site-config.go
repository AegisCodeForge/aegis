package admin

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindAdminSiteConfigController(ctx *routes.RouterContext) {
	http.HandleFunc("GET /admin/site-config", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		routes.LogTemplateError(ctx.LoadTemplate("admin/site-config").Execute(w, &templates.AdminConfigTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			ErrorMsg: "",
		}))
	}))
	http.HandleFunc("POST /admin/site-config", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		err = r.ParseForm()
		if err != nil {
			routes.LogTemplateError(ctx.LoadTemplate("admin/site-config").Execute(w, &templates.AdminConfigTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: fmt.Sprintf("Error while parsing request: %s. Please contact site owner for this...", err.Error()),
			}))
			return
		}
		switch r.Form.Get("section") {
		case "web":
			ctx.Config.HttpHostName = r.Form.Get("http-host-name")
			ctx.Config.SshHostName = r.Form.Get("ssh-host-name")
			ctx.Config.StaticAssetDirectory = r.Form.Get("static-asset-directory")
			ctx.Config.BindAddress = r.Form.Get("bind-address")
			i, err := strconv.ParseInt(r.Form.Get("bind-port"), 10, 32)
			if err != nil {
				routes.LogTemplateError(ctx.LoadTemplate("admin/site-config").Execute(w, &templates.AdminConfigTemplateModel{
					Config: ctx.Config,
					LoginInfo: loginInfo,
					ErrorMsg: fmt.Sprintf("Error while parsing bind port: %s. ", err.Error()),
				}))
				return
			}
			ctx.Config.BindPort = int(i)
			err = ctx.Config.Sync()
			if err != nil {
				routes.LogTemplateError(ctx.LoadTemplate("admin/site-config").Execute(w, &templates.AdminConfigTemplateModel{
					Config: ctx.Config,
					LoginInfo: loginInfo,
					ErrorMsg: fmt.Sprintf("Error while parsing request: %s. Please contact site owner for this...", err.Error()),
				}))
				return
			}
			routes.LogTemplateError(ctx.LoadTemplate("admin/site-config").Execute(w, &templates.AdminConfigTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: "Updated.",
			}))
		case "basic":
			ctx.Config.DepotName = r.Form.Get("depot-name")
			ctx.Config.GitRoot = r.Form.Get("root")
			ctx.Config.GitUser = r.Form.Get("git-user")
			ctx.Config.UseNamespace = false
			if r.Form.Has("use-namespace") && r.Form.Get("use-namespace") == "on" {
				ctx.Config.UseNamespace = true
			}
			ctx.Config.AllowRegistration = false
			if r.Form.Has("allow-registration") && r.Form.Get("allow-registration") == "on" {
				ctx.Config.AllowRegistration = true
			}
			ctx.Config.EmailConfirmationRequired = false
			if r.Form.Has("email-confirmation-required") && r.Form.Get("email-confirmation-required") == "on" {
				ctx.Config.UseNamespace = true
			}
			ctx.Config.ManualApproval = false
			if r.Form.Has("manual-approval") && r.Form.Get("manual-approval") == "on" {
				ctx.Config.UseNamespace = true
			}
			err = ctx.Config.Sync()
			if err != nil {
				routes.LogTemplateError(ctx.LoadTemplate("admin/site-config").Execute(w, &templates.AdminConfigTemplateModel{
					Config: ctx.Config,
					LoginInfo: loginInfo,
					ErrorMsg: fmt.Sprintf("Error while saving config: %s. Please contact site owner for this...", err.Error()),
				}))
				return
			}
			routes.LogTemplateError(ctx.LoadTemplate("admin/site-config").Execute(w, &templates.AdminConfigTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: "Updated.",
			}))
		}
	}))
}


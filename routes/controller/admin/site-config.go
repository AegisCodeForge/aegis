package admin

import (
	"fmt"
	"net/http"
	"strconv"

	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindAdminSiteConfigController(ctx *RouterContext) {
	http.HandleFunc("GET /admin/site-config", UseMiddleware(
		[]Middleware{Logged, LoginRequired, AdminRequired,
			GlobalVisibility, ErrorGuard,
		}, ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			LogTemplateError(rc.LoadTemplate("admin/site-config").Execute(w, &templates.AdminConfigTemplateModel{
				Config: rc.Config,
				LoginInfo: rc.LoginInfo,
				ErrorMsg: "",
			}))
		},
	))

	http.HandleFunc("POST /admin/site-config", UseMiddleware(
		[]Middleware{Logged, ValidPOSTRequestRequired,
			LoginRequired, AdminRequired,
			GlobalVisibility, ErrorGuard,
		}, ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			rc.Config.LockForSync()
			defer rc.Config.Unlock()
			switch r.Form.Get("section") {
			case "web":
				rc.Config.HttpHostName = r.Form.Get("http-host-name")
				rc.Config.SshHostName = r.Form.Get("ssh-host-name")
				rc.Config.StaticAssetDirectory = r.Form.Get("static-asset-directory")
				rc.Config.BindAddress = r.Form.Get("bind-address")
				i, err := strconv.ParseInt(r.Form.Get("bind-port"), 10, 32)
				if err != nil {
					LogTemplateError(rc.LoadTemplate("admin/site-config").Execute(w, &templates.AdminConfigTemplateModel{
						Config: rc.Config,
						LoginInfo: rc.LoginInfo,
						ErrorMsg: fmt.Sprintf("Error while parsing bind port: %s. ", err.Error()),
					}))
					return
				}
				rc.Config.BindPort = int(i)
				err = rc.Config.Sync()
				if err != nil {
					LogTemplateError(rc.LoadTemplate("admin/site-config").Execute(w, &templates.AdminConfigTemplateModel{
						Config: rc.Config,
						LoginInfo: rc.LoginInfo,
						ErrorMsg: fmt.Sprintf("Error while parsing request: %s. Please contact site owner for this...", err.Error()),
					}))
					return
				}
				LogTemplateError(rc.LoadTemplate("admin/site-config").Execute(w, &templates.AdminConfigTemplateModel{
					Config: rc.Config,
					LoginInfo: rc.LoginInfo,
					ErrorMsg: "Updated.",
				}))
			case "basic":
				rc.Config.DepotName = r.Form.Get("depot-name")
				rc.Config.GitRoot = r.Form.Get("root")
				rc.Config.GitUser = r.Form.Get("git-user")
				rc.Config.UseNamespace = false
				if r.Form.Has("use-namespace") && r.Form.Get("use-namespace") == "on" {
					rc.Config.UseNamespace = true
				}
				rc.Config.AllowRegistration = false
				if r.Form.Has("allow-registration") && r.Form.Get("allow-registration") == "on" {
					rc.Config.AllowRegistration = true
				}
				rc.Config.EmailConfirmationRequired = false
				if r.Form.Has("email-confirmation-required") && r.Form.Get("email-confirmation-required") == "on" {
					rc.Config.EmailConfirmationRequired = true
				}
				rc.Config.ManualApproval = false
				if r.Form.Has("manual-approval") && r.Form.Get("manual-approval") == "on" {
					rc.Config.ManualApproval = true
				}
				rc.Config.RecalculateProperPath()
				err := rc.Config.Sync()
				if err != nil {
					LogTemplateError(rc.LoadTemplate("admin/site-config").Execute(w, &templates.AdminConfigTemplateModel{
						Config: rc.Config,
						LoginInfo: rc.LoginInfo,
						ErrorMsg: fmt.Sprintf("Error while saving config: %s. Please contact site owner for this...", err.Error()),
					}))
					return
				}
				LogTemplateError(rc.LoadTemplate("admin/site-config").Execute(w, &templates.AdminConfigTemplateModel{
					Config: rc.Config,
					LoginInfo: rc.LoginInfo,
					ErrorMsg: "Updated.",
				}))
			}
		},
	))
}


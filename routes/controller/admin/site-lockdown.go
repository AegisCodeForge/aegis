package admin

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindAdminSiteLockdownController(ctx *routes.RouterContext) {
	http.HandleFunc("GET /admin/site-lockdown", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		routes.LogTemplateError(ctx.LoadTemplate("admin/site-lockdown").Execute(w, &templates.AdminSiteLockdownTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			CurrentMode: ctx.Config.GlobalVisibility,
			PrivateNoticeMessage: ctx.Config.PrivateNoticeMessage,
			ShutdownNoticeMessage: ctx.Config.ShutdownMessage,
			FullAccessUser: strings.Join(ctx.Config.FullAccessUser, ","),
			MaintenanceNoticeMessage: ctx.Config.MaintenanceMessage,
		}))
	}))
	
	http.HandleFunc("POST /admin/site-lockdown", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		err = r.ParseForm()
		if err != nil {
			ctx.ReportRedirect("/admin/site-lockdown", 5, "Invalid Request", "Invalid request. Please try again.", w, r)
			return
		}
		t := strings.TrimSpace(r.Form.Get("type"))
		switch t {
		case "public":
			ctx.Config.GlobalVisibility = aegis.GLOBAL_VISIBILITY_PUBLIC
		case "private":
			ctx.Config.GlobalVisibility = aegis.GLOBAL_VISIBILITY_PRIVATE
			ctx.Config.PrivateNoticeMessage = r.Form.Get("private-notice-message")
		case "shutdown":
			ctx.Config.GlobalVisibility = aegis.GLOBAL_VISIBILITY_SHUTDOWN
			ul := make([]string, 0)
			for k := range strings.SplitSeq(r.Form.Get("full-access-user"), ",") {
				ul = append(ul, strings.TrimSpace(k))
			}
			ctx.Config.FullAccessUser = ul
			ctx.Config.ShutdownMessage = r.Form.Get("shutdown-notice-message")
		case "maintenance":
			ctx.Config.GlobalVisibility = aegis.GLOBAL_VISIBILITY_MAINTENANCE
			ctx.Config.MaintenanceMessage = r.Form.Get("maintenance-notice-message")
		}
		err = ctx.Config.Sync()
		if err != nil {
			ctx.ReportRedirect("/admin/site-lockdown", 0, "Internal Error", fmt.Sprintf("Failed to save config due to error: %s", err.Error()), w, r)
			return
		}
		ctx.ReportRedirect("/admin/site-lockdown", 3, "Configuration Saved", "Configuration saved.", w, r)
	}))
}


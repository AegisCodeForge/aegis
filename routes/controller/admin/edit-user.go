package admin

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
	"golang.org/x/crypto/bcrypt"
)


func bindAdminEditUserController(ctx *routes.RouterContext) {
	http.HandleFunc("GET /admin/user/{username}/edit", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		un := r.PathValue("username")
		u, err := ctx.DatabaseInterface.GetUserByName(un)
		if err != nil {
			routes.LogTemplateError(ctx.LoadTemplate("admin/_redirect-with-message").Execute(w, &templates.AdminRedirectWithMessageModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: "",
				Timeout: 0,
				RedirectUrl: "/admin/user-list",
				MessageTitle: "Error",
				MessageText: fmt.Sprintf("Failed to fetch user: %s", err.Error()),
			}))
			return
		}
		if !loginInfo.IsSuperAdmin && u.Status == model.SUPER_ADMIN {
			routes.LogTemplateError(ctx.LoadTemplate("admin/_redirect-with-message").Execute(w, &templates.AdminRedirectWithMessageModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: "",
				Timeout: 3,
				RedirectUrl: "/admin/user-list",
				MessageTitle: "Error",
				MessageText: "Not enough permission.",
			}))
			return
		}
		routes.LogTemplateError(ctx.LoadTemplate("admin/user-edit").Execute(w, &templates.AdminUserEditTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			User: u,
		}))
		return
	}))
 	http.HandleFunc("POST /admin/user/{username}/edit", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		un := r.PathValue("username")
		user, err := ctx.DatabaseInterface.GetUserByName(un)
		if !loginInfo.IsSuperAdmin && user.Status == model.SUPER_ADMIN {
			routes.LogTemplateError(ctx.LoadTemplate("admin/_redirect-with-message").Execute(w, &templates.AdminRedirectWithMessageModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: "",
				Timeout: 3,
				RedirectUrl: "/admin/user-list",
				MessageTitle: "Error",
				MessageText: "Not enough permission.",
			}))
			return
		}
		err = r.ParseForm()
		if err != nil {
			routes.LogTemplateError(ctx.LoadTemplate("admin/_redirect-with-message").Execute(w, &templates.AdminRedirectWithMessageModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: "",
				Timeout: 3,
				RedirectUrl: "/admin/user-list",
				MessageTitle: "Error",
				MessageText: "Invalid request.",
			}))
			return
		}
		switch r.Form.Get("type") {
		case "info":
			if len(r.Form.Get("title")) > 0 { user.Title = r.Form.Get("title") }
			if len(r.Form.Get("email")) > 0 { user.Email = r.Form.Get("email") }
			if len(r.Form.Get("website")) > 0 { user.Website = r.Form.Get("website") }
			if len(r.Form.Get("bio")) > 0 { user.Bio = r.Form.Get("bio") }
			i, err := strconv.ParseInt(r.Form.Get("status"), 10, 32)
			if !loginInfo.IsSuperAdmin && model.AegisUserStatus(i) != model.SUPER_ADMIN {
				routes.LogTemplateError(ctx.LoadTemplate("admin/_redirect-with-message").Execute(w, &templates.AdminRedirectWithMessageModel{
					Config: ctx.Config,
					LoginInfo: loginInfo,
					ErrorMsg: "",
					Timeout: 3,
					RedirectUrl: "/admin/user-list",
					MessageTitle: "Error",
					MessageText: "Not enough permission.",
				}))
				return
			}
			if err != nil {
				routes.LogTemplateError(ctx.LoadTemplate("admin/_redirect-with-message").Execute(w, &templates.AdminRedirectWithMessageModel{
					Config: ctx.Config,
					LoginInfo: loginInfo,
					ErrorMsg: "",
					Timeout: 3,
					RedirectUrl: "/admin/user-list",
					MessageTitle: "Error",
					MessageText: "Invalid format for user status.",
				}))
				return
			}
			user.Status = model.AegisUserStatus(i)
			err = ctx.DatabaseInterface.UpdateUserInfo(un, user)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
		case "password":
			// we will have confirm check at the frontend; this is
			// here for the people who disabled javascript.
			if r.Form.Get("new-password") != r.Form.Get("confirm-new-password") {
				routes.LogTemplateError(ctx.LoadTemplate("setting-user-info").Execute(w, templates.AdminUserEditTemplateModel{
					User: user,
					Config: ctx.Config,
					LoginInfo: loginInfo,
					ErrorMsg: struct{Type string; Message string}{
						Type: r.Form.Get("type"),
						Message: "New password mismatch",
					},
				}))
				return
			}
			err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(r.Form.Get("old-password")))
			if err == bcrypt.ErrMismatchedHashAndPassword {
				routes.LogTemplateError(ctx.LoadTemplate("setting-user-info").Execute(w, templates.AdminUserEditTemplateModel{
					User: user,
					Config: ctx.Config,
					LoginInfo: loginInfo,
					ErrorMsg: struct{Type string; Message string}{
						Type: r.Form.Get("type"),
						Message: "Wrong old password",
					},
				}))
				return
			}
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			newpwh, err := bcrypt.GenerateFromPassword([]byte(r.Form.Get("new-password")), bcrypt.DefaultCost)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			ctx.DatabaseInterface.UpdateUserPassword(un, string(newpwh))
		}
		routes.LogTemplateError(ctx.LoadTemplate("admin/user-edit").Execute(w, templates.AdminUserEditTemplateModel{
			User: user,
			Config: ctx.Config,
			LoginInfo: loginInfo,
			ErrorMsg: struct{Type string; Message string}{
				Type: r.Form.Get("type"),
				Message: "Updated.",
			},
		}))
	}))
}


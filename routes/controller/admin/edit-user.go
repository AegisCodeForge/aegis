package admin

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/pkg/auxfuncs"
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
			ctx.ReportRedirect("/admin/user-list", 0, "Internal Error", fmt.Sprintf("Failed to fetch user %s: %s", un, err.Error()), w, r)
			return
		}
		if !loginInfo.IsSuperAdmin && u.Status == model.SUPER_ADMIN {
			ctx.ReportRedirect("/admin/user-list", 3, "Error", "Your account does not have enough privilege for this action.", w, r)
			return
		}
		routes.LogTemplateError(ctx.LoadTemplate("admin/user/edit").Execute(w, &templates.AdminUserEditTemplateModel{
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
		if err != nil {
			ctx.ReportRedirect("/admin/user-list", 0, "Internal Error", fmt.Sprintf("Failed to fetch user %s: %s", un, err.Error()), w, r)
			return
		}
		if !loginInfo.IsSuperAdmin && user.Status == model.SUPER_ADMIN {
			ctx.ReportRedirect("/admin/user-list", 3, "Error", "Your account does not have enough privilege for this action.", w, r)
			return
		}
		err = r.ParseForm()
		if err != nil {
			ctx.ReportNormalError("Invalid request.", w, r)
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
				ctx.ReportRedirect("/admin/user-list", 0, "Error", "Not enough permission.", w, r)
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
		routes.LogTemplateError(ctx.LoadTemplate("admin/user/edit").Execute(w, templates.AdminUserEditTemplateModel{
			User: user,
			Config: ctx.Config,
			LoginInfo: loginInfo,
			ErrorMsg: struct{Type string; Message string}{
				Type: r.Form.Get("type"),
				Message: "Updated.",
			},
		}))
	}))
	
	http.HandleFunc("GET /admin/user/{username}/ssh", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		un := r.PathValue("username")
		u, err := ctx.DatabaseInterface.GetUserByName(un)
		if err != nil {
			ctx.ReportRedirect("/admin/user-list", 0, "Internal Error", fmt.Sprintf("Failed to fetch user %s: %s", un, err.Error()), w, r)
			return
		}
		if !loginInfo.IsSuperAdmin && u.Status == model.SUPER_ADMIN {
			ctx.ReportRedirect("/admin/user-list", 3, "Error", "Your account does not have enough privilege for this action.", w, r)
			return
		}
		s, err := ctx.DatabaseInterface.GetAllAuthKeyByUsername(un)
		if err != nil {
			ctx.ReportRedirect("/admin/user-list", 0, "Internal Error", fmt.Sprintf("Failed to fetch SSH keys of user %s: %s", un, err.Error()), w, r)
			return
		}
		fmt.Println("user ", u)
		routes.LogTemplateError(ctx.LoadTemplate("admin/user/ssh-key").Execute(w, &templates.AdminUserSSHKeyTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			User: u,
			KeyList: s,
		}))
		return
	}))
	
	http.HandleFunc("GET /admin/user/{username}/ssh/{keyName}/edit", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		un := r.PathValue("username")
		u, err := ctx.DatabaseInterface.GetUserByName(un)
		if err != nil {
			ctx.ReportRedirect("/admin/user-list", 0, "Internal Error", fmt.Sprintf("Failed to fetch user %s: %s", un, err.Error()), w, r)
			return
		}
		if !loginInfo.IsSuperAdmin && u.Status == model.SUPER_ADMIN {
			ctx.ReportRedirect("/admin/user-list", 3, "Error", "Your account does not have enough privilege for this action.", w, r)
			return
		}
		kn := strings.TrimSpace(r.PathValue("keyName"))
		k, err := ctx.DatabaseInterface.GetAuthKeyByName(un, kn)
		if err != nil {
			ctx.ReportRedirect("/admin/user-list", 0, "Internal Error", fmt.Sprintf("Failed to fetch SSH key of user %s: %s", un, err.Error()), w, r)
			return
		}
		routes.LogTemplateError(ctx.LoadTemplate("admin/user/edit-ssh-key").Execute(w, &templates.AdminUserEditSSHKeyTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			User: u,
			Key: k,
		}))
		return
	}))
	
	http.HandleFunc("POST /admin/user/{username}/ssh/{keyName}/edit", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		un := r.PathValue("username")
		u, err := ctx.DatabaseInterface.GetUserByName(un)
		if err != nil {
			ctx.ReportRedirect("/admin/user-list", 0, "Internal Error", fmt.Sprintf("Failed to fetch user %s: %s", un, err.Error()), w, r)
			return
		}
		if !loginInfo.IsSuperAdmin && u.Status == model.SUPER_ADMIN {
			ctx.ReportRedirect("/admin/user-list", 3, "Error", "Your account does not have enough privilege for this action.", w, r)
			return
		}
		err = r.ParseForm()
		if err != nil {
			ctx.ReportNormalError("Invalid request", w, r)
			return
		}
		kn := strings.TrimSpace(r.PathValue("keyName"))
		ktext := r.Form.Get("key-text")
		err = ctx.DatabaseInterface.UpdateAuthKey(un, kn, ktext)
		if err != nil {
			ctx.ReportRedirect("/admin/user-list", 0, "Internal Error", fmt.Sprintf("Failed to update SSH key of user %s: %s", un, err.Error()), w, r)
			return
		}
		ctx.ReportRedirect(fmt.Sprintf("/admin/user/%s/ssh", un), 3, "Updated", "The specified SSH key has been updated.", w, r)
		return
	}))
	
	http.HandleFunc("GET /admin/user/{username}/ssh/{keyname}/delete", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		un := r.PathValue("username")
		u, err := ctx.DatabaseInterface.GetUserByName(un)
		if err != nil {
			ctx.ReportRedirect("/admin/user-list", 0, "Internal Error", fmt.Sprintf("Failed to fetch user %s: %s", un, err.Error()), w, r)
			return
		}
		if !loginInfo.IsSuperAdmin && u.Status == model.SUPER_ADMIN {
			ctx.ReportRedirect("/admin/user-list", 3, "Error", "Your account does not have enough privilege for this action.", w, r)
			return
		}
		keyname := r.PathValue("keyname")
		err = ctx.DatabaseInterface.RemoveAuthKey(un, keyname)
		if err != nil {
			ctx.ReportRedirect(fmt.Sprintf("/admin/user/%s/ssh", un), 0, "Internal Error", fmt.Sprintf("Failed to delete SSH keys of user %s: %s", un, err.Error()), w, r)
			return
		}
		ctx.SSHKeyManagingContext.RemoveAuthorizedKey(un, keyname)
		err = ctx.SSHKeyManagingContext.Sync()
		if err != nil {
			ctx.ReportRedirect(fmt.Sprintf("/admin/user/%s/ssh", un), 0, "Internal Error", fmt.Sprintf("Failed to delete SSH keys of user %s: %s", un, err.Error()), w, r)
			return
		}
		ctx.ReportRedirect(fmt.Sprintf("/admin/user/%s/ssh", un), 3, "Deleted", "The specified SSH key has been deleted.", w, r)
		return
	}))

	http.HandleFunc("POST /admin/user/{username}/ssh", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/"); return }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/"); return }
		un := loginInfo.UserName
		err = r.ParseForm()
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		keyText := strings.TrimSpace(r.Form.Get("key-text"))
		if len(strings.TrimSpace(keyText)) <= 0 {
			ctx.ReportRedirect(fmt.Sprintf("/admin/user/%s/ssh", un), 3, "Invalid Key Format", "The key text cannot be empty.", w, r)
			return
		}
		s := strings.Split(keyText, " ")
		keyName := ""
		if len(s) < 3 {
			keyName = "key_" + auxfuncs.GenSym(8)
		} else {
			keyName = s[2]
		}
		keyName = strings.TrimSpace(keyName)
		err = ctx.DatabaseInterface.RegisterAuthKey(un, keyName, keyText)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		ctx.SSHKeyManagingContext.AddAuthorizedKey(un, keyName, keyText)
		err = ctx.SSHKeyManagingContext.Sync()
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		ctx.ReportRedirect(fmt.Sprintf("/admin/user/%s/ssh", un), 3, "Updated", "The key you've provided has been added to the database.", w, r)
	}))
	
	http.HandleFunc("GET /admin/user/{username}/gpg", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		un := r.PathValue("username")
		u, err := ctx.DatabaseInterface.GetUserByName(un)
		if err != nil {
			ctx.ReportRedirect("/admin/user-list", 0, "Internal Error", fmt.Sprintf("Failed to fetch user %s: %s", un, err.Error()), w, r)
			return
		}
		if !loginInfo.IsSuperAdmin && u.Status == model.SUPER_ADMIN {
			ctx.ReportRedirect("/admin/user-list", 3, "Error", "Your account does not have enough privilege for this action.", w, r)
			return
		}
		s, err := ctx.DatabaseInterface.GetAllSignKeyByUsername(un)
		if err != nil {
			ctx.ReportRedirect("/admin/user-list", 0, "Internal Error", fmt.Sprintf("Failed to fetch SSH keys of user %s: %s", un, err.Error()), w, r)
			return
		}
		routes.LogTemplateError(ctx.LoadTemplate("admin/user/gpg-key").Execute(w, &templates.AdminUserGPGKeyTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			User: u,
			KeyList: s,
		}))
		return
	}))

	http.HandleFunc("GET /admin/user/{username}/gpg/{keyName}/edit", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		un := r.PathValue("username")
		u, err := ctx.DatabaseInterface.GetUserByName(un)
		if err != nil {
			ctx.ReportRedirect("/admin/user-list", 0, "Internal Error", fmt.Sprintf("Failed to fetch user %s: %s", un, err.Error()), w, r)
			return
		}
		if !loginInfo.IsSuperAdmin && u.Status == model.SUPER_ADMIN {
			ctx.ReportRedirect("/admin/user-list", 3, "Error", "Your account does not have enough privilege for this action.", w, r)
			return
		}
		kn := strings.TrimSpace(r.PathValue("keyName"))
		k, err := ctx.DatabaseInterface.GetSignKeyByName(un, kn)
		if err != nil {
			ctx.ReportRedirect("/admin/user-list", 0, "Internal Error", fmt.Sprintf("Failed to fetch SSH key of user %s: %s", un, err.Error()), w, r)
			return
		}
		routes.LogTemplateError(ctx.LoadTemplate("admin/user/edit-gpg-key").Execute(w, &templates.AdminUserEditGPGKeyTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			User: u,
			Key: k,
		}))
		return
	}))

	http.HandleFunc("POST /admin/user/{username}/gpg/{keyName}/edit", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		un := r.PathValue("username")
		u, err := ctx.DatabaseInterface.GetUserByName(un)
		if err != nil {
			ctx.ReportRedirect("/admin/user-list", 0, "Internal Error", fmt.Sprintf("Failed to fetch user %s: %s", un, err.Error()), w, r)
			return
		}
		if !loginInfo.IsSuperAdmin && u.Status == model.SUPER_ADMIN {
			ctx.ReportRedirect("/admin/user-list", 3, "Error", "Your account does not have enough privilege for this action.", w, r)
			return
		}
		err = r.ParseForm()
		if err != nil {
			ctx.ReportNormalError("Invalid request", w, r)
			return
		}
		kn := strings.TrimSpace(r.PathValue("keyName"))
		ktext := r.Form.Get("key-text")
		err = ctx.DatabaseInterface.UpdateSignKey(un, kn, ktext)
		if err != nil {
			ctx.ReportRedirect("/admin/user-list", 0, "Internal Error", fmt.Sprintf("Failed to update GPG key of user %s: %s", un, err.Error()), w, r)
			return
		}
		ctx.ReportRedirect(fmt.Sprintf("/admin/user/%s/gpg", un), 3, "Updated", "The specified GPG key has been updated.", w, r)
		return
	}))
	
	http.HandleFunc("GET /admin/user/{username}/gpg/{keyname}/delete", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		un := r.PathValue("username")
		u, err := ctx.DatabaseInterface.GetUserByName(un)
		if err != nil {
			ctx.ReportRedirect("/admin/user-list", 0, "Internal Error", fmt.Sprintf("Failed to fetch user %s: %s", un, err.Error()), w, r)
			return
		}
		if !loginInfo.IsSuperAdmin && u.Status == model.SUPER_ADMIN {
			ctx.ReportRedirect("/admin/user-list", 3, "Error", "Your account does not have enough privilege for this action.", w, r)
			return
		}
		keyname := r.PathValue("keyname")
		err = ctx.DatabaseInterface.RemoveSignKey(un, keyname)
		if err != nil {
			ctx.ReportRedirect(fmt.Sprintf("/admin/user/%s/ssh", un), 0, "Internal Error", fmt.Sprintf("Failed to delete SSH keys of user %s: %s", un, err.Error()), w, r)
			return
		}
		ctx.ReportRedirect(fmt.Sprintf("/admin/user/%s/gpg", un), 3, "Deleted", "The specified GPG key has been deleted.", w, r)
		return
	}))
	
	http.HandleFunc("POST /admin/user/{username}/gpg", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/"); return }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/"); return }
		un := loginInfo.UserName
		err = r.ParseForm()
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		keyText := strings.TrimSpace(r.Form.Get("key-text"))
		if len(strings.TrimSpace(keyText)) <= 0 {
			ctx.ReportRedirect(fmt.Sprintf("/admin/user/%s/gpg", un), 3, "Invalid Key Format", "The key text cannot be empty.", w, r)
			return
		}
		s := strings.Split(keyText, " ")
		keyName := ""
		if len(s) < 3 {
			keyName = "key_" + auxfuncs.GenSym(8)
		} else {
			keyName = s[2]
		}
		keyName = strings.TrimSpace(keyName)
		err = ctx.DatabaseInterface.RegisterSignKey(un, keyName, keyText)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		ctx.ReportRedirect(fmt.Sprintf("/admin/user/%s/gpg", un), 3, "Updated", "The key you've provided has been added to the database.", w, r)
	}))
}


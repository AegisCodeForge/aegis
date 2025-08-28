package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/pkg/auxfuncs"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
	"golang.org/x/crypto/bcrypt"
)


func bindSettingSSHController(ctx *RouterContext) {
	http.HandleFunc("GET /setting/ssh", WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !CheckGlobalVisibleToUser(ctx, loginInfo) {
			switch ctx.Config.GlobalVisibility {
			case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
				FoundAt(w, "/maintenance-notice")
				return
			case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
				FoundAt(w, "/shutdown-notice")
				return
			case aegis.GLOBAL_VISIBILITY_PRIVATE:
				FoundAt(w, "/login")
				return
			}
		}
		if !loginInfo.LoggedIn { FoundAt(w, "/"); return }
		un := loginInfo.UserName
		if !model.ValidUserName(un) {
			ctx.ReportNotFound(un, "User", "Depot", w, r)
			return
		}
		s, err := ctx.DatabaseInterface.GetAllAuthKeyByUsername(un)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		LogTemplateError(ctx.LoadTemplate("setting/ssh-key").Execute(w, templates.SettingSSHKeyTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			KeyList: s,
		}))
	}))
	
	http.HandleFunc("POST /setting/ssh", WithLog(func(w http.ResponseWriter, r *http.Request){
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !CheckGlobalVisibleToUser(ctx, loginInfo) {
			switch ctx.Config.GlobalVisibility {
			case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
				FoundAt(w, "/maintenance-notice")
				return
			case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
				FoundAt(w, "/shutdown-notice")
				return
			case aegis.GLOBAL_VISIBILITY_PRIVATE:
				FoundAt(w, "/login")
				return
			}
		}
		if !loginInfo.LoggedIn { FoundAt(w, "/"); return }
		un := loginInfo.UserName
		if !model.ValidUserName(un) {
			ctx.ReportNotFound(un, "User", "Depot", w, r)
			return
		}
		err = r.ParseForm()
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		keyList, err := ctx.DatabaseInterface.GetAllAuthKeyByUsername(un)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		confirmPassword := strings.TrimSpace(r.Form.Get("password"))
		u, err := ctx.DatabaseInterface.GetUserByName(un)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(confirmPassword))
		if err == bcrypt.ErrMismatchedHashAndPassword {
			LogTemplateError(ctx.LoadTemplate("setting/ssh-key").Execute(w, templates.SettingSSHKeyTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				KeyList: keyList,
				ErrorMsg: struct{Type string; Message string}{
					Type: "",
					Message: "Invalid confirmation password",
				},
			}))
			return
		}
		keyText := strings.TrimSpace(r.Form.Get("key-text"))
		if len(strings.TrimSpace(keyText)) <= 0 {
			LogTemplateError(ctx.LoadTemplate("setting/ssh-key").Execute(w, templates.SettingSSHKeyTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				KeyList: keyList,
				ErrorMsg: struct{Type string; Message string}{
					Type: "",
					Message: "Invalid key text",
				},
			}))
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
		FoundAt(w, "/setting/ssh")
	}))
	
	http.HandleFunc("GET /setting/ssh/{keyName}/delete", WithLog(func(w http.ResponseWriter, r *http.Request){
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !CheckGlobalVisibleToUser(ctx, loginInfo) {
			switch ctx.Config.GlobalVisibility {
			case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
				FoundAt(w, "/maintenance-notice")
				return
			case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
				FoundAt(w, "/shutdown-notice")
				return
			case aegis.GLOBAL_VISIBILITY_PRIVATE:
				FoundAt(w, "/login")
				return
			}
		}
		if !loginInfo.LoggedIn { FoundAt(w, "/"); return }
		un := loginInfo.UserName
		if !model.ValidUserName(un) {
			ctx.ReportNotFound(un, "User", "Depot", w, r)
			return
		}
		keyName := r.PathValue("keyName")
		err = ctx.DatabaseInterface.RemoveAuthKey(un, keyName)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		ctx.SSHKeyManagingContext.RemoveAuthorizedKey(un, keyName)
		err = ctx.SSHKeyManagingContext.Sync()
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		FoundAt(w, "/setting/ssh")
	}))
	
	http.HandleFunc("GET /setting/ssh/{keyName}/edit", WithLog(func(w http.ResponseWriter, r *http.Request){
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !CheckGlobalVisibleToUser(ctx, loginInfo) {
			switch ctx.Config.GlobalVisibility {
			case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
				FoundAt(w, "/maintenance-notice")
				return
			case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
				FoundAt(w, "/shutdown-notice")
				return
			case aegis.GLOBAL_VISIBILITY_PRIVATE:
				FoundAt(w, "/login")
				return
			}
		}
		if !loginInfo.LoggedIn { FoundAt(w, "/"); return }
		un := loginInfo.UserName
		if !model.ValidUserName(un) {
			ctx.ReportNotFound(un, "User", "Depot", w, r)
			return
		}
		k, err := ctx.DatabaseInterface.GetAuthKeyByName(un, r.PathValue("keyName"))
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		LogTemplateError(ctx.LoadTemplate("setting/edit-ssh-key").Execute(w, &templates.SettingEditSSHKeyTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			Key: k,
		}))
	}))
	
	http.HandleFunc("POST /setting/ssh/{keyName}/edit", WithLog(func(w http.ResponseWriter, r *http.Request){
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !CheckGlobalVisibleToUser(ctx, loginInfo) {
			switch ctx.Config.GlobalVisibility {
			case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
				FoundAt(w, "/maintenance-notice")
				return
			case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
				FoundAt(w, "/shutdown-notice")
				return
			case aegis.GLOBAL_VISIBILITY_PRIVATE:
				FoundAt(w, "/login")
				return
			}
		}
		if !loginInfo.LoggedIn { FoundAt(w, "/"); return }
		un := loginInfo.UserName
		if !model.ValidUserName(un) {
			ctx.ReportNotFound(un, "User", "Depot", w, r)
			return
		}
		err = r.ParseForm()
		if err != nil {
			ctx.ReportNormalError("Invalid request", w, r)
			return
		}
		keyName := r.PathValue("keyName")
		chkres, err := checkUserPassword(ctx, un, r.Form.Get("password"))
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
		}
		if !chkres {
			ctx.ReportRedirect(fmt.Sprintf("/setting/ssh/%s/edit", keyName), 3, "Password Mismatch", "The password you've provided does not match. Please try again.", w, r)
			return
		}
		keyText := r.Form.Get("key-text")
		err = ctx.DatabaseInterface.UpdateAuthKey(un, keyName, keyText)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		ctx.ReportRedirect(fmt.Sprintf("/setting/ssh/%s/edit", keyName), 3, "Updated", "Updated.", w, r)
	}))
}


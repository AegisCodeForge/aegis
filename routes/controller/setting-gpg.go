package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/auxfuncs"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindSettingGPGController(ctx *RouterContext) {
	http.HandleFunc("GET /setting/gpg", WithLog(func(w http.ResponseWriter, r *http.Request) {
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
		s, err := ctx.DatabaseInterface.GetAllSignKeyByUsername(un)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		LogTemplateError(ctx.LoadTemplate("setting/gpg-key").Execute(w, templates.SettingGPGKeyTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			KeyList: s,
		}))
	}))
	
	http.HandleFunc("POST /setting/gpg", WithLog(func(w http.ResponseWriter, r *http.Request){
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
		err = r.ParseForm()
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		confirmPassword := strings.TrimSpace(r.Form.Get("password"))
		chkres, err := checkUserPassword(ctx, un, confirmPassword)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !chkres {
			ctx.ReportRedirect("/setting/gpg", 3, "Password Mismatch", "The password you've provided does not match. Please try again.", w, r)
			return
		}
		keyText := strings.TrimSpace(r.Form.Get("key-text"))
		if len(keyText) <= 0 {
			ctx.ReportRedirect("/setting/gpg", 3, "Invalid Key Text", "Key text cannot be empty.", w, r)
			return
		}
		s := strings.Split(keyText, " ")
		keyName := ""
		if len(s) < 3 {
			keyName = "key_" + auxfuncs.GenSym(8)
		} else {
			keyName = s[2]
		}
		err = ctx.DatabaseInterface.RegisterSignKey(un, keyName, keyText)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		FoundAt(w, "/setting/gpg")
	}))
	
	http.HandleFunc("GET /setting/gpg/{keyName}/delete", WithLog(func(w http.ResponseWriter, r *http.Request){
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
		if !loginInfo.LoggedIn {
			ctx.ReportRedirect("/", 3, "Not Logged In", "Please login before you perform this action.", w, r)
			return
		}
		un := loginInfo.UserName
		err = ctx.DatabaseInterface.RemoveSignKey(un, r.PathValue("keyName"))
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		FoundAt(w, "/setting/gpg")
	}))
	
	http.HandleFunc("GET /setting/gpg/{keyName}/edit", WithLog(func(w http.ResponseWriter, r *http.Request){
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
		if !loginInfo.LoggedIn {
			ctx.ReportRedirect("/", 3, "Not Logged In", "Please login before you perform this action.", w, r)
			return
		}
		un := loginInfo.UserName
		k, err := ctx.DatabaseInterface.GetSignKeyByName(un, r.PathValue("keyName"))
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		LogTemplateError(ctx.LoadTemplate("setting/edit-gpg-key").Execute(w, &templates.SettingEditGPGKeyTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			Key: k,
		}))
	}))
	
	http.HandleFunc("POST /setting/gpg/{keyName}/edit", WithLog(func(w http.ResponseWriter, r *http.Request){
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
		if !loginInfo.LoggedIn {
			ctx.ReportRedirect("/", 3, "Not Logged In", "Please login before you perform this action.", w, r)
			return
		}
		un := loginInfo.UserName
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
			ctx.ReportRedirect(fmt.Sprintf("/setting/gpg/%s/edit", keyName), 3, "Password Mismatch", "The password you've provided does not match. Please try again.", w, r)
			return
		}
		keyText := r.Form.Get("key-text")
		err = ctx.DatabaseInterface.UpdateSignKey(un, keyName, keyText)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		ctx.ReportRedirect(fmt.Sprintf("/setting/gpg/%s/edit", keyName), 3, "Updated", "Updated.", w, r)
	}))
}


package controller

import (
	"net/http"
	"strings"

	. "github.com/bctnry/gitus/routes"
	"github.com/bctnry/gitus/templates"
	"golang.org/x/crypto/bcrypt"
)


func bindSettingGPGController(ctx *RouterContext) {
	http.HandleFunc("GET /setting/gpg", WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !loginInfo.LoggedIn { FoundAt(w, "/"); return }
		un := loginInfo.UserName
		s, err := ctx.DatabaseInterface.GetAllAuthKeyByUsername(un)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		LogTemplateError(ctx.LoadTemplate("setting-gpg-key").Execute(w, templates.SettingGPGKeyTemplateModel{
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
		if !loginInfo.LoggedIn { FoundAt(w, "/"); return }
		un := loginInfo.UserName
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
			LogTemplateError(ctx.LoadTemplate("setting-gpg-key").Execute(w, templates.SettingGPGKeyTemplateModel{
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
		keyText := r.Form.Get("key-text")
		if len(strings.TrimSpace(keyText)) <= 0 {
			LogTemplateError(ctx.LoadTemplate("setting-gpg-key").Execute(w, templates.SettingGPGKeyTemplateModel{
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
			keyName = "key_" + mkname(8)
		} else {
			keyName = s[2]
		}
		err = ctx.DatabaseInterface.RegisterAuthKey(un, keyName, keyText)
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
		if !loginInfo.LoggedIn { FoundAt(w, "/"); return }
		un := loginInfo.UserName
		err = ctx.DatabaseInterface.RemoveAuthKey(un, r.PathValue("keyName"))
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		FoundAt(w, "/setting/gpg")
	}))
}

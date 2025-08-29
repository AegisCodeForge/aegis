package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/pkg/auxfuncs"
	"github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindSettingGPGController(ctx *routes.RouterContext) {
	http.HandleFunc("GET /setting/gpg", routes.UseMiddleware(
		[]routes.Middleware{
			routes.Logged, routes.LoginRequired,
			routes.GlobalVisibility,
			routes.ErrorGuard,
		}, ctx, 
		func(ctx *routes.RouterContext, w http.ResponseWriter, r *http.Request) {
			un := ctx.LoginInfo.UserName
			s, err := ctx.DatabaseInterface.GetAllSignKeyByUsername(un)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			routes.LogTemplateError(ctx.LoadTemplate("setting/gpg-key").Execute(w, templates.SettingGPGKeyTemplateModel{
				Config: ctx.Config,
				LoginInfo: ctx.LoginInfo,
				KeyList: s,
			}))
		},
	))
	
	http.HandleFunc("POST /setting/gpg", routes.UseMiddleware(
		[]routes.Middleware{
			routes.Logged, routes.LoginRequired,
			routes.GlobalVisibility,
			routes.ErrorGuard,
		}, ctx,
		func(ctx *routes.RouterContext, w http.ResponseWriter, r *http.Request) {
			un := ctx.LoginInfo.UserName
			if !model.ValidUserName(un) {
				ctx.ReportNotFound(un, "User", "Depot", w, r)
				return
			}
			err := r.ParseForm()
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
			routes.FoundAt(w, "/setting/gpg")
		},
	))
	
	http.HandleFunc("GET /setting/gpg/{keyName}/delete", routes.UseMiddleware(
		[]routes.Middleware{
			routes.Logged, routes.LoginRequired,
			routes.GlobalVisibility,
			routes.ErrorGuard,
		}, ctx,
		func(ctx *routes.RouterContext, w http.ResponseWriter, r *http.Request){
			un := ctx.LoginInfo.UserName
			if !model.ValidUserName(un) {
				ctx.ReportNotFound(un, "User", "Depot", w, r)
				return
			}
			err := ctx.DatabaseInterface.RemoveSignKey(un, r.PathValue("keyName"))
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			routes.FoundAt(w, "/setting/gpg")
		},
	))
	
	http.HandleFunc("GET /setting/gpg/{keyName}/edit", routes.UseMiddleware(
		[]routes.Middleware{
			routes.Logged, routes.LoginRequired,
			routes.GlobalVisibility,
			routes.ErrorGuard,
		}, ctx,
		func(ctx *routes.RouterContext, w http.ResponseWriter, r *http.Request){
			un := ctx.LoginInfo.UserName
			if !model.ValidUserName(un) {
				ctx.ReportNotFound(un, "User", "Depot", w, r)
				return
			}
			k, err := ctx.DatabaseInterface.GetSignKeyByName(un, r.PathValue("keyName"))
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			routes.LogTemplateError(ctx.LoadTemplate("setting/edit-gpg-key").Execute(w, &templates.SettingEditGPGKeyTemplateModel{
				Config: ctx.Config,
				LoginInfo: ctx.LoginInfo,
				Key: k,
			}))
		},
	))
	
	http.HandleFunc("POST /setting/gpg/{keyName}/edit", routes.UseMiddleware(
		[]routes.Middleware{
			routes.Logged, routes.LoginRequired,
			routes.GlobalVisibility,
			routes.ErrorGuard,
		}, ctx,
		func(ctx *routes.RouterContext, w http.ResponseWriter, r *http.Request){
			un := ctx.LoginInfo.UserName
			if !model.ValidUserName(un) {
				ctx.ReportNotFound(un, "User", "Depot", w, r)
				return
			}
			err := r.ParseForm()
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
		},
	))
}


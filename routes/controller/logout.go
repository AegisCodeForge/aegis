package controller

import (
	"net/http"

	. "github.com/bctnry/gitus/routes"
)


func bindLogoutController(ctx *RouterContext) {
	http.HandleFunc("GET /logout", WithLog(func(w http.ResponseWriter, r *http.Request) {
		sk, err := r.Cookie("session")
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		err = ctx.SessionInterface.RevokeSession(sk.Value)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		w.Header().Add("Set-Cookie", (&http.Cookie{
			Name: COOKIE_KEY_SESSION,
			Value: "",
			Path: "/",
			MaxAge: -1,
			HttpOnly: true,
			Secure: true,
			SameSite: http.SameSiteDefaultMode,
		}).String())
		w.Header().Add("Set-Cookie", (&http.Cookie{
			Name: COOKIE_KEY_USERNAME,
			Value: "",
			Path: "/",
			MaxAge: -1,
			HttpOnly: true,
			Secure: true,
			SameSite: http.SameSiteDefaultMode,
		}).String())
		FoundAt(w, "/")
	}))
}



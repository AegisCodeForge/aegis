package controller

import (
	"fmt"
	"net/http"

	. "github.com/bctnry/aegis/routes"
)

func bindDynamicAssetController(ctx *RouterContext) {
	http.HandleFunc("GET /dynamic-asset/style-const-default.css", UseMiddleware(
		[]Middleware{Logged, UseLoginInfo, ErrorGuard},
		ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			if !rc.LoginInfo.LoggedIn {
				FoundAt(w, "/static/style-const-default.css")
				return
			}
			u, err := rc.DatabaseInterface.GetUserByName(rc.LoginInfo.UserName)
			if err != nil {
				w.WriteHeader(500)
				fmt.Fprintf(w, "Failed to read username %s: %s", rc.LoginInfo.UserName, err)
				return
			}
			w.Header().Add("Content-Type", "text/css")
			w.WriteHeader(200)
			fmt.Fprintf(w, ":root { --foreground-color: %s; --background-color: %s; }", u.WebsitePreference.ForegroundColor, u.WebsitePreference.BackgroundColor)
		},
	))
}


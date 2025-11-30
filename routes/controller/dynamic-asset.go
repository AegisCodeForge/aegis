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
			// NOTE: in the future we should probably also save site-wide default
			// colors as files when we sync the config - the reason why i'm not doing
			// that now is because this probably shouldn't be the responsiblity
			// of config.Sync concept-wise. currently the templates loads both static
			// assets as some kind of carry-over of the fact that we don't sync
			// site-wide config to static files as of now (2025.11.30).
			// once we've decided this dynamic assets wouldn't become too big of
			// a problem in terms of speed we'll switch to the dynamic asset
			// completely and deprecate /static/style-const-default.css .
			var foreground string
			var background string
			if !rc.LoginInfo.LoggedIn {
				foreground = rc.Config.Theme.ForegroundColor
				background = rc.Config.Theme.BackgroundColor
			} else {
				u, err := rc.DatabaseInterface.GetUserByName(rc.LoginInfo.UserName)
				if err != nil {
					foreground = rc.Config.Theme.ForegroundColor
					background = rc.Config.Theme.BackgroundColor
				} else if u.WebsitePreference.UseSiteWideThemeConfig {
					foreground = rc.Config.Theme.ForegroundColor
					background = rc.Config.Theme.BackgroundColor
				} else {
					foreground = u.WebsitePreference.ForegroundColor
					background = u.WebsitePreference.BackgroundColor
				}
			}
			w.Header().Add("Content-Type", "text/css")
			w.WriteHeader(200)
			fmt.Fprintf(w, ":root { --foreground-color: %s; --background-color: %s; }", foreground, background)
		},
	))
}


package controller

import (
	"net/http"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/pkg/aegis/session"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
	"golang.org/x/crypto/bcrypt"
)


func bindLoginController(ctx *RouterContext) {
	http.HandleFunc("GET /login", WithLog(func(w http.ResponseWriter, r *http.Request) {
		if ctx.Config.GlobalVisibility == aegis.GLOBAL_VISIBILITY_MAINTENANCE {
			FoundAt(w, "/maintenance-notice")
			return
		}
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			loginInfo.LoggedIn = false
			loginInfo.UserName = ""
		}
		if loginInfo.LoggedIn { FoundAt(w, "/") }
		LogTemplateError(ctx.LoadTemplate("login").Execute(w, templates.LoginTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
 		}))
	}))

	http.HandleFunc("POST /login", UseMiddleware(
		[]Middleware{
			Logged, RateLimit,
		}, ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			if rc.Config.GlobalVisibility == aegis.GLOBAL_VISIBILITY_MAINTENANCE {
				FoundAt(w, "/maintenance-notice")
				return
			}
			err := r.ParseForm()
			if err != nil {
				rc.ReportInternalError(err.Error(), w, r)
				return
			}
			un := r.Form.Get("username")
			ph := r.Form.Get("password")
			u, err := rc.DatabaseInterface.GetUserByName(un)
			if err != nil {
				rc.ReportInternalError(err.Error(), w, r)
				return
			}
			switch u.Status {
			case model.BANNED: 
				LogTemplateError(rc.LoadTemplate("login").Execute(w, templates.LoginTemplateModel{
					Config: rc.Config,
					ErrorMsg: "User suspended.",
				}))
				return
			case model.NORMAL_USER_APPROVAL_NEEDED:
				LogTemplateError(rc.LoadTemplate("login").Execute(w, templates.LoginTemplateModel{
					Config: rc.Config,
					ErrorMsg: "User waiting for approval.",
				}))
				return
			case model.NORMAL_USER_CONFIRM_NEEDED:
				LogTemplateError(rc.LoadTemplate("login").Execute(w, templates.LoginTemplateModel{
					Config: rc.Config,
					ErrorMsg: "Confirmation needed.",
				}))
				return
			default:
				err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(ph))
				if err == bcrypt.ErrMismatchedHashAndPassword {
					LogTemplateError(rc.LoadTemplate("login").Execute(w, templates.LoginTemplateModel{
						Config: rc.Config,
						ErrorMsg: "Invalid username or password.",
					}))
					return
				} else if err != nil {
					LogTemplateError(rc.LoadTemplate("login").Execute(w, templates.LoginTemplateModel{
						Config: rc.Config,
						ErrorMsg: "Internal error: " + err.Error(),
					}))
					return
				} else {
					ss := session.NewSessionString()
					err = rc.SessionInterface.RegisterSession(un, ss)
					if err != nil {
						rc.ReportInternalError(err.Error(), w, r)
						return
					}
					
					w.Header().Add("Set-Cookie", (&http.Cookie{
						Name: COOKIE_KEY_SESSION,
						Value: ss,
						Path: "/",
						MaxAge: 3600,
						HttpOnly: true,
						Secure: true,
						SameSite: http.SameSiteDefaultMode,
					}).String())
					w.Header().Add("Set-Cookie", (&http.Cookie{
						Name: "username",
						Value: un,
						Path: "/",
						MaxAge: 3600,
						HttpOnly: true,
						Secure: true,
						SameSite: http.SameSiteDefaultMode,
					}).String())
					FoundAt(w, "/")
				}
			}
		},
	))
}



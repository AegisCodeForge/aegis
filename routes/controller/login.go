package controller

import (
	"net/http"

	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/pkg/aegis/session"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
	"golang.org/x/crypto/bcrypt"
)


func bindLoginController(ctx *RouterContext) {
	http.HandleFunc("GET /login", WithLog(func(w http.ResponseWriter, r *http.Request) {
		LogTemplateError(ctx.LoadTemplate("login").Execute(w, templates.LoginTemplateModel{
			Config: ctx.Config,
		}))
	}))

	http.HandleFunc("POST /login", WithLog(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		un := r.Form.Get("username")
		ph := r.Form.Get("password")
		u, err := ctx.DatabaseInterface.GetUserByName(un)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		switch u.Status {
		case model.DELETED:
			LogTemplateError(ctx.LoadTemplate("login").Execute(w, templates.LoginTemplateModel{
				Config: ctx.Config,
				ErrorMsg: "Invalid username or password.",
			}))
			return
		case model.BANNED:
			LogTemplateError(ctx.LoadTemplate("login").Execute(w, templates.LoginTemplateModel{
				Config: ctx.Config,
				ErrorMsg: "User suspended.",
			}))
			return
		case model.NORMAL_USER_APPROVAL_NEEDED:
			LogTemplateError(ctx.LoadTemplate("login").Execute(w, templates.LoginTemplateModel{
				Config: ctx.Config,
				ErrorMsg: "User waiting for approval.",
			}))
			return
		case model.NORMAL_USER_CONFIRM_NEEDED:
			LogTemplateError(ctx.LoadTemplate("login").Execute(w, templates.LoginTemplateModel{
				Config: ctx.Config,
				ErrorMsg: "Confirmation needed.",
			}))
			return
		default:
			err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(ph))
			if err == bcrypt.ErrMismatchedHashAndPassword {
				LogTemplateError(ctx.LoadTemplate("login").Execute(w, templates.LoginTemplateModel{
					Config: ctx.Config,
					ErrorMsg: "Invalid username or password.",
				}))
				return
			} else if err != nil {
				LogTemplateError(ctx.LoadTemplate("login").Execute(w, templates.LoginTemplateModel{
					Config: ctx.Config,
					ErrorMsg: "Internal error: " + err.Error(),
				}))
				return
			} else {
				ss := session.NewSessionString()
				err = ctx.SessionInterface.RegisterSession(un, ss)
				if err != nil {
					ctx.ReportInternalError(err.Error(), w, r)
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
	}))
}



package routes

import (
	"fmt"
	"log"
	"net/http"
	"slices"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/templates"
)

// middleware...

type Middleware func(HandlerFunc)HandlerFunc;
type HandlerFunc func(*RouterContext, http.ResponseWriter, *http.Request);

func UseMiddleware(w []Middleware, ctx *RouterContext, f HandlerFunc) http.HandlerFunc {
	if len(w) <= 0 {
		return func(w http.ResponseWriter, r *http.Request) {
			f(ctx, w, r);
		}
	}
	var res HandlerFunc = w[len(w)-1](f)
	i := len(w)-2
	for i >= 0 { res = w[i](res); i -= 1; }
	return func(w http.ResponseWriter, r *http.Request) {
		rc := ctx.NewLocal()
		res(rc, w, r);
	}
}

var Logged Middleware = func(f HandlerFunc) HandlerFunc {
	return func(ctx *RouterContext, w http.ResponseWriter, r *http.Request) {
		log.Printf(" %s %s %s\n", r.RemoteAddr, r.Method, r.URL.Path)
		f(ctx, w, r)
	}
}

var ValidPOSTRequestRequired Middleware = func(f HandlerFunc) HandlerFunc {
	return func(ctx *RouterContext, w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			ctx.ReportNormalError("Invalid request", w, r)
			return
		}
		f(ctx, w, r)
	}
}

var JSONRequestRequired Middleware = func(f HandlerFunc) HandlerFunc {
	return func(ctx *RouterContext, w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			ctx.ReportNormalError("Invalid request", w, r)
			return
		}
		f(ctx, w, r)
	}
}

var UseLoginInfo Middleware = func(f HandlerFunc) HandlerFunc {
	return func(ctx *RouterContext, w http.ResponseWriter, r *http.Request) {
		if !ctx.Config.PlainMode {
			ctx.LoginInfo, ctx.LastError = GenerateLoginInfoModel(ctx, r)
		}
		f(ctx, w, r)
	}
}

var LoginRequired Middleware = func(f HandlerFunc) HandlerFunc {
	return func(ctx *RouterContext, w http.ResponseWriter, r *http.Request) {
		if !ctx.Config.PlainMode {
			ctx.LoginInfo, ctx.LastError = GenerateLoginInfoModel(ctx, r)
			if ctx.LastError != nil {
				ctx.ReportRedirect("/login", 0, "Login Check Failed", fmt.Sprintf("Failed while checking login status: %s.", ctx.LastError), w, r)
				return
			}
			if !ctx.LoginInfo.LoggedIn {
				ctx.ReportRedirect("/login", 0, "Login Required", "The action you requested requires you to log in. Please log in and try again.", w, r)
				return
			}
		}
		f(ctx, w, r)
	}
}

var AdminRequired Middleware = func(f HandlerFunc) HandlerFunc {
	return func(ctx *RouterContext, w http.ResponseWriter, r *http.Request) {
		if !ctx.Config.PlainMode {
			if ctx.LoginInfo == nil {
				ctx.LoginInfo, ctx.LastError = GenerateLoginInfoModel(ctx, r)
				if ctx.LastError != nil {
					ctx.ReportRedirect("/login", 0, "Login Check Failed", fmt.Sprintf("Failed while checking login status: %s.", ctx.LastError), w, r)
					return
				}
			}
			if !ctx.LoginInfo.IsAdmin {
				ctx.ReportRedirect("/", 0, "Permission Denied", "You need administrator prividege to perform this action.", w, r)
				return
			} else {
				f(ctx, w, r)
			}
		} else {
			ctx.ReportNotFound(r.URL.Path, "Route", "here", w, r)
		}
	}
}

var ErrorGuard Middleware = func(f HandlerFunc) HandlerFunc {
	return func(ctx *RouterContext, w http.ResponseWriter, r *http.Request) {
		if ctx.LastError != nil {
			ctx.ReportInternalError(fmt.Sprintf("Internal error: %s\n", ctx.LastError), w, r)
			return
		}
		f(ctx, w, r)
	}
}

func ValidRepositoryNameRequired(s string) Middleware {
	return func(f HandlerFunc) HandlerFunc {
		return func(ctx *RouterContext, w http.ResponseWriter, r *http.Request) {
			repoName := r.PathValue(s)
			if !model.ValidRepositoryName(repoName) {
				ctx.ReportNotFound(repoName, "Repository", "Namespace", w, r)
				return
			}
			f(ctx, w, r)
		}
	}
}

func CheckGlobalVisibleToUser(ctx *RouterContext, loginInfo *templates.LoginInfoModel) bool {
	if ctx.Config.PlainMode { return true }
	if loginInfo == nil { return false }
	switch ctx.Config.GlobalVisibility {
	case aegis.GLOBAL_VISIBILITY_PUBLIC: return true
	case aegis.GLOBAL_VISIBILITY_PRIVATE: return loginInfo.LoggedIn
	case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
		return slices.Contains(ctx.Config.FullAccessUser, loginInfo.UserName)
	case aegis.GLOBAL_VISIBILITY_MAINTENANCE: return false
	default: return false
	}
}

var GlobalVisibility Middleware = func(f HandlerFunc) HandlerFunc {
	return func(ctx *RouterContext, w http.ResponseWriter, r *http.Request) {
		if !CheckGlobalVisibleToUser(ctx, ctx.LoginInfo) {
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
		f(ctx, w, r)
	}
}

var RateLimit Middleware = func(f HandlerFunc) HandlerFunc {
	return func(ctx *RouterContext, w http.ResponseWriter, r *http.Request) {
		if ctx.RateLimiter.IsIPAllowed(ResolveMostPossibleIP(w, r)) {
			f(ctx, w, r)
		} else {
			w.WriteHeader(429)
		}
	}
}



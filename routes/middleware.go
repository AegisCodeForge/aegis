package routes

import (
	"fmt"
	"log"
	"net/http"
	"slices"

	"github.com/bctnry/aegis/pkg/aegis"
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
		res(ctx, w, r);
	}
}

var Logged Middleware = func(f HandlerFunc) HandlerFunc {
	return func(ctx *RouterContext, w http.ResponseWriter, r *http.Request) {
		log.Printf(" %s %s\n", r.Method, r.URL.Path)
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
var ErrorGuard Middleware = func(f HandlerFunc) HandlerFunc {
	return func(ctx *RouterContext, w http.ResponseWriter, r *http.Request) {
		if ctx.LastError != nil {
			ctx.ReportInternalError(fmt.Sprintf("Internal error: %s\n", ctx.LastError), w, r)
			return
		}
		f(ctx, w, r)
	}
}

func CheckGlobalVisibleToUser(ctx *RouterContext, loginInfo *templates.LoginInfoModel) bool {
	if loginInfo == nil { return false }
	if ctx.Config.PlainMode { return true }
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



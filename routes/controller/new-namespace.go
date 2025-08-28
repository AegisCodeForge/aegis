package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/db"
	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/routes"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindNewNamespaceController(ctx *RouterContext) {
	http.HandleFunc("GET /new/namespace", WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !CheckGlobalVisibleToUser(ctx, loginInfo) {
			switch ctx.Config.GlobalVisibility {
			case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
				routes.FoundAt(w, "/maintenance-notice")
				return
			case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
				routes.FoundAt(w, "/shutdown-notice")
				return
			case aegis.GLOBAL_VISIBILITY_PRIVATE:
				routes.FoundAt(w, "/login")
				return
			}
		}
		if !loginInfo.LoggedIn { FoundAt(w, "/"); return }

		LogTemplateError(ctx.LoadTemplate("new/namespace").Execute(w, templates.NewNamespaceTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
		}))
	}))
	http.HandleFunc("POST /new/namespace", WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
 			return
		}
		if !CheckGlobalVisibleToUser(ctx, loginInfo) {
			switch ctx.Config.GlobalVisibility {
			case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
				routes.FoundAt(w, "/maintenance-notice")
				return
			case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
				routes.FoundAt(w, "/shutdown-notice")
				return
			case aegis.GLOBAL_VISIBILITY_PRIVATE:
				routes.FoundAt(w, "/login")
				return
			}
		}
		if !loginInfo.LoggedIn { FoundAt(w, "/"); return }
		err = r.ParseForm()
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		fmt.Println(loginInfo)
		userName := loginInfo.UserName
		newNamespaceName := r.Form.Get("name")
		if !model.ValidNamespaceName(newNamespaceName) {
			ctx.ReportRedirect("/new/namespace", 5, "Invalid Namespace Name", "Namespace name must consists of only upper & lowercase letters (a-z, A-Z), 0-9, underscore and hyphen.", w, r)
			return
		}
		ns, err := ctx.DatabaseInterface.RegisterNamespace(newNamespaceName, userName)
		if err != nil {
			if err == db.ErrEntityAlreadyExists {
				ctx.ReportRedirect("/new/namespace", 5, "Already Exists", fmt.Sprintf("Namespace \"%s\" already exists; please choose another name.", newNamespaceName) , w, r)
			} else {
				ctx.ReportInternalError(err.Error(), w, r)
			}
			return
		}
		newNamespaceTitle := r.Form.Get("title")
		if len(strings.TrimSpace(newNamespaceTitle)) > 0 {
			ns.Title = strings.TrimSpace(newNamespaceTitle)
			// NOTE: we don't care if the title setting failed; we have the
			// namespace already.
			ctx.DatabaseInterface.UpdateNamespaceInfo(newNamespaceName, ns)
		}
		FoundAt(w, fmt.Sprintf("/s/%s", newNamespaceName))
	}))
}


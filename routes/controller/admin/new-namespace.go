package admin

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/bctnry/aegis/pkg/aegis/model"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindAdminNewNamespaceController(ctx *RouterContext) {
	http.HandleFunc("GET /admin/new-namespace", UseMiddleware(
		[]Middleware{Logged, LoginRequired, AdminRequired,
			GlobalVisibility, ErrorGuard,
		}, ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			LogTemplateError(rc.LoadTemplate("admin/new-namespace").Execute(w, &templates.AdminConfigTemplateModel{
				Config: rc.Config,
				LoginInfo: rc.LoginInfo,
				ErrorMsg: "",
			}))
		},
	))
	
	http.HandleFunc("POST /admin/new-namespace", UseMiddleware(
		[]Middleware{Logged, LoginRequired, AdminRequired,
			GlobalVisibility, ErrorGuard,
		}, ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			if err != nil {
				rc.ReportNormalError("Invalid request", w, r)
				return
			}
			owner := r.Form.Get("owner")
			name := r.Form.Get("name")
			if !model.ValidNamespaceName(name) {
				rc.ReportRedirect("/admin/new-namespace", 5, "Invalid Namespace Name", "Invalid namespace name; namespace name must only contains uppercase & lowercase letters (a-z, A-Z), 0-9, underscore and hyphen.\n", w, r)
				return
			}
			title := r.Form.Get("title")
			email := r.Form.Get("email")
			description := r.Form.Get("description")
			i, err := strconv.ParseInt(r.Form.Get("status"), 10, 32)
			if err != nil {
				rc.ReportNormalError("Invalid request", w, r)
				return
			}
			ns, err := rc.DatabaseInterface.RegisterNamespace(name, owner)
			if err != nil {
				rc.ReportInternalError(fmt.Sprintf("Failed to register namespace: %s", err), w, r)
				return
			}
			ns.Status = model.AegisNamespaceStatus(i)
			ns.Title = title
			ns.Email = email
			ns.Owner = owner
			ns.Description = description
			err = rc.DatabaseInterface.UpdateNamespaceInfo(name, ns)
			if err != nil {
				rc.ReportInternalError(fmt.Sprintf("Failed to update namespace info: %s", err), w, r)
				return
			}
			FoundAt(w, "/admin/namespace-list")
		},
	))
}


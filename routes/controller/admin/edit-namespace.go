package admin

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)


func bindAdminEditNamespaceController(ctx *routes.RouterContext) {
	http.HandleFunc("GET /admin/namespace/{name}/edit", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		nsn := r.PathValue("name")
		if !model.ValidNamespaceName(nsn) { routes.FoundAt(w, "/") }
		ns, err := ctx.DatabaseInterface.GetNamespaceByName(nsn)
		if err != nil {
			ctx.ReportRedirect("/admin/namespace-list", 0, "Error",
				fmt.Sprintf("Failed to fetch namespace: %s", err.Error()),
				w, r,
			)
			return
		}
		routes.LogTemplateError(ctx.LoadTemplate("admin/namespace-edit").Execute(w, &templates.AdminNamespaceEditTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			Namespace: ns,
		}))
		return
	}))
	
 	http.HandleFunc("POST /admin/namespace/{name}/edit", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		nsn := r.PathValue("name")
		if !model.ValidNamespaceName(nsn) { routes.FoundAt(w, "/") }
		ns, err := ctx.DatabaseInterface.GetNamespaceByName(nsn)
		if err != nil {
			routes.LogTemplateError(ctx.LoadTemplate("admin/_redirect-with-message").Execute(w, &templates.AdminRedirectWithMessageModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: "",
				Timeout: 0,
				RedirectUrl: "/admin/namespace-list",
				MessageTitle: "Error",
				MessageText: fmt.Sprintf("Failed to fetch namespace: %s", err.Error()),
			}))
			return
		}
		err = r.ParseForm()
		if err != nil {
			ctx.ReportRedirect("/admin/namespace-list", 3, "Error", "Invalid request", w, r)
			return
		}
		title := r.Form.Get("title")
		owner := r.Form.Get("owner")
		email := r.Form.Get("email")
		i, err := strconv.Atoi(r.Form.Get("status"))
		if err != nil {
			ctx.ReportRedirect("/admin/namespace-list", 3, "Error", "Invalid status value. Please try again.", w, r)
			return
		}
		description := r.Form.Get("description")
		ns.Title = title
		ns.Email = email
		ns.Owner = owner
		ns.Status = model.AegisNamespaceStatus(i)
		ns.Description = description
		err = ctx.DatabaseInterface.UpdateNamespaceInfo(nsn, ns)
		if err != nil {
			routes.LogTemplateError(ctx.LoadTemplate("admin/namespace-edit").Execute(w, templates.AdminNamespaceEditTemplateModel{
				Namespace: ns,
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: struct{Type string; Message string}{
					Type: r.Form.Get("type"),
					Message: fmt.Sprintf("Failed to update namespace: %s.", err.Error()),
				},
			}))
		}
		routes.LogTemplateError(ctx.LoadTemplate("admin/namespace-edit").Execute(w, templates.AdminNamespaceEditTemplateModel{
			Namespace: ns,
			Config: ctx.Config,
			LoginInfo: loginInfo,
			ErrorMsg: struct{Type string; Message string}{
				Type: r.Form.Get("type"),
				Message: "Updated.",
			},
		}))
	}))
}


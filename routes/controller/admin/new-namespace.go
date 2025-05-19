package admin

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindAdminNewNamespaceController(ctx *routes.RouterContext) {
	http.HandleFunc("GET /admin/new-namespace", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		routes.LogTemplateError(ctx.LoadTemplate("admin/new-namespace").Execute(w, &templates.AdminConfigTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			ErrorMsg: "",
		}))
	}))
	http.HandleFunc("POST /admin/new-namespace", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		err = r.ParseForm()
		if err != nil {
			routes.LogTemplateError(ctx.LoadTemplate("admin/new-namespace").Execute(w, templates.AdminConfigTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: fmt.Sprintf("Failed to parse request: %s\n", err.Error()),
			}))
			return
		}
		owner := r.Form.Get("owner")
		name := r.Form.Get("name")
		title := r.Form.Get("title")
		email := r.Form.Get("email")
		description := r.Form.Get("description")
		i, err := strconv.ParseInt(r.Form.Get("status"), 10, 32)
		if err != nil {
			routes.LogTemplateError(ctx.LoadTemplate("admin/new-namespace").Execute(w, templates.AdminConfigTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: fmt.Sprintf("Invalid user status: %s.\n", err.Error()),
			}))
			return
		}
		ns, err := ctx.DatabaseInterface.RegisterNamespace(name, owner)
		if err != nil {
			routes.LogTemplateError(ctx.LoadTemplate("admin/new-namespace").Execute(w, templates.AdminConfigTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: fmt.Sprintf("Failed to create new namespace: %s.\n", err.Error()),
			}))
			return
		}
		ns.Status = model.AegisNamespaceStatus(i)
		ns.Title = title
		ns.Email = email
		ns.Owner = owner
		ns.Description = description
		err = ctx.DatabaseInterface.UpdateNamespaceInfo(name, ns)
		if err != nil {
			routes.LogTemplateError(ctx.LoadTemplate("admin/new-namespace").Execute(w, templates.AdminConfigTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: fmt.Sprintf("Failed to update namespace info: %s. Namespace is created but the info did not update.\n", err.Error()),
			}))
			return
		}
		routes.FoundAt(w, "/admin/namespace-list")
	}))
}

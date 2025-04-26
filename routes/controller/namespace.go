package controller

import (
	"net/http"

	. "github.com/bctnry/gitus/routes"
	"github.com/bctnry/gitus/templates"
)

func bindNamespaceController(ctx RouterContext) {
	if !ctx.Config.UseNamespace { return }
	http.HandleFunc("GET /s/{namespace}", WithLog(func(w http.ResponseWriter, r *http.Request) {
		namespaceName := r.PathValue("namespace")
		ns, ok := ctx.GitNamespaceList[namespaceName]
		if !ok {
			ctx.ReportNotFound(namespaceName, "Namespace", ctx.Config.DepotName, w, r)
			return
		}
		LogTemplateError(ctx.LoadTemplate("namespace").Execute(w, templates.NamespaceTemplateModel{
			DepotName: ctx.Config.DepotName,
			Namespace: ns,
		}))
	}))
}

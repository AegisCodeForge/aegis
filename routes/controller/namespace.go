package controller

import (
	"net/http"

	"github.com/bctnry/aegis/pkg/aegis/model"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindNamespaceController(ctx *RouterContext) {
	if !ctx.Config.UseNamespace { return }
	http.HandleFunc("GET /s/{namespace}", WithLog(func(w http.ResponseWriter, r *http.Request) {
		namespaceName := r.PathValue("namespace")
		var ns *model.Namespace
		var ok bool
		var err error
		if ctx.Config.PlainMode {
			ns, ok = ctx.GitNamespaceList[namespaceName]
			if !ok {
				err = ctx.SyncAllNamespacePlain()
				if err != nil {
					ctx.ReportInternalError(err.Error(), w, r)
					return
				}
				ns, ok = ctx.GitNamespaceList[namespaceName]
				if !ok {
					ctx.ReportNotFound(namespaceName, "Namespace", ctx.Config.DepotName, w, r)
					return
				}
			}
			LogTemplateError(ctx.LoadTemplate("namespace").Execute(w, templates.NamespaceTemplateModel{
				Namespace: ns,
				Config: ctx.Config,
			}))
		} else {
			ns, err = ctx.DatabaseInterface.GetNamespaceByName(namespaceName)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			s, err := ctx.DatabaseInterface.GetAllRepositoryFromNamespace(ns.Name)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			ns.RepositoryList = s
			loginInfo, err := GenerateLoginInfoModel(ctx, r)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			loginInfo.IsOwner = ns.Owner == loginInfo.UserName
			v := ns.ACL.GetUserPrivilege(loginInfo.UserName).HasSettingPrivilege()
			loginInfo.IsSettingMember = v
			LogTemplateError(ctx.LoadTemplate("namespace").Execute(w, templates.NamespaceTemplateModel{
				Namespace: ns,
				LoginInfo: loginInfo,
				Config: ctx.Config,
			}))
		}
	}))
}


package controller

import (
	"net/http"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/model"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindNamespaceController(ctx *RouterContext) {
	if !ctx.Config.UseNamespace { return }
	http.HandleFunc("GET /s/{namespace}", WithLog(func(w http.ResponseWriter, r *http.Request) {
		namespaceName := r.PathValue("namespace")
		if !model.ValidNamespaceName(namespaceName) {
			ctx.ReportNotFound(namespaceName, "Repository", "Depot", w, r)
			return
		}
		var ns *model.Namespace
		var ok bool
		var err error
		var loginInfo *templates.LoginInfoModel
		if !ctx.Config.PlainMode {
			loginInfo, err = GenerateLoginInfoModel(ctx, r)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
		}
		if ctx.Config.PlainMode || !CheckGlobalVisibleToUser(ctx, loginInfo) {
			switch ctx.Config.GlobalVisibility {
			case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
				FoundAt(w, "/maintenance-notice")
				return
			case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
				FoundAt(w, "/shutdown-notice")
				return
			case aegis.GLOBAL_VISIBILITY_PRIVATE:
				if !ctx.Config.PlainMode {
					FoundAt(w, "/login")
				} else {
					FoundAt(w, "/private-notice")
				}
				return
			}
		}
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
			loginInfo.IsOwner = ns.Owner == loginInfo.UserName
			v := ns.ACL.GetUserPrivilege(loginInfo.UserName).HasSettingPrivilege()
			loginInfo.IsSettingMember = v
			s, err := ctx.DatabaseInterface.GetAllVisibleRepositoryFromNamespace(loginInfo.UserName, ns.Name)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			ns.RepositoryList = make(map[string]*model.Repository, 0)
			for _, k := range s {
				ns.RepositoryList[k.Name] = k
			}
			LogTemplateError(ctx.LoadTemplate("namespace").Execute(w, templates.NamespaceTemplateModel{
				Namespace: ns,
				LoginInfo: loginInfo,
				Config: ctx.Config,
			}))
		}
	}))
}


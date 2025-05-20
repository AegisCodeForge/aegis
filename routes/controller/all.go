package controller

import (
	"fmt"
	"net/http"

	"github.com/bctnry/aegis/pkg/aegis/model"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

// list of namespace if enabled,
// list of repo if namespace is not used,
func bindAllController(ctx *RouterContext) {
	http.HandleFunc("GET /all", WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if ctx.Config.UseNamespace {
			var nsl map[string]*model.Namespace
			if ctx.Config.PlainMode {
				nsl, err = ctx.Config.GetAllNamespacePlain()
				if err != nil {
					ctx.ReportInternalError(err.Error(), w, r)
					return
				}
			} else {
				nsl, err = ctx.DatabaseInterface.GetAllVisibleNamespace(loginInfo.UserName)
				if err != nil {
					ctx.ReportInternalError(err.Error(), w, r)
					return
				}
			}
			for k, v := range nsl {
				fmt.Println(k, v)
			}
			LogTemplateError(ctx.LoadTemplate("all/namespace-list").Execute(w, templates.DepotNamespaceModel{
				DepotName: ctx.Config.DepotName,
				NamespaceList: nsl,
				Config: ctx.Config,
				LoginInfo: loginInfo,
			}))
		} else {
			var repol map[string]*model.Repository
			if ctx.Config.PlainMode {
				repol, err = ctx.Config.GetAllRepositoryPlain()
				if err != nil {
					ctx.ReportInternalError(err.Error(), w, r)
					return
				}
			} else {
				repol, err = ctx.DatabaseInterface.GetAllVisibleRepositoryFromNamespace(loginInfo.UserName, "")
				if err != nil {
					ctx.ReportInternalError(err.Error(), w, r)
					return
				}
			}
			LogTemplateError(ctx.LoadTemplate("all/repository-list").Execute(w, templates.AllRepositoryListModel{
				RepositoryList: repol,
				DepotName: ctx.Config.DepotName,
				Config: ctx.Config,
				LoginInfo: loginInfo,
			}))
		}
	}))
}


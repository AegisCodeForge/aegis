package controller

import (
	"net/http"

	"github.com/bctnry/aegis/pkg/aegis/model"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindIndexController(ctx *RouterContext) {
	http.HandleFunc("GET /", WithLog(func(w http.ResponseWriter, r *http.Request) {
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
				nsl, err = ctx.DatabaseInterface.GetAllNamespace()
				if err != nil {
					ctx.ReportInternalError(err.Error(), w, r)
					return
				}
			}
			LogTemplateError(ctx.LoadTemplate("depot-namespace").Execute(w, templates.DepotNamespaceModel{
				DepotName: ctx.Config.DepotName,
				NamespaceList: nsl,
				Config: ctx.Config,
				LoginInfo: loginInfo,
			}))
		} else {
			var repol map[string]*model.Repository
			if ctx.Config.PlainMode {
				repols, err := ctx.Config.GetAllRepositoryPlain()
				repol = make(map[string]*model.Repository, 0)
				for _, item := range repols {
					repol[item.Name] = item
				}
				if err != nil {
					ctx.ReportInternalError(err.Error(), w, r)
					return
				}
			} else {
				repol, err = ctx.DatabaseInterface.GetAllRepositoryFromNamespace("")
				if err != nil {
					ctx.ReportInternalError(err.Error(), w, r)
					return
				}
			}
			grmodel := make([]struct{RelPath string; Description string}, 0)
			for _, item := range repol {
				grmodel = append(grmodel, struct{
					RelPath string
					Description string
				}{
					RelPath: item.FullName(),
					Description: item.Description,
				})
			}
			LogTemplateError(ctx.LoadTemplate("depot-no-namespace").Execute(w, templates.DepotNoNamespaceModel{
				RepositoryList: grmodel,
				DepotName: ctx.Config.DepotName,
				Config: ctx.Config,
				LoginInfo: loginInfo,
			}))
		}
	}))
}


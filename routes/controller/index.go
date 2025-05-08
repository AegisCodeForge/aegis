package controller

import (
	"net/http"

	"github.com/bctnry/gitus/pkg/gitus/model"
	. "github.com/bctnry/gitus/routes"
	"github.com/bctnry/gitus/templates"
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
				repol, err = ctx.Config.GetAllRepositoryPlain()
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
			for key, item := range repol {
				grmodel = append(grmodel, struct{
					RelPath string
					Description string
				}{
					RelPath: key,
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


package controller

import (
	"fmt"
	"net/http"

	. "github.com/bctnry/gitus/routes"
	"github.com/bctnry/gitus/templates"
)

func bindIndexController(ctx RouterContext) {
	http.HandleFunc("GET /", WithLog(func(w http.ResponseWriter, r *http.Request) {
		if ctx.Config.UseNamespace {
			LogTemplateError(ctx.LoadTemplate("depot-namespace").Execute(w, templates.DepotNamespaceModel{
				DepotName: ctx.Config.DepotName,
				NamespaceList: ctx.GitNamespaceList,
			}))
		} else {
			fmt.Println(ctx.GitRepositoryList)
			grmodel := make([]struct{RelPath string; Description string}, 0)
			for key, item := range ctx.GitRepositoryList {
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
			}))
		}
	}))
}


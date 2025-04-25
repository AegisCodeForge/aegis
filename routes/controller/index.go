package controller

import (
	"net/http"
	
	. "github.com/bctnry/gitus/routes"
	"github.com/bctnry/gitus/templates"
)

func bindIndexController(ctx RouterContext) {
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
	http.HandleFunc("GET /", WithLog(func(w http.ResponseWriter, r *http.Request) {
		LogTemplateError(ctx.LoadTemplate("index").Execute(w, templates.IndexModel{
			RepositoryList: grmodel,
			DepotName: ctx.Config.DepotName,
		}))
	}))
}


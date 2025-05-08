package controller

import (
	"fmt"
	"net/http"

	"github.com/bctnry/gitus/routes"
	. "github.com/bctnry/gitus/routes"
	"github.com/bctnry/gitus/templates"
)

func bindRepositoryController(ctx *RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/", WithLog(func(w http.ResponseWriter, r *http.Request) {
		rfn := r.PathValue("repoName")
		namespaceName, _, s, err := ctx.ResolveRepositoryFullName(rfn)
		fmt.Println("nsn", namespaceName)
		if err != nil {
			errCode := 500
			if routes.IsRouteError(err) {
				if err.(*RouteError).ErrorType == NOT_FOUND {
					errCode = 404
				}
			}
			LogTemplateError(ctx.LoadTemplate("error").Execute(w, templates.ErrorTemplateModel{
				ErrorCode: errCode,
				ErrorMessage: err.Error(),
			}))
			return
		}
		
		err = s.Repository.SyncAllBranchList()
		if err != nil {
			LogTemplateError(ctx.LoadTemplate("error").Execute(w, templates.ErrorTemplateModel{
				ErrorCode: 500,
				ErrorMessage: fmt.Sprintf("Failed to sync branch list: %s", err.Error()),
			}))
			return
		}
		err = s.Repository.SyncAllTagList()
		if err != nil {
			LogTemplateError(ctx.LoadTemplate("error").Execute(w, templates.ErrorTemplateModel{
				ErrorCode: 500,
				ErrorMessage: fmt.Sprintf("Failed to sync tag list: %s", err.Error()),
			}))
			return
		}

		LogTemplateError(ctx.LoadTemplate("repository").Execute(w, templates.RepositoryModel{
			RepoName: rfn,
			RepoObj: s.Repository,
			RepoHeaderInfo: templates.RepoHeaderTemplateModel{
				NamespaceName: namespaceName,
				RepoName: rfn,
				TypeStr: "",
				NodeName: "",
				RepoDescription: s.Description,
				RepoLabelList: nil,
				RepoURL: fmt.Sprintf("%s/repo/%s", ctx.Config.HostName, rfn),
			},
			BranchList: s.Repository.BranchIndex,
			TagList: s.Repository.TagIndex,
		}))
	}))
}


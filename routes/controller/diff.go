package controller

import (
	"fmt"
	"net/http"

	"github.com/bctnry/aegis/pkg/gitlib"
	"github.com/bctnry/aegis/routes"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindDiffController(ctx *RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/diff/{commitId}/", WithLog(func(w http.ResponseWriter, r *http.Request){
		rfn := r.PathValue("repoName")
		_, _, repo, err := ctx.ResolveRepositoryFullName(rfn)
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
		commitId := r.PathValue("commitId")
		cobj, err := repo.Repository.ReadObject(commitId)
		if err != nil {
			LogTemplateError(ctx.LoadTemplate("error").Execute(w, templates.ErrorTemplateModel{
				ErrorCode: 500,
				ErrorMessage: fmt.Sprintf("Failed to read commit %s: %s", commitId, err),
			}))
			return
		}
		diff, err := repo.Repository.GetDiff(commitId)
		if err != nil {
			LogTemplateError(ctx.LoadTemplate("error").Execute(w, templates.ErrorTemplateModel{
				ErrorCode: 500,
				ErrorMessage: fmt.Sprintf("Failed to read diff %s: %s", commitId, err),
			}))
			return
		}
		
		var loginInfo *templates.LoginInfoModel = nil
		if !ctx.Config.PlainMode {
			loginInfo, err = GenerateLoginInfoModel(ctx, r)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
		}
		
		LogTemplateError(ctx.LoadTemplate("diff").Execute(w, templates.DiffTemplateModel{
			RepoHeaderInfo: *GenerateRepoHeader(ctx, repo, "commit", commitId),
			CommitInfo: templates.CommitInfoTemplateModel{
				RootPath: fmt.Sprintf("/repo/%s", rfn),
				Commit: cobj.(*gitlib.CommitObject),
			},
			Diff: diff,
			LoginInfo: loginInfo,
			Config: ctx.Config,
		}))
	}))
}


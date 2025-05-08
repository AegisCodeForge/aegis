package controller

import (
	"fmt"
	"net/http"

	"github.com/bctnry/gitus/pkg/gitlib"
	"github.com/bctnry/gitus/routes"
	. "github.com/bctnry/gitus/routes"
	"github.com/bctnry/gitus/templates"
)

func bindDiffController(ctx *RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/diff/{commitId}/", WithLog(func(w http.ResponseWriter, r *http.Request){
		rfn := r.PathValue("repoName")
		namespaceName, _, repo, err := ctx.ResolveRepositoryFullName(rfn)
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
				ErrorMessage: fmt.Sprintf("Failed to read commit %s: $s", commitId, err),
			}))
			return
		}
		diff, err := repo.Repository.GetDiff(commitId)
		if err != nil {
			LogTemplateError(ctx.LoadTemplate("error").Execute(w, templates.ErrorTemplateModel{
				ErrorCode: 500,
				ErrorMessage: fmt.Sprintf("Failed to read diff %s: $s", commitId, err),
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
			RepoHeaderInfo: templates.RepoHeaderTemplateModel{
				NamespaceName: namespaceName,
				RepoName: rfn,
				RepoDescription: repo.Description,
				TypeStr: "commit",
				NodeName: commitId,
				RepoLabelList: nil,
				RepoURL: fmt.Sprintf("%s/repo/%s", ctx.Config.HostName, rfn),
			},
			CommitInfo: templates.CommitInfoTemplateModel{
				RepoName: rfn,
				Commit: cobj.(*gitlib.CommitObject),
			},
			Diff: diff,
			LoginInfo: loginInfo,
		}))
	}))
}


package controller

import (
	"fmt"
	"net/http"
	
	. "github.com/bctnry/gitus/routes"
	"github.com/bctnry/gitus/templates"
	"github.com/bctnry/gitus/pkg/gitlib"
)

func bindDiffController(ctx RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/diff/{commitId}/", WithLog(func(w http.ResponseWriter, r *http.Request){
		rn := r.PathValue("repoName")
		repo, ok := ctx.GitRepositoryList[rn]
		if !ok {
			LogTemplateError(ctx.LoadTemplate("error").Execute(w, templates.ErrorTemplateModel{
				ErrorCode: 404,
				ErrorMessage: fmt.Sprintf("Repository %s not found.", rn),
			}))
			return
		}
		commitId := r.PathValue("commitId")
		cobj, err := repo.ReadObject(commitId)
		if err != nil {
			LogTemplateError(ctx.LoadTemplate("error").Execute(w, templates.ErrorTemplateModel{
				ErrorCode: 500,
				ErrorMessage: fmt.Sprintf("Failed to read commit %s: $s", commitId, err),
			}))
			return
		}
		diff, err := repo.GetDiff(commitId)
		if err != nil {
			LogTemplateError(ctx.LoadTemplate("error").Execute(w, templates.ErrorTemplateModel{
				ErrorCode: 500,
				ErrorMessage: fmt.Sprintf("Failed to read diff %s: $s", commitId, err),
			}))
			return
		}
		LogTemplateError(ctx.LoadTemplate("diff").Execute(w, templates.DiffTemplateModel{
			RepoHeaderInfo: templates.RepoHeaderTemplateModel{
				RepoName: rn,
				RepoDescription: repo.Description,
				TypeStr: "commit",
				NodeName: commitId,
				RepoLabelList: nil,
			},
			CommitInfo: templates.CommitInfoTemplateModel{
				RepoName: rn,
				Commit: cobj.(*gitlib.CommitObject),
			},
			Diff: diff,
		}))
	}))
}


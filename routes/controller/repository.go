package controller

import (
	"fmt"
	"net/http"

	. "github.com/bctnry/gitus/routes"
	"github.com/bctnry/gitus/templates"
)

func bindRepositoryController(ctx RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/", WithLog(func(w http.ResponseWriter, r *http.Request) {
		rn := r.PathValue("repoName")
		s, ok := ctx.GitRepositoryList[rn]
		if !ok {
			LogTemplateError(ctx.LoadTemplate("error").Execute(w, templates.ErrorTemplateModel{
				ErrorCode: 404,
				ErrorMessage: fmt.Sprintf("Repository %s not found.", rn),
			}))
			return
		}

		err := s.SyncAllBranchList()
		if err != nil {
			LogTemplateError(ctx.LoadTemplate("error").Execute(w, templates.ErrorTemplateModel{
				ErrorCode: 500,
				ErrorMessage: fmt.Sprintf("Failed to sync branch list: %s", err.Error()),
			}))
			return
		}
		err = s.SyncAllTagList()
		if err != nil {
			LogTemplateError(ctx.LoadTemplate("error").Execute(w, templates.ErrorTemplateModel{
				ErrorCode: 500,
				ErrorMessage: fmt.Sprintf("Failed to sync tag list: %s", err.Error()),
			}))
			return
		}

		LogTemplateError(ctx.LoadTemplate("repository").Execute(w, templates.RepositoryModel{
			RepoName: rn,
			RepoObj: s,
			RepoHeaderInfo: templates.RepoHeaderTemplateModel{
				RepoName: rn,
				TypeStr: "",
				NodeName: "",
				RepoDescription: s.Description,
				RepoLabelList: nil,
			},
			BranchList: s.BranchIndex,
			TagList: s.TagIndex,
		}))
	}))
}


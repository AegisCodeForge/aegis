package controller

import (
	"fmt"
	"strings"
	"net/http"
	
	. "github.com/bctnry/gitus/routes"
	"github.com/bctnry/gitus/templates"
	"github.com/bctnry/gitus/pkg/gitlib"
)

func bindHistoryController(ctx RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/history/{nodeName}", WithLog(func(w http.ResponseWriter, r *http.Request) {
		rn := r.PathValue("repoName")
		repo, ok := ctx.GitRepositoryList[rn]
		if !ok {
			LogTemplateError(ctx.LoadTemplate("error").Execute(w, templates.ErrorTemplateModel{
				ErrorCode: 404,
				ErrorMessage: fmt.Sprintf("Repository %s not found.", rn),
			}))
			return
		}
		nodeName := r.PathValue("nodeName")
		nodeNameElem := strings.Split(nodeName, ":")
		typeStr := string(nodeNameElem[0])
		cid := string(nodeNameElem[1])
		if string(nodeNameElem[0]) == "branch" {
			err := repo.SyncAllBranchList()
			if err != nil {
				LogTemplateError(ctx.LoadTemplate("error").Execute(w, templates.ErrorTemplateModel{
					ErrorCode: 500,
					ErrorMessage: fmt.Sprintf("Failed at syncing branch list for repository %s: %s", rn, err.Error()),
				}))
				return
			}
			br, ok := repo.BranchIndex[string(nodeNameElem[1])]
			if !ok {
				LogTemplateError(ctx.LoadTemplate("error").Execute(w, templates.ErrorTemplateModel{
					ErrorCode: 404,
					ErrorMessage: fmt.Sprintf("Repository %s not found.", rn),
				}))
				return
			}
			cid = br.HeadId
		}
		cobj, err := repo.ReadObject(cid)
		if err != nil {
			LogTemplateError(ctx.LoadTemplate("error").Execute(w, templates.ErrorTemplateModel{
				ErrorCode: 500,
				ErrorMessage: fmt.Sprintf(
					"Failed to read commit object %s: %s",
					cid,
					err,
				),
			}))
			return
		}
		h, err := repo.GetCommitHistory(cid)
		if err != nil {
			LogTemplateError(ctx.LoadTemplate("error").Execute(w, templates.ErrorTemplateModel{
				ErrorCode: 500,
				ErrorMessage: fmt.Sprintf(
					"Failed to read commit history of object %s: %s",
					cid,
					err,
				),
			}))
			return
		}
		LogTemplateError(ctx.LoadTemplate("commit-history").Execute(
			w,
			templates.CommitHistoryModel{
				RepoHeaderInfo: templates.RepoHeaderTemplateModel{
					RepoName: rn,
					RepoDescription: repo.Description,
					TypeStr: typeStr,
					NodeName: nodeNameElem[1],
					RepoLabelList: nil,
					RepoURL: fmt.Sprintf("%s/repo/%s", ctx.Config.HostName, rn),
				},
				Commit: *(cobj.(*gitlib.CommitObject)),
				CommitHistory: h,
			},
		))
	}))
}

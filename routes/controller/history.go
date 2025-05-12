package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bctnry/gitus/pkg/gitlib"
	"github.com/bctnry/gitus/routes"
	. "github.com/bctnry/gitus/routes"
	"github.com/bctnry/gitus/templates"
)

func bindHistoryController(ctx *RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/history/{nodeName}", WithLog(func(w http.ResponseWriter, r *http.Request) {
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
		nodeName := r.PathValue("nodeName")
		nodeNameElem := strings.Split(nodeName, ":")
		typeStr := string(nodeNameElem[0])
		cid := string(nodeNameElem[1])
		if string(nodeNameElem[0]) == "branch" {
			err := repo.Repository.SyncAllBranchList()
			if err != nil {
				LogTemplateError(ctx.LoadTemplate("error").Execute(w, templates.ErrorTemplateModel{
					ErrorCode: 500,
					ErrorMessage: fmt.Sprintf("Failed at syncing branch list for repository %s: %s", rfn, err.Error()),
				}))
				return
			}
			br, ok := repo.Repository.BranchIndex[string(nodeNameElem[1])]
			if !ok {
				LogTemplateError(ctx.LoadTemplate("error").Execute(w, templates.ErrorTemplateModel{
					ErrorCode: 404,
					ErrorMessage: fmt.Sprintf("Repository %s not found.", rfn),
				}))
				return
			}
			cid = br.HeadId
		}
		cobj, err := repo.Repository.ReadObject(cid)
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
		h, err := repo.Repository.GetCommitHistory(cid)
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
		
		var loginInfo *templates.LoginInfoModel = nil
		if !ctx.Config.PlainMode {
			loginInfo, err = GenerateLoginInfoModel(ctx, r)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
		}
		
		LogTemplateError(ctx.LoadTemplate("commit-history").Execute(
			w,
			templates.CommitHistoryModel{
				RepoHeaderInfo: *GenerateRepoHeader(ctx, repo, typeStr, nodeNameElem[1]),
				Commit: *(cobj.(*gitlib.CommitObject)),
				CommitHistory: h,
				LoginInfo: loginInfo,
				Config: ctx.Config,
			},
		))
	}))
}

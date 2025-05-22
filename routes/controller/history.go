package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/pkg/gitlib"
	"github.com/bctnry/aegis/routes"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindHistoryController(ctx *RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/history/{nodeName}", WithLog(func(w http.ResponseWriter, r *http.Request) {
		rfn := r.PathValue("repoName")
		_, _, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
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
		
		var loginInfo *templates.LoginInfoModel = nil
		if !ctx.Config.PlainMode {
			loginInfo, err = GenerateLoginInfoModel(ctx, r)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
		}
		if !ctx.Config.PlainMode && repo.Status == model.REPO_NORMAL_PRIVATE {
			t := repo.AccessControlList.GetUserPrivilege(loginInfo.UserName)
			if t == nil {
				t = ns.ACL.GetUserPrivilege(loginInfo.UserName)
			}
			if t == nil {
				LogTemplateError(ctx.LoadTemplate("error").Execute(w, templates.ErrorTemplateModel{
					LoginInfo: loginInfo,
					ErrorCode: 403,
					ErrorMessage: "Not enough privilege.",
				}))
				return
			}
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
		h, err := repo.Repository.GetCommitHistoryN(cid, 11)
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
				RepoHeaderInfo: *GenerateRepoHeader(ctx, repo, typeStr, nodeNameElem[1]),
				Commit: *(cobj.(*gitlib.CommitObject)),
				CommitHistory: h[:len(h)-1],
				LoginInfo: loginInfo,
				Config: ctx.Config,
				NextPageCommitId: h[len(h)-1].Id,
			},
		))
	}))
}

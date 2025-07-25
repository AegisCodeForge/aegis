package controller

import (
	"fmt"
	"net/http"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/pkg/gitlib"
	"github.com/bctnry/aegis/routes"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindDiffController(ctx *RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/diff/{commitId}/", WithLog(func(w http.ResponseWriter, r *http.Request){
		var err error
		var loginInfo *templates.LoginInfoModel = nil
		if !ctx.Config.PlainMode {
			loginInfo, err = GenerateLoginInfoModel(ctx, r)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
		}
		if ctx.Config.PlainMode || !CheckGlobalVisibleToUser(ctx, loginInfo) {
			switch ctx.Config.GlobalVisibility {
			case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
				FoundAt(w, "/maintenance-notice")
				return
			case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
				FoundAt(w, "/shutdown-notice")
				return
			case aegis.GLOBAL_VISIBILITY_PRIVATE:
				if !ctx.Config.PlainMode {
					FoundAt(w, "/login")
				} else {
					FoundAt(w, "/private-notice")
				}
				return
			}
		}
		
		rfn := r.PathValue("repoName")
		_, _, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		if err == routes.ErrNotFound {
			ctx.ReportNotFound(rfn, "Repository", "Depot", w, r)
			return
		}
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !ctx.Config.PlainMode {
			loginInfo.IsOwner = (repo.Owner == loginInfo.UserName) || (ns.Owner == loginInfo.UserName)
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
		
		commitId := r.PathValue("commitId")
		cobj, err := repo.Repository.ReadObject(commitId)
		if err != nil {
			ctx.ReportInternalError(
				fmt.Sprintf("Failed to read commit %s: %s", commitId, err.Error()),
				w, r,
			)
			return
		}
		diff, err := repo.Repository.GetDiff(commitId)
		if err != nil {
			ctx.ReportInternalError(
				fmt.Sprintf("Failed to read diff of %s: %s", commitId, err.Error()),
				w, r,
			)
			return
		}
		
		LogTemplateError(ctx.LoadTemplate("diff").Execute(w, templates.DiffTemplateModel{
			Repository: repo,
			RepoHeaderInfo: *GenerateRepoHeader("commit", commitId),
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


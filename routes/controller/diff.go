package controller

import (
	"fmt"
	"net/http"

	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/pkg/gitlib"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindDiffController(ctx *RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/diff/{commitId}/", UseMiddleware(
		[]Middleware{Logged, RateLimit, UseLoginInfo, GlobalVisibility, ErrorGuard},
		ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			rfn := r.PathValue("repoName")
			if !model.ValidRepositoryName(rfn) {
				rc.ReportNotFound(rfn, "Repository", "Depot", w, r)
				return
			}
			_, _, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
			if err == ErrNotFound {
				ctx.ReportNotFound(rfn, "Repository", "Depot", w, r)
				return
			}
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			if repo.Type != model.REPO_TYPE_GIT {
				ctx.ReportNormalError("The repository you have requested isn't a Git repository.", w, r)
				return
			}
			if !ctx.Config.PlainMode {
				rc.LoginInfo.IsOwner = (repo.Owner == rc.LoginInfo.UserName) || (ns.Owner == rc.LoginInfo.UserName)
			}
			
			if !ctx.Config.PlainMode && repo.Status == model.REPO_NORMAL_PRIVATE {
				t := repo.AccessControlList.GetUserPrivilege(rc.LoginInfo.UserName)
				if t == nil {
					t = ns.ACL.GetUserPrivilege(rc.LoginInfo.UserName)
				}
				if t == nil {
					LogTemplateError(ctx.LoadTemplate("error").Execute(w, templates.ErrorTemplateModel{
						LoginInfo: rc.LoginInfo,
						ErrorCode: 403,
						ErrorMessage: "Not enough privilege.",
					}))
					return
				}
			}

			rr := repo.Repository.(*gitlib.LocalGitRepository)
			commitId := r.PathValue("commitId")
			cobj, err := rr.ReadObject(commitId)
			if err != nil {
				ctx.ReportInternalError(
					fmt.Sprintf("Failed to read commit %s: %s", commitId, err.Error()),
					w, r,
				)
				return
			}
			co := cobj.(*gitlib.CommitObject)
			m := make(map[string]string, 0)
			m[co.AuthorInfo.AuthorEmail] = ""
			m[co.CommitterInfo.AuthorEmail] = ""
			m, _ = ctx.DatabaseInterface.ResolveMultipleEmailToUsername(m)
			diff, err := rr.GetDiff(commitId)
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
					EmailUserMapping: m,
				},
				Diff: diff,
				LoginInfo: rc.LoginInfo,
				Config: ctx.Config,
			}))
		},
	))
}


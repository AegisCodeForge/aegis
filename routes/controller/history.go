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
	http.HandleFunc("GET /repo/{repoName}/history/{nodeName}", UseMiddleware(
		[]Middleware{Logged, UseLoginInfo, GlobalVisibility, ErrorGuard}, ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			rfn := r.PathValue("repoName")
			if !model.ValidRepositoryName(rfn) {
				ctx.ReportNotFound(rfn, "Repository", "Depot", w, r)
				return
			}
			_, _, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
			if err == routes.ErrNotFound {
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
			nodeName := r.PathValue("nodeName")
			nodeNameElem := strings.Split(nodeName, ":")
			typeStr := string(nodeNameElem[0])
			cid := string(nodeNameElem[1])
			if string(nodeNameElem[0]) == "branch" {
				err := rr.SyncAllBranchList()
				if err != nil {
					LogTemplateError(ctx.LoadTemplate("error").Execute(w, templates.ErrorTemplateModel{
						ErrorCode: 500,
						ErrorMessage: fmt.Sprintf("Failed at syncing branch list for repository %s: %s", rfn, err.Error()),
					}))
					return
				}
				br, ok := rr.BranchIndex[string(nodeNameElem[1])]
				if !ok {
					LogTemplateError(ctx.LoadTemplate("error").Execute(w, templates.ErrorTemplateModel{
						ErrorCode: 404,
						ErrorMessage: fmt.Sprintf("Repository %s not found.", rfn),
					}))
					return
				}
				cid = br.HeadId
			}
			cobj, err := rr.ReadObject(cid)
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
			h, err := rr.GetCommitHistoryN(cid, 21)
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
			
			m := make(map[string]string, 0)
			for _, k := range h {
				m[k.AuthorInfo.AuthorEmail] = ""
				m[k.CommitterInfo.AuthorEmail] = ""
			}
			_, err = ctx.DatabaseInterface.ResolveMultipleEmailToUsername(m)
			// NOTE: we don't check, we just assume the emails are not verified
			// to anyone if an error occur.
			
			LogTemplateError(ctx.LoadTemplate("commit-history").Execute(
				w,
				templates.CommitHistoryModel{
					Repository: repo,
					RepoHeaderInfo: *GenerateRepoHeader(typeStr, nodeNameElem[1]),
					Commit: *(cobj.(*gitlib.CommitObject)),
					CommitHistory: h[:len(h)-1],
					LoginInfo: rc.LoginInfo,
					Config: ctx.Config,
					NextPageCommitId: h[len(h)-1].Id,
					EmailUserMapping: m,
				},
			))
		},
	))
}

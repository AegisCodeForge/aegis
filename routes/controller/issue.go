package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindIssueController(ctx *routes.RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/issue", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !routes.CheckGlobalVisibleToUser(ctx, loginInfo) {
			switch ctx.Config.GlobalVisibility {
			case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
				routes.FoundAt(w, "/maintenance-notice")
				return
			case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
				routes.FoundAt(w, "/shutdown-notice")
				return
			case aegis.GLOBAL_VISIBILITY_PRIVATE:
				routes.FoundAt(w, "/login")
				return
			}
		}
		rfn := r.PathValue("repoName")
		if !model.ValidRepositoryName(rfn) {
			ctx.ReportNotFound(rfn, "Repository", "Depot", w, r)
			return
		}
		nsName, repoName, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		loginInfo.IsOwner = ns.Owner == loginInfo.UserName || repo.Owner == loginInfo.UserName
		loginInfo.IsStrictOwner = repo.Owner == loginInfo.UserName
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		pStr := strings.TrimSpace(r.URL.Query().Get("p"))
		sStr := strings.TrimSpace(r.URL.Query().Get("s"))
		fStr := strings.TrimSpace(r.URL.Query().Get("f"))
		p64, err := strconv.ParseInt(pStr, 10, 64)
		if err != nil { p64 = 1 }
		s, err := strconv.ParseInt(sStr, 10, 64)
		if err != nil { s = 30 }
		f, err := strconv.ParseInt(fStr, 10, 64)
		if err != nil { f = 0 }
		count, err := ctx.DatabaseInterface.CountIssue(q, nsName, repoName, int(f))
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		pageCount := int(count) / int(s)
		if (int(count) % int(s)) > 0 { pageCount += 1 }
		p := int(p64)
		if p < 1 { p = 1 }
		if p > pageCount { p = pageCount }
		pageInfo := &templates.PageInfoModel{
			PageNum: int(p),
			PageSize: int(s),
			TotalPage: pageCount,
		}
		issueList, err := ctx.DatabaseInterface.SearchIssuePaginated(q, nsName, repoName, int(f), int(p-1), int(s))
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		routes.LogTemplateError(ctx.LoadTemplate("issue/issue-list").Execute(w, &templates.RepositoryIssueListTemplateModel{
			Config: ctx.Config,
			Repository: repo,
			RepoHeaderInfo: routes.GenerateRepoHeader("", ""),
			LoginInfo: loginInfo,
			ErrorMsg: "",
			IssueList: issueList,
			PageInfo: pageInfo,
			FilterType: int(f),
			Query: q,
		}))
	}))

	http.HandleFunc("GET /repo/{repoName}/issue/new", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !routes.CheckGlobalVisibleToUser(ctx, loginInfo) {
			switch ctx.Config.GlobalVisibility {
			case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
				routes.FoundAt(w, "/maintenance-notice")
				return
			case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
				routes.FoundAt(w, "/shutdown-notice")
				return
			case aegis.GLOBAL_VISIBILITY_PRIVATE:
				routes.FoundAt(w, "/login")
				return
			}
		}
		rfn := r.PathValue("repoName")
		if !model.ValidRepositoryName(rfn) {
			ctx.ReportNotFound(rfn, "Repository", "Depot", w, r)
			return
		}
		_, _, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		loginInfo.IsOwner = ns.Owner == loginInfo.UserName || repo.Owner == loginInfo.UserName
		loginInfo.IsStrictOwner = repo.Owner == loginInfo.UserName
		routes.LogTemplateError(ctx.LoadTemplate("issue/new-issue").Execute(w, &templates.RepositoryNewIssueTemplateModel{
			Config: ctx.Config,
			Repository: repo,
			RepoHeaderInfo: routes.GenerateRepoHeader("", ""),
			LoginInfo: loginInfo,
			ErrorMsg: "",
		}))
	}))

	http.HandleFunc("POST /repo/{repoName}/issue/new", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !routes.CheckGlobalVisibleToUser(ctx, loginInfo) {
			switch ctx.Config.GlobalVisibility {
			case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
				routes.FoundAt(w, "/maintenance-notice")
				return
			case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
				routes.FoundAt(w, "/shutdown-notice")
				return
			case aegis.GLOBAL_VISIBILITY_PRIVATE:
				routes.FoundAt(w, "/login")
				return
			}
		}
		rfn := r.PathValue("repoName")
		if !model.ValidRepositoryName(rfn) {
			ctx.ReportNotFound(rfn, "Repository", "Depot", w, r)
			return
		}
		nsName, repoName, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		if !loginInfo.LoggedIn {
			ctx.ReportRedirect("/login", 3, "Not Logged In", "You must log in before creating an issue.", w, r)
			return
		}
		loginInfo.IsOwner = ns.Owner == loginInfo.UserName || repo.Owner == loginInfo.UserName
		nsPriv := ns.ACL.GetUserPrivilege(loginInfo.UserName)
		repoPriv := repo.AccessControlList.GetUserPrivilege(loginInfo.UserName)
		isMember := nsPriv != nil || repoPriv != nil
		if (repo.Status == model.REPO_NORMAL_PRIVATE) && !loginInfo.IsAdmin && !loginInfo.IsOwner && !isMember {
			ctx.ReportNotFound(repoName, "Repository", "", w, r)
			return
		}
		err = r.ParseForm()
		if err != nil {
			ctx.ReportNormalError("Invalid request", w, r)
			return
		}
		title := r.Form.Get("title")
		content := r.Form.Get("content")
		iid, err := ctx.DatabaseInterface.NewRepositoryIssue(nsName, repoName, loginInfo.UserName, title, content)
		routes.FoundAt(w, fmt.Sprintf("/repo/%s/issue/%d", rfn, iid))
	}))
	
	http.HandleFunc("GET /repo/{repoName}/issue/{id}", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !routes.CheckGlobalVisibleToUser(ctx, loginInfo) {
			switch ctx.Config.GlobalVisibility {
			case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
				routes.FoundAt(w, "/maintenance-notice")
				return
			case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
				routes.FoundAt(w, "/shutdown-notice")
				return
			case aegis.GLOBAL_VISIBILITY_PRIVATE:
				routes.FoundAt(w, "/login")
				return
			}
		}
		rfn := r.PathValue("repoName")
		if !model.ValidRepositoryName(rfn) {
			ctx.ReportNotFound(rfn, "Repository", "Depot", w, r)
			return
		}
		nsName, repoName, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		loginInfo.IsOwner = ns.Owner == loginInfo.UserName || repo.Owner == loginInfo.UserName
		loginInfo.IsStrictOwner = repo.Owner == loginInfo.UserName
		nsPriv := ns.ACL.GetUserPrivilege(loginInfo.UserName)
		repoPriv := repo.AccessControlList.GetUserPrivilege(loginInfo.UserName)
		isMember := nsPriv != nil || repoPriv != nil
		if (repo.Status == model.REPO_NORMAL_PRIVATE) && !loginInfo.IsAdmin && !loginInfo.IsOwner && !isMember {
			ctx.ReportNotFound(repoName, "Repository", "", w, r)
			return
		}
		iid, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			ctx.ReportNormalError(err.Error(), w, r)
			return
		}
		issue, err := ctx.DatabaseInterface.GetRepositoryIssue(nsName, repoName, iid)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		eventList, err := ctx.DatabaseInterface.GetAllIssueEvent(nsName, repoName, iid)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		routes.LogTemplateError(ctx.LoadTemplate("issue/single-issue").Execute(w, &templates.RepositorySingleIssueTemplateModel{
			Config: ctx.Config,
			Repository: repo,
			RepoHeaderInfo: routes.GenerateRepoHeader("", ""),
			LoginInfo: loginInfo,
			ErrorMsg: "",
			Issue: issue,
			IssueEventList: eventList,
		}))
	}))
	
	http.HandleFunc("POST /repo/{repoName}/issue/{id}", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !routes.CheckGlobalVisibleToUser(ctx, loginInfo) {
			switch ctx.Config.GlobalVisibility {
			case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
				routes.FoundAt(w, "/maintenance-notice")
				return
			case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
				routes.FoundAt(w, "/shutdown-notice")
				return
			case aegis.GLOBAL_VISIBILITY_PRIVATE:
				routes.FoundAt(w, "/login")
				return
			}
		}
		rfn := r.PathValue("repoName")
		if !model.ValidRepositoryName(rfn) {
			ctx.ReportNotFound(rfn, "Repository", "Depot", w, r)
			return
		}
		nsName, repoName, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		iid, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			ctx.ReportNormalError(err.Error(), w, r)
			return
		}
		if !loginInfo.LoggedIn {
			ctx.ReportRedirect(fmt.Sprintf("/repo/%s/issue/%d", rfn, iid), 3, "Not Logged In", "You must login before commenting on an issue.", w, r)
			return
		}
		loginInfo.IsStrictOwner = repo.Owner == loginInfo.UserName
		loginInfo.IsOwner = ns.Owner == loginInfo.UserName || repo.Owner == loginInfo.UserName
		nsPriv := ns.ACL.GetUserPrivilege(loginInfo.UserName)
		repoPriv := repo.AccessControlList.GetUserPrivilege(loginInfo.UserName)
		isMember := nsPriv != nil || repoPriv != nil
		if (repo.Status == model.REPO_NORMAL_PRIVATE) && !loginInfo.IsAdmin && !loginInfo.IsOwner && !isMember {
			ctx.ReportNotFound(repoName, "Repository", "", w, r)
			return
		}
		err = r.ParseForm()
		if err != nil {
			ctx.ReportNormalError(err.Error(), w, r)
			return
		}
		formType := strings.TrimSpace(r.Form.Get("type"))
		if formType == "unpin" || formType == "pin" {
			switch formType {
			case "unpin":
				err = ctx.DatabaseInterface.SetIssuePriority(nsName, repoName, iid, 0)
			case "pin":
				err = ctx.DatabaseInterface.SetIssuePriority(nsName, repoName, iid, 100)
			}
		} else {
			eType := model.EVENT_COMMENT
			author := loginInfo.UserName
			content := ""
			switch formType {
			case "comment":
				content = r.Form.Get("content")
			case "discarded":
				eType = model.EVENT_CLOSED_AS_DISCARDED
			case "solved":
				eType = model.EVENT_CLOSED_AS_SOLVED
			case "reopen":
				eType = model.EVENT_REOPENED
			}
			err = ctx.DatabaseInterface.NewRepositoryIssueEvent(nsName, repoName, iid, eType, author, content)
		}
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		routes.FoundAt(w, fmt.Sprintf("/repo/%s/issue/%d", rfn, iid))
	}))
}


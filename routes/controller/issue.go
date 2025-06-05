package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindIssueController(ctx *routes.RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/issue", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		rfn := r.PathValue("repoName")
		nsName, repoName, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		loginInfo.IsOwner = ns.Owner == loginInfo.UserName || repo.Owner == loginInfo.UserName
		issueList, err := ctx.DatabaseInterface.GetAllRepositoryIssue(nsName, repoName)
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
		}))
	}))

	http.HandleFunc("GET /repo/{repoName}/issue/new", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		rfn := r.PathValue("repoName")
		_, _, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		loginInfo.IsOwner = ns.Owner == loginInfo.UserName || repo.Owner == loginInfo.UserName
		routes.LogTemplateError(ctx.LoadTemplate("issue/new-issue").Execute(w, &templates.RepositoryNewIssueTemplateModel{
			Config: ctx.Config,
			Repository: repo,
			RepoHeaderInfo: routes.GenerateRepoHeader("", ""),
			LoginInfo: loginInfo,
			ErrorMsg: "",
		}))
	}))

	http.HandleFunc("POST /repo/{repoName}/issue/new", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		rfn := r.PathValue("repoName")
		nsName, repoName, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
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
		ctx.ReportRedirect(fmt.Sprintf("/repo/%s/issue/%d", rfn, iid), 3, "Issue Created", "The issue has been created.", w, r)
		return
	}))
	
	http.HandleFunc("GET /repo/{repoName}/issue/{id}", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		rfn := r.PathValue("repoName")
		nsName, repoName, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
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
		rfn := r.PathValue("repoName")
		nsName, repoName, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		iid, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			ctx.ReportNormalError(err.Error(), w, r)
			return
		}
		if !loginInfo.LoggedIn {
			ctx.ReportRedirect(fmt.Sprintf("/repo/%s/issue/%d", rfn, iid), 3, "Not Logged In", "You must login before commenting on an issue.", w, r)
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
			ctx.ReportNormalError(err.Error(), w, r)
			return
		}
		formType := strings.TrimSpace(r.Form.Get("type"))
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
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		routes.FoundAt(w, fmt.Sprintf("/repo/%s/issue/%d", rfn, iid))
	}))
}


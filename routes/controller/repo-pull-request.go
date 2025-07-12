package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/bctnry/aegis/pkg/aegis/db"
	"github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindRepositoryPullRequestController(ctx *routes.RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/pull-request", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		rfn := r.PathValue("repoName")
		if ctx.Config.PlainMode {
			routes.FoundAt(w, fmt.Sprintf("/repo/%s", rfn))
			return
		}
		_, _, _, s, err := ctx.ResolveRepositoryFullName(rfn)
		if err == routes.ErrNotFound {
			ctx.ReportNotFound(rfn, "Repository", "Depot", w, r)
			return
		}
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		pStr := strings.TrimSpace(r.URL.Query().Get("p"))
		sStr := strings.TrimSpace(r.URL.Query().Get("s"))
		fStr := strings.TrimSpace(r.URL.Query().Get("f"))
		p64, err := strconv.ParseInt(pStr, 10, 64)
		if err != nil { p64 = 1 }
		ps, err := strconv.ParseInt(sStr, 10, 64)
		if err != nil { ps = 30 }
		f, err := strconv.ParseInt(fStr, 10, 64)
		if err != nil { f = 0 }
		count, err := ctx.DatabaseInterface.CountPullRequest(q, s.Namespace, s.Name, int(f))
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		pageCount := int(count) / int(ps)
		if (int(count) % int(ps)) > 0 { pageCount += 1 }
		p := int(p64)
		if p < 1 { p = 1 }
		if p > pageCount { p = pageCount }
		pageInfo := &templates.PageInfoModel{
			PageNum: int(p),
			PageSize: int(ps),
			TotalPage: pageCount,
		}
		prList, err := ctx.DatabaseInterface.SearchPullRequestPaginated(q, s.Namespace, s.Name, int(f), int(p-1), int(ps))
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		routes.LogTemplateError(ctx.LoadTemplate("pull-request/pull-request-list").Execute(w, &templates.RepositoryPullRequestListTemplateModel{
			Config: ctx.Config,
			Repository: s,
			RepoHeaderInfo: &templates.RepoHeaderTemplateModel{
				TypeStr: "", NodeName: "",
			},
			LoginInfo: loginInfo,
			PullRequestList: prList,
			PageInfo: pageInfo,
			Query: q,
			FilterType: int(f),
		}))
	}))
	
	http.HandleFunc("GET /repo/{repoName}/pull-request/{prid}", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		rfn := r.PathValue("repoName")
		if ctx.Config.PlainMode {
			routes.FoundAt(w, fmt.Sprintf("/repo/%s", rfn))
			return
		}
		_, _, _, s, err := ctx.ResolveRepositoryFullName(rfn)
		if err == routes.ErrNotFound {
			ctx.ReportNotFound(rfn, "Repository", "Depot", w, r)
			return
		}
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		pridStr := r.PathValue("prid")
		prid, err := strconv.ParseInt(pridStr, 10, 64)
		if err != nil {
			ctx.ReportNotFound(pridStr, "Pull request", rfn, w, r)
			return
		}
		pr, err := ctx.DatabaseInterface.GetPullRequest(s.Namespace, s.Name, prid)
		if err != nil {
			if err == db.ErrEntityNotFound {
				ctx.ReportRedirect(fmt.Sprintf("/repo/%s/pull-request", rfn), 5, "Not Found", "The pull request you've specified does not exist in this repository.", w, r)
				return
			}
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		pnstr := r.URL.Query().Get("p")
		pn, err := strconv.ParseInt(pnstr, 10, 64)
		if err != nil { pn = 0 }
		preList, err := ctx.DatabaseInterface.GetAllPullRequestEventPaginated(pr.PRAbsId, int(pn), 30)
		routes.LogTemplateError(ctx.LoadTemplate("pull-request/single-pull-request").Execute(w, &templates.RepositorySinglePullRequestTemplateModel{
			Config: ctx.Config,
			Repository: s,
			RepoHeaderInfo: &templates.RepoHeaderTemplateModel{
				TypeStr: "", NodeName: "",
			},
			LoginInfo: loginInfo,
			PullRequest: pr,
			PullRequestEventList: preList,
			PageNum: int(pn),
		}))
	}))
	

	http.HandleFunc("POST /repo/{repoName}/pull-request/{prid}", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		rfn := r.PathValue("repoName")
		if ctx.Config.PlainMode {
			routes.FoundAt(w, fmt.Sprintf("/repo/%s", rfn))
			return
		}
		_, _, _, s, err := ctx.ResolveRepositoryFullName(rfn)
		if err == routes.ErrNotFound {
			ctx.ReportNotFound(rfn, "Repository", "Depot", w, r)
			return
		}
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		pridStr := r.PathValue("prid")
		if !loginInfo.LoggedIn {
			ctx.ReportRedirect(fmt.Sprintf("/repo/%s/pull-request/%s", rfn, pridStr), 5, "Not Logged In", "You must log in before performing this action.", w, r)
			return
		}
		prid, err := strconv.ParseInt(pridStr, 10, 64)
		if err != nil {
			ctx.ReportNotFound(pridStr, "Pull request", rfn, w, r)
			return
		}
		pr, err := ctx.DatabaseInterface.GetPullRequest(s.Namespace, s.Name, prid)
		if err != nil {
			if err == db.ErrEntityNotFound {
				ctx.ReportRedirect(fmt.Sprintf("/repo/%s/pull-request", rfn), 5, "Not Found", "The pull request you've specified does not exist in this repository.", w, r)
				return
			}
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		err = r.ParseForm()
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		rt := r.Form.Get("type")
		returnPath := fmt.Sprintf("/repo/%s/pull-request/%d", rfn, prid)
		switch rt {
		case "comment":
			_, err = ctx.DatabaseInterface.CommentOnPullRequest(pr.PRAbsId, loginInfo.UserName, r.Form.Get("content"))
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			routes.FoundAt(w, returnPath)
		case "merge-check":
			e, err := ctx.DatabaseInterface.CheckPullRequestMergeConflict(pr.PRAbsId)
			fmt.Println("ee", e)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			routes.FoundAt(w, returnPath)
		case "close-as-merged":
			err = ctx.DatabaseInterface.CheckAndMergePullRequest(pr.PRAbsId, loginInfo.UserName)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			routes.FoundAt(w, returnPath)
		case "close-as-not-merged":
			err = ctx.DatabaseInterface.ClosePullRequestAsNotMerged(pr.PRAbsId, loginInfo.UserName)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			routes.FoundAt(w, returnPath)
		case "reopen":
			err = ctx.DatabaseInterface.ReopenPullRequest(pr.PRAbsId, loginInfo.UserName)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			routes.FoundAt(w, returnPath)
		default:
			ctx.ReportNormalError("Invalid Request", w, r)
			return
		}

		routes.FoundAt(w, returnPath)
	}))
	
	http.HandleFunc("GET /repo/{repoName}/pull-request/new", routes.WithLog(func(w http.ResponseWriter, r *http.Request){
		rfn := r.PathValue("repoName")
		if ctx.Config.PlainMode {
			routes.FoundAt(w, fmt.Sprintf("/repo/%s", rfn))
			return
		}
		_, _, _, s, err := ctx.ResolveRepositoryFullName(rfn)
		if err == routes.ErrNotFound {
			ctx.ReportNotFound(rfn, "Repository", "Depot", w, r)
			return
		}
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !loginInfo.LoggedIn {
			ctx.ReportRedirect(fmt.Sprintf("/repo/%s", rfn), 5, "Not Logged In", "You must log in before creating a pull request.", w, r)
			return
		}
		err = s.Repository.SyncAllBranchList()
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		receiverBranch := strings.TrimSpace(r.URL.Query().Get("recv-br"))
		if len(receiverBranch) > 0 {
			_, ok := s.Repository.BranchIndex[receiverBranch]
			if !ok {
				ctx.ReportRedirect(fmt.Sprintf("/repo/%s/pull-request/new", rfn), 5, "Not Found", fmt.Sprintf("Branch \"%s\" does not exist in repository %s. Please choose an existing branch.", receiverBranch, rfn), w, r)
				return
			}
		}
		providerRepositoryName := strings.TrimSpace(r.URL.Query().Get("repo"))
		if len(providerRepositoryName) <= 0 {
			fr, err := ctx.DatabaseInterface.GetForkRepositoryOfUser(loginInfo.UserName, s.Namespace, s.Name)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			routes.LogTemplateError(ctx.LoadTemplate("pull-request/new-pull-request").Execute(w, &templates.RepositoryNewPullRequestTemplateModel{
				Config: ctx.Config,
				Repository: s,
				LoginInfo: loginInfo,
				ProviderRepository: fr,
			}))
		} else {
			
			_, _, _, provider, err := ctx.ResolveRepositoryFullName(providerRepositoryName)
			if err == routes.ErrNotFound {
				ctx.ReportNotFound(provider.FullName(), "Repository", "Depot", w, r)
				return
			}
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			branchNameList := make([]string, 0)
			err = provider.Repository.SyncAllBranchList()
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			for k := range provider.Repository.BranchIndex {
				branchNameList = append(branchNameList, k)
			}
			routes.LogTemplateError(ctx.LoadTemplate("pull-request/new-pull-request").Execute(w, &templates.RepositoryNewPullRequestTemplateModel{
				Config: ctx.Config,
				Repository: s,
				LoginInfo: loginInfo,
				ReceiverBranch: receiverBranch,
				ChosenProviderRepository: provider,
				ProviderBranchList: branchNameList,
			}))
		}
	}))
	
	http.HandleFunc("POST /repo/{repoName}/pull-request/new", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		rfn := r.PathValue("repoName")
		if ctx.Config.PlainMode {
			routes.FoundAt(w, fmt.Sprintf("/repo/%s", rfn))
			return
		}
		_, _, _, s, err := ctx.ResolveRepositoryFullName(rfn)
		if err == routes.ErrNotFound {
			ctx.ReportNotFound(rfn, "Repository", "Depot", w, r)
			return
		}
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !loginInfo.LoggedIn {
			ctx.ReportRedirect(fmt.Sprintf("/repo/%s", rfn), 5, "Not Logged In", "You must log in before creating a pull request.", w, r)
			return
		}
		err = r.ParseForm()
		if err != nil {
			ctx.ReportNormalError("Invalid request", w, r)
			return
		}
		title := r.Form.Get("title")
		receiverBranch := r.Form.Get("receiver-branch")
		providerNamespace := r.Form.Get("provider-namespace")
		providerName := r.Form.Get("provider-name")
		providerBranch := r.Form.Get("provider-branch")
		resId, err := ctx.DatabaseInterface.NewPullRequest(loginInfo.UserName, title, s.Namespace, s.Name, receiverBranch, providerNamespace, providerName, providerBranch)
		if err != nil {
			ctx.ReportRedirect(fmt.Sprintf("/repo/%s/pull-request/new", rfn), 0, "Internal Error", fmt.Sprintf("Failed to create pull request: %s", err.Error()), w, r)
			return
		}
		routes.FoundAt(w, fmt.Sprintf("/repo/%s/pull-request/%d", rfn, resId))
	}))
}

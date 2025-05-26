package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/pkg/auxfuncs"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindRepositorySettingController(ctx *RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/setting", WithLog(func(w http.ResponseWriter, r *http.Request){
		rfn := r.PathValue("repoName")
		_, _, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if ctx.Config.UseNamespace && ns == nil {
			ctx.ReportNotFound(repo.Namespace, "Namespace", "depot", w, r)
			return
		}
		if repo == nil {
			ctx.ReportNotFound(repo.Name, "Repository", repo.Namespace, w, r)
			return
		}
		repoPath := fmt.Sprintf("/repo/%s", repo.FullName())
		// NOTE: we don't support editing namespace from web ui when in plain mode.
		if ctx.Config.PlainMode { FoundAt(w, repoPath); return }
		
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !loginInfo.LoggedIn { FoundAt(w, repoPath); return }
		
		isRepoOwner := repo.Owner == loginInfo.UserName
		isNsOwner := ns.Owner == loginInfo.UserName
		repoPriv := repo.AccessControlList.GetUserPrivilege(loginInfo.UserName)
		nsPriv := ns.ACL.GetUserPrivilege(loginInfo.UserName)
		isSettingMember := repoPriv.HasSettingPrivilege() || nsPriv.HasSettingPrivilege()
		if !loginInfo.IsAdmin && !isRepoOwner && !isNsOwner && !isSettingMember {
			LogTemplateError(ctx.LoadTemplate("repo-setting/_redirect-with-message").Execute(w, templates.RedirectWithMessageModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				Timeout: 3,
				RedirectUrl: fmt.Sprintf("/repo/%s", rfn),
				MessageTitle: "Not enough permission",
				MessageText: "You don't have enough permission to change the settings of this repository.",
			}))
			return
		}
		loginInfo.IsSettingMember = true
		LogTemplateError(ctx.LoadTemplate("repo-setting/change-info").Execute(w, templates.RepositorySettingTemplateModel{
			Config: ctx.Config,
			Repository: repo,
			RepoHeaderInfo: GenerateRepoHeader(ctx, repo, "", ""),
			RepoFullName: rfn,
			LoginInfo: loginInfo,
		}))
	}))

	http.HandleFunc("POST /repo/{repoName}/setting", WithLog(func(w http.ResponseWriter, r *http.Request) {
		rfn := r.PathValue("repoName")
		_, _, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if ctx.Config.UseNamespace && ns == nil {
			ctx.ReportNotFound(repo.Namespace, "Namespace", "depot", w, r)
			return
		}
		if repo == nil {
			ctx.ReportNotFound(repo.Name, "Repository", repo.Namespace, w, r)
			return
		}
		repoPath := fmt.Sprintf("/repo/%s", rfn)
			
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !loginInfo.LoggedIn { FoundAt(w, repoPath); return }

		isRepoOwner := repo.Owner == loginInfo.UserName
		isNsOwner := ns.Owner == loginInfo.UserName
		isOwner := isRepoOwner || isNsOwner
		repoPriv := repo.AccessControlList.GetUserPrivilege(loginInfo.UserName)
		nsPriv := ns.ACL.GetUserPrivilege(loginInfo.UserName)
		isSettingMember := repoPriv.HasSettingPrivilege() || nsPriv.HasSettingPrivilege()
		if !loginInfo.IsAdmin && !isOwner && !isSettingMember {
			LogTemplateError(ctx.LoadTemplate("repo-setting/_redirect-with-message").Execute(w, templates.RedirectWithMessageModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				Timeout: 3,
				RedirectUrl: fmt.Sprintf("/repo/%s", rfn),
				MessageTitle: "Not enough permission",
				MessageText: "You don't have enough permission to change the settings of this repository.",
			}))
			return
		}
		loginInfo.IsSettingMember = isSettingMember
		err = r.ParseForm()
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		anythingChanged := false
		newOwner := strings.TrimSpace(r.Form.Get("owner"))
		if repo.Owner != newOwner {
			if !isOwner && !isNsOwner && !loginInfo.IsAdmin && newOwner != repo.Owner {
				LogTemplateError(ctx.LoadTemplate("repo-setting/_redirect-with-message").Execute(w, templates.RedirectWithMessageModel{
					Config: ctx.Config,
					LoginInfo: loginInfo,
					Timeout: 3,
					RedirectUrl: fmt.Sprintf("/repo/%s/setting", rfn),
					MessageTitle: "Not enough permission",
					MessageText: "You don't have enough permission to move the ownership of this repository.",
				}))
				return
			}
			repo.Owner = newOwner
			anythingChanged = true
		}
		newDescription := strings.TrimSpace(r.Form.Get("description"))
		if repo.Description != newDescription {
			anythingChanged = true
		}
		i, err := strconv.Atoi(r.Form.Get("status"))
		if err != nil {
			LogTemplateError(ctx.LoadTemplate("namespace-setting/change-info").Execute(w, &templates.RepositorySettingTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				Repository: repo,
				RepoHeaderInfo: GenerateRepoHeader(ctx, repo, "", ""),
				RepoFullName: rfn,
				ErrorMsg: fmt.Sprintf("Invalid status value. Please try again."),
			}))
			return
		}
		hasEditInfoPriv := loginInfo.IsAdmin || isOwner || (repoPriv != nil && repoPriv.EditInfo) || (nsPriv != nil && nsPriv.EditInfo)
		if anythingChanged && !hasEditInfoPriv {
			LogTemplateError(ctx.LoadTemplate("repo-setting/change-info").Execute(w, &templates.RepositorySettingTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				Repository: repo,
				RepoHeaderInfo: GenerateRepoHeader(ctx, repo, "", ""),
				RepoFullName: rfn,
				ErrorMsg: fmt.Sprintf("Not enough privilege."),
			}))
			return
		}
		repo.Description = newDescription
		repo.Status = model.AegisRepositoryStatus(i)
		err = ctx.DatabaseInterface.UpdateRepositoryInfo(repo.Namespace, repo.Name, repo)
		if err != nil {
			LogTemplateError(ctx.LoadTemplate("repo-setting/change-info").Execute(w, &templates.RepositorySettingTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				Repository: repo,
				RepoHeaderInfo: GenerateRepoHeader(ctx, repo, "", ""),
				RepoFullName: rfn,
				ErrorMsg: fmt.Sprintf("Internal error while updating repo: %s.", err.Error()),
			}))
			return
		}
		LogTemplateError(ctx.LoadTemplate("repo-setting/change-info").Execute(w, &templates.RepositorySettingTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			Repository: repo,
			RepoHeaderInfo: GenerateRepoHeader(ctx, repo, "", ""),
			RepoFullName: rfn,
			ErrorMsg: "Updated.",
		}))
	}))

	http.HandleFunc("GET /repo/{repoName}/delete", WithLog(func(w http.ResponseWriter, r *http.Request) {
		rfn := r.PathValue("repoName")
		_, _, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if ctx.Config.UseNamespace && ns == nil {
			ctx.ReportNotFound(repo.Namespace, "Namespace", "depot", w, r)
			return
		}
		if repo == nil {
			ctx.ReportNotFound(repo.Name, "Repository", repo.Namespace, w, r)
			return
		}
		repoPath := fmt.Sprintf("/repo/%s", repo.FullName())
		// NOTE: we don't support editing namespace from web ui when in plain mode.
		if ctx.Config.PlainMode { FoundAt(w, repoPath); return }
		
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !loginInfo.LoggedIn { FoundAt(w, repoPath); return }

		isRepoOwner := repo.Owner == loginInfo.UserName
		isNsOwner := ns.Owner == loginInfo.UserName
		repoPriv := repo.AccessControlList.GetUserPrivilege(loginInfo.UserName)
		nsPriv := ns.ACL.GetUserPrivilege(loginInfo.UserName)
		isDeleteCapableMember := (repoPriv != nil && repoPriv.DeleteRepository) || (nsPriv != nil && nsPriv.DeleteRepository)
		if !loginInfo.IsAdmin && !isRepoOwner && !isNsOwner && !isDeleteCapableMember {
			LogTemplateError(ctx.LoadTemplate("namespace-setting/_redirect-with-message").Execute(w, templates.RedirectWithMessageModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				Timeout: 3,
				RedirectUrl: fmt.Sprintf("/repo/%s/setting", rfn),
				MessageTitle: "Failed to delete repository",
				MessageText: "Not enough privilege.",
			}))
			return
		}

		err = ctx.DatabaseInterface.HardDeleteRepository(repo.Namespace, repo.Name)
		if err != nil {
			LogTemplateError(ctx.LoadTemplate("namespace-setting/_redirect-with-message").Execute(w, templates.RedirectWithMessageModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				Timeout: 3,
				RedirectUrl: fmt.Sprintf("/repo/%s/setting", rfn),
				MessageTitle: "Internal error",
				MessageText: fmt.Sprintf("Failed to delete repository: %s.", err.Error()),
			}))
			return
		}
		if ctx.Config.UseNamespace {
			FoundAt(w, fmt.Sprint("/s/%s", ns.Name))
		} else {
			FoundAt(w, "/")
		}
	}))

	http.HandleFunc("GET /repo/{repoName}/member", WithLog(func(w http.ResponseWriter, r *http.Request) {
		rfn := r.PathValue("repoName")
		_, _, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if ctx.Config.UseNamespace && ns == nil {
			ctx.ReportNotFound(repo.Namespace, "Namespace", "depot", w, r)
			return
		}
		if repo == nil {
			ctx.ReportNotFound(repo.Name, "Repository", repo.Namespace, w, r)
			return
		}
		repoPath := fmt.Sprintf("/repo/%s", rfn)

		
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !loginInfo.LoggedIn { FoundAt(w, repoPath); return }
		
		isRepoOwner := repo.Owner == loginInfo.UserName
		isNsOwner := ns.Owner == loginInfo.UserName
		isOwner := isRepoOwner || isNsOwner
		repoPriv := repo.AccessControlList.GetUserPrivilege(loginInfo.UserName)
		nsPriv := ns.ACL.GetUserPrivilege(loginInfo.UserName)
		isSettingMember := repoPriv.HasSettingPrivilege() || nsPriv.HasSettingPrivilege()
		if !loginInfo.IsAdmin && !isOwner && !isSettingMember {
			LogTemplateError(ctx.LoadTemplate("repo-setting/_redirect-with-message").Execute(w, templates.RedirectWithMessageModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				Timeout: 3,
				RedirectUrl: fmt.Sprintf("/repo/%s/setting", rfn),
				MessageTitle: "Not enough permission",
				MessageText: "You don't have enough permission to change the settings of this repository.",
			}))
			return
		}
		loginInfo.IsSettingMember = isSettingMember

		var userList map[string]*model.ACLTuple
		if repo.AccessControlList == nil {
			userList = nil
		} else {
			userList = repo.AccessControlList.ACL
		}
		totalMemberCount := len(userList)
		pageInfo, err := GeneratePageInfo(r, totalMemberCount)
		k := auxfuncs.SortedKeys(userList)
		
		stidx := (pageInfo.PageNum-1)*pageInfo.PageSize
		eidx := min(stidx+pageInfo.PageSize, totalMemberCount)
		k = k[stidx:eidx]
		page := make(map[string]*model.ACLTuple, 4)
		if eidx - stidx > 0 {
			for _, item := range k {
				page[item] = userList[item]
			}
			userList = page
		}
		
		LogTemplateError(ctx.LoadTemplate("repo-setting/member-list").Execute(w, templates.RepositorySettingMemberListTemplateModel{
			Config: ctx.Config,
			Repository: repo,
			RepoHeaderInfo: GenerateRepoHeader(ctx, repo, "", ""),
			RepoFullName: rfn,
			LoginInfo: loginInfo,
			ACL: page,
			PageInfo: pageInfo,
		}))
	}))
	
	http.HandleFunc("POST /repo/{repoName}/member", WithLog(func(w http.ResponseWriter, r *http.Request) {
		rfn := r.PathValue("repoName")
		_, _, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if ctx.Config.UseNamespace && ns == nil {
			ctx.ReportNotFound(repo.Namespace, "Namespace", "depot", w, r)
			return
		}
		if repo == nil {
			ctx.ReportNotFound(repo.Name, "Repository", repo.Namespace, w, r)
			return
		}
		repoPath := fmt.Sprintf("/repo/%s", rfn)

		
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !loginInfo.LoggedIn { FoundAt(w, repoPath); return }
		
		isRepoOwner := repo.Owner == loginInfo.UserName
		isNsOwner := ns.Owner == loginInfo.UserName
		isOwner := isRepoOwner || isNsOwner
		repoPriv := repo.AccessControlList.GetUserPrivilege(loginInfo.UserName)
		nsPriv := ns.ACL.GetUserPrivilege(loginInfo.UserName)
		hasAddMemberPriv := (nsPriv != nil && nsPriv.AddMember) || (repoPriv != nil && repoPriv.AddMember)
		if !loginInfo.IsAdmin && !isOwner && !hasAddMemberPriv {
			LogTemplateError(ctx.LoadTemplate("repo-setting/_redirect-with-message").Execute(w, templates.RedirectWithMessageModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				Timeout: 3,
				RedirectUrl: fmt.Sprintf("/repo/%s/setting", rfn),
				MessageTitle: "Not enough permission",
				MessageText: "You don't have enough permission to change the settings of this repository.",
			}))
			return
		}

		err = r.ParseForm()
		if err != nil {
			LogTemplateError(ctx.LoadTemplate("repo-setting/_redirect-with-message").Execute(w, templates.RedirectWithMessageModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				Timeout: 3,
				RedirectUrl: fmt.Sprintf("/repo/%s/setting", rfn),
				MessageTitle: "Invalid request",
				MessageText: "Failed to parse request.",
			}))
			return
		}
		username := strings.TrimSpace(r.Form.Get("username"))
		if len(username) <= 0 {
			LogTemplateError(ctx.LoadTemplate("namespace-setting/_redirect-with-message").Execute(w, templates.RedirectWithMessageModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				Timeout: 3,
				RedirectUrl: fmt.Sprintf("/repo/%s/member", rfn),
				MessageTitle: "Invalid request",
				MessageText: "User name cannot be empty.",
			}))
			return
		}
		t := &model.ACLTuple{
			AddMember: len(r.Form.Get("addMember")) > 0,
			DeleteMember: len(r.Form.Get("deleteMember")) > 0,
			EditMember: len(r.Form.Get("editMember")) > 0,
			AddRepository: false,
			EditInfo: len(r.Form.Get("editInfo")) > 0,
			PushToRepository: len(r.Form.Get("pushToRepo")) > 0,
			ArchiveRepository: len(r.Form.Get("archiveRepo")) > 0,
			DeleteRepository: len(r.Form.Get("deleteRepo")) > 0,
		}
		err = ctx.DatabaseInterface.SetRepositoryACL(repo.Namespace, repo.Name, username, t)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		FoundAt(w, fmt.Sprintf("/repo/%s/member", rfn))
	}))

	http.HandleFunc("GET /repo/{repoName}/member/{userName}/edit", WithLog(func(w http.ResponseWriter, r *http.Request) {
		rfn := r.PathValue("repoName")
		_, _, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if ctx.Config.UseNamespace && ns == nil {
			ctx.ReportNotFound(repo.Namespace, "Namespace", "depot", w, r)
			return
		}
		if repo == nil {
			ctx.ReportNotFound(repo.Name, "Repository", repo.Namespace, w, r)
			return
		}
		repoPath := fmt.Sprintf("/repo/%s", rfn)

		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !loginInfo.LoggedIn { FoundAt(w, repoPath); return }
		
		isRepoOwner := repo.Owner == loginInfo.UserName
		isNsOwner := ns.Owner == loginInfo.UserName
		isOwner := isRepoOwner || isNsOwner
		repoPriv := repo.AccessControlList.GetUserPrivilege(loginInfo.UserName)
		nsPriv := ns.ACL.GetUserPrivilege(loginInfo.UserName)
		hasEditMemberPriv := (nsPriv != nil && nsPriv.EditMember) || (repoPriv != nil && repoPriv.EditMember)
		if !loginInfo.IsAdmin && !isOwner && !hasEditMemberPriv {
			LogTemplateError(ctx.LoadTemplate("repo-setting/_redirect-with-message").Execute(w, templates.RedirectWithMessageModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				Timeout: 3,
				RedirectUrl: fmt.Sprintf("/repo/%s/member", rfn),
				MessageTitle: "Not enough permission",
				MessageText: "You don't have enough permission to change the settings of this repository.",
			}))
			return
		}

		err = r.ParseForm()
		if err != nil {
			LogTemplateError(ctx.LoadTemplate("repo-setting/_redirect-with-message").Execute(w, templates.RedirectWithMessageModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				Timeout: 3,
				RedirectUrl: fmt.Sprintf("/repo/%s/setting", rfn),
				MessageTitle: "Invalid request",
				MessageText: "Failed to parse request.",
			}))
			return
		}
		
		targetUsername := r.PathValue("userName")
		userPriv := repo.AccessControlList.GetUserPrivilege(targetUsername)
		fmt.Println("userpriv", userPriv)
		LogTemplateError(ctx.LoadTemplate("repo-setting/edit-member").Execute(w, templates.RepositorySettingEditMemberTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			Repository: repo,
			RepoFullName: rfn,
			RepoHeaderInfo: GenerateRepoHeader(ctx, repo, "", ""),
			Username: targetUsername,
			ACLTuple: userPriv,
		}))
	}))

	http.HandleFunc("POST /repo/{repoName}/member/{userName}/edit", WithLog(func(w http.ResponseWriter, r *http.Request) {
		rfn := r.PathValue("repoName")
		_, _, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if ctx.Config.UseNamespace && ns == nil {
			ctx.ReportNotFound(repo.Namespace, "Namespace", "depot", w, r)
			return
		}
		if repo == nil {
			ctx.ReportNotFound(repo.Name, "Repository", repo.Namespace, w, r)
			return
		}
		repoPath := fmt.Sprintf("/repo/%s", rfn)

		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !loginInfo.LoggedIn { FoundAt(w, repoPath); return }
		
		isRepoOwner := repo.Owner == loginInfo.UserName
		isNsOwner := ns.Owner == loginInfo.UserName
		isOwner := isRepoOwner || isNsOwner
		repoPriv := repo.AccessControlList.GetUserPrivilege(loginInfo.UserName)
		nsPriv := ns.ACL.GetUserPrivilege(loginInfo.UserName)
		hasEditMemberPriv := (nsPriv != nil && nsPriv.EditMember) || (repoPriv != nil && repoPriv.EditMember)
		if !loginInfo.IsAdmin && !isOwner && !hasEditMemberPriv {
			LogTemplateError(ctx.LoadTemplate("repo-setting/_redirect-with-message").Execute(w, templates.RedirectWithMessageModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				Timeout: 3,
				RedirectUrl: fmt.Sprintf("/repo/%s/member", rfn),
				MessageTitle: "Not enough permission",
				MessageText: "You don't have enough permission to change the settings of this repository.",
			}))
			return
		}

		err = r.ParseForm()
		if err != nil {
			LogTemplateError(ctx.LoadTemplate("repo-setting/_redirect-with-message").Execute(w, templates.RedirectWithMessageModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				Timeout: 3,
				RedirectUrl: fmt.Sprintf("/repo/%s/setting", rfn),
				MessageTitle: "Invalid request",
				MessageText: "Failed to parse request.",
			}))
			return
		}
		
		targetUsername := r.PathValue("userName")
		t := &model.ACLTuple{
			AddMember: len(r.Form.Get("addMember")) > 0,
			DeleteMember: len(r.Form.Get("deleteMember")) > 0,
			EditMember: len(r.Form.Get("editMember")) > 0,
			AddRepository: false,
			EditInfo: len(r.Form.Get("editInfo")) > 0,
			PushToRepository: len(r.Form.Get("pushToRepo")) > 0,
			ArchiveRepository: len(r.Form.Get("archiveRepo")) > 0,
			DeleteRepository: len(r.Form.Get("deleteRepo")) > 0,
		}
		err = ctx.DatabaseInterface.SetRepositoryACL(repo.Namespace, repo.Name, targetUsername, t)
		if err != nil {
			LogTemplateError(ctx.LoadTemplate("repo-setting/_redirect-with-message").Execute(w, templates.RedirectWithMessageModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				Timeout: 0,
				RedirectUrl: fmt.Sprintf("/repo/%s/member", rfn),
				MessageTitle: "Failed to update member privilege",
				MessageText: fmt.Sprintf("Failed to update member privilege: %s. Please contact site owner.", err.Error()),
			}))
			return
		}
		FoundAt(w, fmt.Sprintf("/repo/%s/member", rfn))
	}))
}


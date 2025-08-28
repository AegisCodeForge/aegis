package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/pkg/auxfuncs"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindRepositorySettingController(ctx *RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/setting", WithLog(func(w http.ResponseWriter, r *http.Request){
		rfn := r.PathValue("repoName")
		
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !CheckGlobalVisibleToUser(ctx, loginInfo) {
			switch ctx.Config.GlobalVisibility {
			case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
				FoundAt(w, "/maintenance-notice")
				return
			case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
				FoundAt(w, "/shutdown-notice")
				return
			case aegis.GLOBAL_VISIBILITY_PRIVATE:
				FoundAt(w, "/login")
				return
			}
		}
		
		nsName, repoName, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if ctx.Config.UseNamespace && ns == nil {
			ctx.ReportNotFound(repo.Namespace, "Namespace", "depot", w, r)
			return
		}
		if repo == nil {
			ctx.ReportNotFound(repoName, "Repository", nsName, w, r)
			return
		}
		repoPath := fmt.Sprintf("/repo/%s", repo.FullName())
		if !loginInfo.LoggedIn { FoundAt(w, repoPath); return }
		// NOTE: we don't support editing namespace from web ui when in plain mode.
		if ctx.Config.PlainMode { FoundAt(w, repoPath); return }
		

		isRepoOwner := repo.Owner == loginInfo.UserName
		isNsOwner := ns.Owner == loginInfo.UserName
		loginInfo.IsOwner = isRepoOwner || isNsOwner
		repoPriv := repo.AccessControlList.GetUserPrivilege(loginInfo.UserName)
		nsPriv := ns.ACL.GetUserPrivilege(loginInfo.UserName)
		isSettingMember := repoPriv.HasSettingPrivilege() || nsPriv.HasSettingPrivilege()
		if !loginInfo.IsAdmin && !isRepoOwner && !isNsOwner && !isSettingMember {
			ctx.ReportRedirect(fmt.Sprintf("/repo/%s", rfn), 0,
				"Not enouhg privilege",
				"Your user account seems to not have enough privilege for this action.",
				w, r,
			)
			return
		}
		loginInfo.IsSettingMember = true
		LogTemplateError(ctx.LoadTemplate("repo-setting/change-info").Execute(w, templates.RepositorySettingTemplateModel{
			Config: ctx.Config,
			Repository: repo,
			RepoFullName: rfn,
			LoginInfo: loginInfo,
		}))
	}))

	http.HandleFunc("POST /repo/{repoName}/setting", WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !CheckGlobalVisibleToUser(ctx, loginInfo) {
			switch ctx.Config.GlobalVisibility {
			case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
				FoundAt(w, "/maintenance-notice")
				return
			case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
				FoundAt(w, "/shutdown-notice")
				return
			case aegis.GLOBAL_VISIBILITY_PRIVATE:
				FoundAt(w, "/login")
				return
			}
		}
		rfn := r.PathValue("repoName")
		nsName, repoName, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if ctx.Config.UseNamespace && ns == nil {
			ctx.ReportNotFound(repo.Namespace, "Namespace", "depot", w, r)
			return
		}
		if repo == nil {
			ctx.ReportNotFound(repoName, "Repository", nsName, w, r)
			return
		}
		repoPath := fmt.Sprintf("/repo/%s", rfn)
		if !loginInfo.LoggedIn { FoundAt(w, repoPath); return }
			
		isRepoOwner := repo.Owner == loginInfo.UserName
		isNsOwner := ns.Owner == loginInfo.UserName
		loginInfo.IsOwner = isRepoOwner || isNsOwner
		isOwner := isRepoOwner || isNsOwner
		repoPriv := repo.AccessControlList.GetUserPrivilege(loginInfo.UserName)
		nsPriv := ns.ACL.GetUserPrivilege(loginInfo.UserName)
		isSettingMember := repoPriv.HasSettingPrivilege() || nsPriv.HasSettingPrivilege()
		if !loginInfo.IsAdmin && !isOwner && !isSettingMember {
			ctx.ReportRedirect(fmt.Sprintf("/repo/%s", rfn), 0,
				"Not enough privilege",
				"Your user account seems to not have enough privilege for this action.",
				w, r,
			)
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
				ctx.ReportRedirect(fmt.Sprintf("/repo/%s/setting", rfn), 0,
					"Not enough privilege",
					"Your user account seems to not have enough privilege for this action.",
					w, r,
				)
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
				RepoFullName: rfn,
				ErrorMsg: "Invalid status value. Please try again.",
			}))
			return
		}
		hasEditInfoPriv := loginInfo.IsAdmin || isOwner || (repoPriv != nil && repoPriv.EditInfo) || (nsPriv != nil && nsPriv.EditInfo)
		if anythingChanged && !hasEditInfoPriv {
			LogTemplateError(ctx.LoadTemplate("repo-setting/change-info").Execute(w, &templates.RepositorySettingTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				Repository: repo,
				RepoFullName: rfn,
				ErrorMsg: "Not enough privilege.",
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
				RepoFullName: rfn,
				ErrorMsg: fmt.Sprintf("Internal error while updating repo: %s.", err.Error()),
			}))
			return
		}
		LogTemplateError(ctx.LoadTemplate("repo-setting/change-info").Execute(w, &templates.RepositorySettingTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			Repository: repo,
			RepoFullName: rfn,
			ErrorMsg: "Updated.",
		}))
	}))

	http.HandleFunc("GET /repo/{repoName}/delete", WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !CheckGlobalVisibleToUser(ctx, loginInfo) {
			switch ctx.Config.GlobalVisibility {
			case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
				FoundAt(w, "/maintenance-notice")
				return
			case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
				FoundAt(w, "/shutdown-notice")
				return
			case aegis.GLOBAL_VISIBILITY_PRIVATE:
				FoundAt(w, "/login")
				return
			}
		}
		rfn := r.PathValue("repoName")
		nsName, repoName, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if ctx.Config.UseNamespace && ns == nil {
			ctx.ReportNotFound(repo.Namespace, "Namespace", "depot", w, r)
			return
		}
		if repo == nil {
			ctx.ReportNotFound(repoName, "Repository", nsName, w, r)
			return
		}
		repoPath := fmt.Sprintf("/repo/%s", repo.FullName())
		if !loginInfo.LoggedIn { FoundAt(w, repoPath); return }

		isRepoOwner := repo.Owner == loginInfo.UserName
		isNsOwner := ns.Owner == loginInfo.UserName
		loginInfo.IsOwner = isRepoOwner || isNsOwner
		repoPriv := repo.AccessControlList.GetUserPrivilege(loginInfo.UserName)
		nsPriv := ns.ACL.GetUserPrivilege(loginInfo.UserName)
		canDeleteRepo := (repoPriv != nil && repoPriv.DeleteRepository) || (nsPriv != nil && nsPriv.DeleteRepository)
		if !loginInfo.IsAdmin && !isRepoOwner && !isNsOwner && !canDeleteRepo {
			ctx.ReportRedirect(fmt.Sprintf("/repo/%s/setting", rfn), 0,
				"Not enough privilege",
				"Your user account seems to not have enough privilege for this action.",
				w, r,
			)
			return
		}

		err = ctx.DatabaseInterface.HardDeleteRepository(repo.Namespace, repo.Name)
		if err != nil {
			ctx.ReportRedirect(fmt.Sprintf("/repo/%s/setting", rfn), 0,
				"Internal error",
				fmt.Sprintf("Failed to delete repository: %s.", err.Error()),
				w, r,
			)
			return
		}
		redirectTarget := "/"
		if ctx.Config.UseNamespace { redirectTarget = fmt.Sprintf("/s/%s", ns.Name) }
		ctx.ReportRedirect(redirectTarget, 3, "Deleted.", "The specified repository is deleted.", w, r)
	}))

	http.HandleFunc("GET /repo/{repoName}/member", WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !CheckGlobalVisibleToUser(ctx, loginInfo) {
			switch ctx.Config.GlobalVisibility {
			case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
				FoundAt(w, "/maintenance-notice")
				return
			case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
				FoundAt(w, "/shutdown-notice")
				return
			case aegis.GLOBAL_VISIBILITY_PRIVATE:
				FoundAt(w, "/login")
				return
			}
		}
		rfn := r.PathValue("repoName")
		nsName, repoName, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if ctx.Config.UseNamespace && ns == nil {
			ctx.ReportNotFound(repo.Namespace, "Namespace", "depot", w, r)
			return
		}
		if repo == nil {
			ctx.ReportNotFound(repoName, "Repository", nsName, w, r)
			return
		}
		repoPath := fmt.Sprintf("/repo/%s", rfn)
		if !loginInfo.LoggedIn { FoundAt(w, repoPath); return }
		
		isRepoOwner := repo.Owner == loginInfo.UserName
		isNsOwner := ns.Owner == loginInfo.UserName
		loginInfo.IsOwner = isRepoOwner || isNsOwner
		isOwner := isRepoOwner || isNsOwner
		repoPriv := repo.AccessControlList.GetUserPrivilege(loginInfo.UserName)
		nsPriv := ns.ACL.GetUserPrivilege(loginInfo.UserName)
		isSettingMember := repoPriv.HasSettingPrivilege() || nsPriv.HasSettingPrivilege()
		if !loginInfo.IsAdmin && !isOwner && !isSettingMember {
			ctx.ReportRedirect(fmt.Sprintf("/repo/%s/setting", rfn), 0,
				"Not enough privilege",
				"Your user account seems to not have enough privilege for this action.",
				w, r,
			)
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
			RepoFullName: rfn,
			LoginInfo: loginInfo,
			ACL: page,
			PageInfo: pageInfo,
		}))
	}))
	
	http.HandleFunc("POST /repo/{repoName}/member", WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !CheckGlobalVisibleToUser(ctx, loginInfo) {
			switch ctx.Config.GlobalVisibility {
			case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
				FoundAt(w, "/maintenance-notice")
				return
			case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
				FoundAt(w, "/shutdown-notice")
				return
			case aegis.GLOBAL_VISIBILITY_PRIVATE:
				FoundAt(w, "/login")
				return
			}
		}
		rfn := r.PathValue("repoName")
		nsName, repoName, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if ctx.Config.UseNamespace && ns == nil {
			ctx.ReportNotFound(repo.Namespace, "Namespace", "depot", w, r)
			return
		}
		if repo == nil {
			ctx.ReportNotFound(repoName, "Repository", nsName, w, r)
			return
		}
		repoPath := fmt.Sprintf("/repo/%s", rfn)
		if !loginInfo.LoggedIn { FoundAt(w, repoPath); return }
		
		isRepoOwner := repo.Owner == loginInfo.UserName
		isNsOwner := ns.Owner == loginInfo.UserName
		loginInfo.IsOwner = isRepoOwner || isNsOwner
		isOwner := isRepoOwner || isNsOwner
		repoPriv := repo.AccessControlList.GetUserPrivilege(loginInfo.UserName)
		nsPriv := ns.ACL.GetUserPrivilege(loginInfo.UserName)
		hasAddMemberPriv := (nsPriv != nil && nsPriv.AddMember) || (repoPriv != nil && repoPriv.AddMember)
		if !loginInfo.IsAdmin && !isOwner && !hasAddMemberPriv {
			ctx.ReportRedirect(fmt.Sprintf("/repo/%s/setting", rfn), 0,
				"Not enough privilege",
				"Your user account seems to not have enough privilege for this action.",
				w, r,
			)
			return
		}

		err = r.ParseForm()
		if err != nil {
			ctx.ReportRedirect(fmt.Sprintf("/repo/%s/member", rfn), 0,
				"Invalid request",
				"Failed to parse request.",
				w, r,
			)
			return
		}
		username := strings.TrimSpace(r.Form.Get("username"))
		if len(username) <= 0 {
			ctx.ReportRedirect(fmt.Sprintf("/repo/%s/member", rfn), 0,
				"Invalid request",
				"User name cannot be empty.",
				w, r,
			)
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
			EditHooks: len(r.Form.Get("editHooks")) > 0,
		}
		err = ctx.DatabaseInterface.SetRepositoryACL(repo.Namespace, repo.Name, username, t)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		FoundAt(w, fmt.Sprintf("/repo/%s/member", rfn))
	}))

	http.HandleFunc("GET /repo/{repoName}/member/{userName}/edit", WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !CheckGlobalVisibleToUser(ctx, loginInfo) {
			switch ctx.Config.GlobalVisibility {
			case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
				FoundAt(w, "/maintenance-notice")
				return
			case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
				FoundAt(w, "/shutdown-notice")
				return
			case aegis.GLOBAL_VISIBILITY_PRIVATE:
				FoundAt(w, "/login")
				return
			}
		}
		rfn := r.PathValue("repoName")
		nsName, repoName, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if ctx.Config.UseNamespace && ns == nil {
			ctx.ReportNotFound(repo.Namespace, "Namespace", "depot", w, r)
			return
		}
		if repo == nil {
			ctx.ReportNotFound(repoName, "Repository", nsName, w, r)
			return
		}
		repoPath := fmt.Sprintf("/repo/%s", rfn)
		if !loginInfo.LoggedIn { FoundAt(w, repoPath); return }
		
		isRepoOwner := repo.Owner == loginInfo.UserName
		isNsOwner := ns.Owner == loginInfo.UserName
		loginInfo.IsOwner = isRepoOwner || isNsOwner
		isOwner := isRepoOwner || isNsOwner
		repoPriv := repo.AccessControlList.GetUserPrivilege(loginInfo.UserName)
		nsPriv := ns.ACL.GetUserPrivilege(loginInfo.UserName)
		hasEditMemberPriv := (nsPriv != nil && nsPriv.EditMember) || (repoPriv != nil && repoPriv.EditMember)
		if !loginInfo.IsAdmin && !isOwner && !hasEditMemberPriv {
			ctx.ReportRedirect(fmt.Sprintf("/repo/%s/member", rfn), 0,
				"Not enough privilege",
				"Your user account seems to not have enough privilege for this action.",
				w, r,
			)
			return
		}

		err = r.ParseForm()
		if err != nil {
			ctx.ReportRedirect(fmt.Sprintf("/repo/%s/member", rfn), 0,
				"Invalid request",
				"Failed to parse request.",
				w, r,
			)
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
			Username: targetUsername,
			ACLTuple: userPriv,
		}))
	}))

	http.HandleFunc("POST /repo/{repoName}/member/{userName}/edit", WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !CheckGlobalVisibleToUser(ctx, loginInfo) {
			switch ctx.Config.GlobalVisibility {
			case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
				FoundAt(w, "/maintenance-notice")
				return
			case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
				FoundAt(w, "/shutdown-notice")
				return
			case aegis.GLOBAL_VISIBILITY_PRIVATE:
				FoundAt(w, "/login")
				return
			}
		}
		rfn := r.PathValue("repoName")
		nsName, repoName, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if ctx.Config.UseNamespace && ns == nil {
			ctx.ReportNotFound(repo.Namespace, "Namespace", "depot", w, r)
			return
		}
		if repo == nil {
			ctx.ReportNotFound(repoName, "Repository", nsName, w, r)
			return
		}
		repoPath := fmt.Sprintf("/repo/%s", rfn)
		if !loginInfo.LoggedIn { FoundAt(w, repoPath); return }
		
		isRepoOwner := repo.Owner == loginInfo.UserName
		isNsOwner := ns.Owner == loginInfo.UserName
		loginInfo.IsOwner = isRepoOwner || isNsOwner
		isOwner := isRepoOwner || isNsOwner
		repoPriv := repo.AccessControlList.GetUserPrivilege(loginInfo.UserName)
		nsPriv := ns.ACL.GetUserPrivilege(loginInfo.UserName)
		hasEditMemberPriv := (nsPriv != nil && nsPriv.EditMember) || (repoPriv != nil && repoPriv.EditMember)
		if !loginInfo.IsAdmin && !isOwner && !hasEditMemberPriv {
			ctx.ReportRedirect(fmt.Sprintf("/repo/%s/member", rfn), 0,
				"Not enough privilege",
				"Your user account seems to not have enough privilege for this action.",
				w, r,
			)
			return
		}

		err = r.ParseForm()
		if err != nil {
			ctx.ReportRedirect(fmt.Sprintf("/repo/%s/member", rfn), 0,
				"Invalid request",
				"Failed to parse request.",
				w, r,
			)
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
			EditHooks: len(r.Form.Get("editHooks")) > 0,
		}
		err = ctx.DatabaseInterface.SetRepositoryACL(repo.Namespace, repo.Name, targetUsername, t)
		if err != nil {
			ctx.ReportRedirect(fmt.Sprintf("/repo/%s/member", rfn), 0,
				"Failed to update member privilege",
				fmt.Sprintf("Failed to update member privilege: %s. Please contact site owner.", err.Error()),
				w, r,
			)
			return
		}
		FoundAt(w, fmt.Sprintf("/repo/%s/member", rfn))
	}))

	
	http.HandleFunc("GET /repo/{repoName}/member/{username}/delete", WithLog(func(w http.ResponseWriter, r *http.Request) {
		rfn := r.PathValue("repoName")
		repoPath := fmt.Sprintf("/repo/%s", rfn)
		if ctx.Config.PlainMode { FoundAt(w, repoPath); return }
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !CheckGlobalVisibleToUser(ctx, loginInfo) {
			switch ctx.Config.GlobalVisibility {
			case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
				FoundAt(w, "/maintenance-notice")
				return
			case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
				FoundAt(w, "/shutdown-notice")
				return
			case aegis.GLOBAL_VISIBILITY_PRIVATE:
				FoundAt(w, "/login")
				return
			}
		}
		if !loginInfo.LoggedIn { FoundAt(w, repoPath); return }
		nsName, repoName, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		isRepoOwner := repo.Owner == loginInfo.UserName
		isNsOwner := ns.Owner == loginInfo.UserName
		loginInfo.IsOwner = isRepoOwner || isNsOwner
		nsPriv := ns.ACL.GetUserPrivilege(loginInfo.UserName)
		repoPriv := repo.AccessControlList.GetUserPrivilege(loginInfo.UserName)
		privSufficient := loginInfo.IsOwner || loginInfo.IsAdmin || (nsPriv != nil && nsPriv.DeleteMember) || (repoPriv != nil && repoPriv.DeleteMember)
		if !privSufficient {
			ctx.ReportRedirect(fmt.Sprintf("/s/%s/member", ns.Name), 0,
				"Not enough privilege",
				"You seem to not have not enough privilege for this action.",
				w, r,
			)
			return
		}
		targetUsername := r.PathValue("username")
		err = ctx.DatabaseInterface.SetRepositoryACL(nsName, repoName, targetUsername, nil)
		if err != nil {
			LogTemplateError(ctx.LoadTemplate("namespace-setting/_redirect-with-message").Execute(w, templates.RedirectWithMessageModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				Timeout: 3,
				RedirectUrl: fmt.Sprintf("/s/%s/member", nsName),
				MessageTitle: "Failed to delete member",
				MessageText: fmt.Sprintf("Error: %s", err.Error()),
			}))
			return
		}
		FoundAt(w, fmt.Sprintf("/s/%s/member", nsName))
	}))
}


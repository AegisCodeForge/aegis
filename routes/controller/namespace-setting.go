package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/model"
	auxfuncs "github.com/bctnry/aegis/pkg/auxfuncs"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindNamespaceSettingController(ctx *RouterContext) {
	http.HandleFunc("GET /s/{namespace}/setting", WithLog(func(w http.ResponseWriter, r *http.Request){
		namespaceName := r.PathValue("namespace")
		if !model.ValidNamespaceName(namespaceName) {
			ctx.ReportNotFound(namespaceName, "Repository", "Depot", w, r)
			return
		}
		namespacePath := fmt.Sprintf("/s/%s", namespaceName)
		// NOTE: we don't support editing namespace from web ui when in plain mode.
		if ctx.Config.PlainMode { FoundAt(w, namespacePath); return }
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
		if !loginInfo.LoggedIn { FoundAt(w, namespacePath); return }
		ns, err := ctx.DatabaseInterface.GetNamespaceByName(namespaceName)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		isOwner := ns.Owner == loginInfo.UserName
		priv := ns.ACL.GetUserPrivilege(loginInfo.UserName)
		isMember := priv != nil
		isSettingMember := isMember && priv.HasSettingPrivilege()
		if !loginInfo.IsAdmin && !isOwner && !isSettingMember {
			ctx.ReportRedirect(fmt.Sprintf("/s/%s", namespaceName), 0,
				"Not enough privilege",
				"Your user account seems to not have enough privilege for this action.",
				w, r,
			)
			return
		}
		loginInfo.IsOwner = isOwner
		LogTemplateError(ctx.LoadTemplate("namespace-setting/change-info").Execute(w, templates.NamespaceSettingTemplateModel{
			Namespace: ns,
			LoginInfo: loginInfo,
			Config: ctx.Config,
		}))
	}))

	http.HandleFunc("POST /s/{namespace}/setting", WithLog(func(w http.ResponseWriter, r *http.Request) {
		namespaceName := r.PathValue("namespace")
		if !model.ValidNamespaceName(namespaceName) {
			ctx.ReportNotFound(namespaceName, "Repository", "Depot", w, r)
			return
		}
		namespacePath := fmt.Sprintf("/s/%s", namespaceName)
		if ctx.Config.PlainMode { FoundAt(w, namespacePath); return }
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
		if !loginInfo.LoggedIn { FoundAt(w, namespacePath); return }
		ns, err := ctx.DatabaseInterface.GetNamespaceByName(namespaceName)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		isOwner := ns.Owner == loginInfo.UserName
		loginInfo.IsOwner = isOwner
		priv := ns.ACL.GetUserPrivilege(loginInfo.UserName)
		canEditInfo := priv != nil && priv.EditInfo
		if !loginInfo.IsAdmin && !isOwner && !canEditInfo {
			ctx.ReportRedirect(fmt.Sprintf("/s/%s", namespaceName), 0,
				"Not enough privilege",
				"Your user account seems to not have enough privilege for this action.",
				w, r,
			)
			return
		}
		loginInfo.IsSettingMember = loginInfo.IsAdmin || isOwner || priv.HasSettingPrivilege()
		err = r.ParseForm()
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		newOwner := strings.TrimSpace(r.Form.Get("owner"))
		if ns.Owner != newOwner {
			if !isOwner && !loginInfo.IsAdmin {
				ctx.ReportRedirect(fmt.Sprintf("/s/%s/setting", namespaceName), 0,
					"Not enough privilege",
					"You are neither owner nor an admin, thus cannot change this namespace's ownership.",
					w, r,
				)
				return
			}
			ns.Owner = newOwner
		}
		ns.Title = r.Form.Get("title")
		ns.Description = r.Form.Get("description")
		ns.Email = r.Form.Get("email")
		i, err := strconv.Atoi(r.Form.Get("status"))
		if err != nil {
			ctx.ReportRedirect(
				fmt.Sprintf("/s/%s/setting", namespaceName), 0,
				"Invalid Request",
				"Invalid status value. Please try again.",
				w, r,
			)
			return
		}
		ns.Status = model.AegisNamespaceStatus(i)
		err = ctx.DatabaseInterface.UpdateNamespaceInfo(namespaceName, ns)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		ctx.ReportRedirect(fmt.Sprintf("/s/%s/setting", namespaceName), 3,
			"Updated",
			"Namespace info updated.",
			w, r,
		)
	}))

	http.HandleFunc("GET /s/{namespace}/delete", WithLog(func(w http.ResponseWriter, r *http.Request) {
		namespaceName := r.PathValue("namespace")
		if !model.ValidNamespaceName(namespaceName) {
			ctx.ReportNotFound(namespaceName, "Repository", "Depot", w, r)
			return
		}
		namespacePath := fmt.Sprintf("/s/%s", namespaceName)
		if ctx.Config.PlainMode { FoundAt(w, namespacePath); return }
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
		if !loginInfo.LoggedIn { FoundAt(w, namespacePath); return }
		ns, err := ctx.DatabaseInterface.GetNamespaceByName(namespaceName)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !loginInfo.IsAdmin &&  ns.Owner != loginInfo.UserName {
			ctx.ReportRedirect(fmt.Sprintf("/s/%s/setting", namespaceName), 3,
				"Not enough privilege",
				"Your user account seems to not have enough privilege for this action.",
				w, r,
			)
		}
		err = ctx.DatabaseInterface.HardDeleteNamespaceByName(namespaceName)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		ctx.ReportRedirect("/", 3,
			"Deleted",
			"Namespace deleted..",
			w, r,
		)
	}))

	http.HandleFunc("GET /s/{namespace}/member", WithLog(func(w http.ResponseWriter, r *http.Request) {
		namespaceName := r.PathValue("namespace")
		if !model.ValidNamespaceName(namespaceName) {
			ctx.ReportNotFound(namespaceName, "Repository", "Depot", w, r)
			return
		}
		namespacePath := fmt.Sprintf("/s/%s", namespaceName)
		if ctx.Config.PlainMode { FoundAt(w, namespacePath); return }
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
		if !loginInfo.LoggedIn { FoundAt(w, namespacePath); return }
		ns, err := ctx.DatabaseInterface.GetNamespaceByName(namespaceName)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		isOwner := ns.Owner == loginInfo.UserName
		loginInfo.IsOwner = isOwner
		priv := ns.ACL.GetUserPrivilege(loginInfo.UserName)
		isMember := priv != nil
		isSettingMember := isMember && priv.HasSettingPrivilege()
		if !loginInfo.IsAdmin && !isOwner && !isSettingMember {
			ctx.ReportRedirect(fmt.Sprintf("/s/%s/setting", ns.Name), 0,
				"Not enough privilege",
				"Your user account seems to not have enough privilege for this action.",
				w, r,
			)
			return
		}
		var userList map[string]*model.ACLTuple
		if ns.ACL == nil {
			userList = nil
		} else {
			userList = ns.ACL.ACL
		}
		totalMemberCount := len(userList)
		pageInfo, err := GeneratePageInfo(r, totalMemberCount)
		// the reason we do this is the fact that maps in go does not
		// guarantee the order of keys when doing a range over them.
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
		LogTemplateError(ctx.LoadTemplate("namespace-setting/member-list").Execute(w, templates.NamespaceSettingMemberListTemplateModel{
			Namespace: ns,
			LoginInfo: loginInfo,
			Config: ctx.Config,
			ACL: userList,
			PageInfo: pageInfo,
		}))
	}))
	
	http.HandleFunc("POST /s/{namespace}/member", WithLog(func(w http.ResponseWriter, r *http.Request) {
		namespaceName := r.PathValue("namespace")
		if !model.ValidNamespaceName(namespaceName) {
			ctx.ReportNotFound(namespaceName, "Repository", "Depot", w, r)
			return
		}
		namespacePath := fmt.Sprintf("/s/%s", namespaceName)
		if ctx.Config.PlainMode { FoundAt(w, namespacePath); return }
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
		if !loginInfo.LoggedIn { FoundAt(w, namespacePath); return }
		ns, err := ctx.DatabaseInterface.GetNamespaceByName(namespaceName)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		isOwner := ns.Owner == loginInfo.UserName
		loginInfo.IsOwner = isOwner
		priv := ns.ACL.GetUserPrivilege(loginInfo.UserName)
		canAddMember := priv != nil && priv.AddMember
		if !loginInfo.IsAdmin && !isOwner && !canAddMember {
			ctx.ReportRedirect(fmt.Sprintf("/s/%s/setting", ns.Name), 0,
				"Not enough privilege",
				"Your user account seems to not have enough privilege for this action.",
				w, r,
			)
			return
		}
		err = r.ParseForm()
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		username := strings.TrimSpace(r.Form.Get("username"))
		if len(username) <= 0 {
			ctx.ReportRedirect(fmt.Sprintf("/s/%s/setting", ns.Name), 0,
				"Not enough privilege",
				"Your user account seems to not have enough privilege for this action.",
				w, r,
			)
			return
		}
		t := &model.ACLTuple{
			AddMember: len(r.Form.Get("addMember")) > 0,
			DeleteMember: len(r.Form.Get("deleteMember")) > 0,
			EditMember: len(r.Form.Get("editMember")) > 0,
			AddRepository: len(r.Form.Get("addRepo")) > 0,
			EditInfo: len(r.Form.Get("editInfo")) > 0,
			PushToRepository: len(r.Form.Get("pushToRepo")) > 0,
			ArchiveRepository: len(r.Form.Get("archiveRepo")) > 0,
			DeleteRepository: len(r.Form.Get("deleteRepo")) > 0,
			EditHooks: len(r.Form.Get("editHooks")) > 0,
		}
		err = ctx.DatabaseInterface.SetNamespaceACL(namespaceName, username, t)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		ctx.ReportRedirect(fmt.Sprintf("/s/%s/member", ns.Name), 3,
			"Updated",
			"Member list updated.",
			w, r,
		)
	}))

	http.HandleFunc("GET /s/{namespace}/member/{username}/delete", WithLog(func(w http.ResponseWriter, r *http.Request) {
		namespaceName := r.PathValue("namespace")
		if !model.ValidNamespaceName(namespaceName) {
			ctx.ReportNotFound(namespaceName, "Repository", "Depot", w, r)
			return
		}
		namespacePath := fmt.Sprintf("/s/%s", namespaceName)
		if ctx.Config.PlainMode { FoundAt(w, namespacePath); return }
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
		if !loginInfo.LoggedIn { FoundAt(w, namespacePath); return }
		ns, err := ctx.DatabaseInterface.GetNamespaceByName(namespaceName)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		isOwner := ns.Owner == loginInfo.UserName
		loginInfo.IsOwner = isOwner
		priv := ns.ACL.GetUserPrivilege(loginInfo.UserName)
		privSufficient := isOwner || loginInfo.IsAdmin || (priv != nil && priv.DeleteMember)
		if !privSufficient {
			ctx.ReportRedirect(fmt.Sprintf("/s/%s/member", ns.Name), 0,
				"Not enough privilege",
				"You seem to not have not enough privilege for this action.",
				w, r,
			)
			return
		}
		targetUsername := r.PathValue("username")
		err = ctx.DatabaseInterface.SetNamespaceACL(namespaceName, targetUsername, nil)
		if err != nil {
			LogTemplateError(ctx.LoadTemplate("namespace-setting/_redirect-with-message").Execute(w, templates.RedirectWithMessageModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				Timeout: 3,
				RedirectUrl: fmt.Sprintf("/s/%s/member", namespaceName),
				MessageTitle: "Failed to delete member",
				MessageText: fmt.Sprintf("Error: %s", err.Error()),
			}))
			return
		}
		FoundAt(w, fmt.Sprintf("/s/%s/member", namespaceName))
	}))
	
	http.HandleFunc("GET /s/{namespace}/member/{username}/edit", WithLog(func(w http.ResponseWriter, r *http.Request) {
		namespaceName := r.PathValue("namespace")
		if !model.ValidNamespaceName(namespaceName) {
			ctx.ReportNotFound(namespaceName, "Repository", "Depot", w, r)
			return
		}
		namespacePath := fmt.Sprintf("/s/%s", namespaceName)
		if ctx.Config.PlainMode { FoundAt(w, namespacePath); return }
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
		if !loginInfo.LoggedIn { FoundAt(w, namespacePath); return }
		ns, err := ctx.DatabaseInterface.GetNamespaceByName(namespaceName)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		isOwner := ns.Owner == loginInfo.UserName
		loginInfo.IsOwner = isOwner
		priv := ns.ACL.GetUserPrivilege(loginInfo.UserName)
		canEditMember := priv != nil && priv.EditMember
		if !loginInfo.IsAdmin && !isOwner && !canEditMember {
			ctx.ReportRedirect(fmt.Sprintf("/s/%s/member", namespaceName), 0,
				"Not enough privilege",
				"Your user account seems to not have enough privilege for this action.",
				w, r,
			)
			return
		}
		targetUsername := r.PathValue("username")
		userPriv := ns.ACL.GetUserPrivilege(targetUsername)
		LogTemplateError(ctx.LoadTemplate("namespace-setting/edit-member").Execute(w, templates.NamespaceSettingEditMemberTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			Namespace: ns,
			Username: targetUsername,
			ACLTuple: userPriv,
		}))
	}))

	http.HandleFunc("POST /s/{namespace}/member/{username}/edit", WithLog(func(w http.ResponseWriter, r *http.Request) {
		namespaceName := r.PathValue("namespace")
		if !model.ValidNamespaceName(namespaceName) {
			ctx.ReportNotFound(namespaceName, "Repository", "Depot", w, r)
			return
		}
		namespacePath := fmt.Sprintf("/s/%s", namespaceName)
		if ctx.Config.PlainMode { FoundAt(w, namespacePath); return }
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
		if !loginInfo.LoggedIn { FoundAt(w, namespacePath); return }
		ns, err := ctx.DatabaseInterface.GetNamespaceByName(namespaceName)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		isOwner := ns.Owner == loginInfo.UserName
		loginInfo.IsOwner = isOwner
		priv := ns.ACL.GetUserPrivilege(loginInfo.UserName)
		canEditMember := priv != nil && priv.EditMember
		if !loginInfo.IsAdmin && !isOwner && !canEditMember {
			ctx.ReportRedirect(fmt.Sprintf("/s/%s/member", namespaceName), 0,
				"Not enough privilege",
				"Your user account seems to not have enough privilege for this action.",
				w, r,
			)
			return
		}
		targetUsername := r.PathValue("username")
		err = r.ParseForm()
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		t := &model.ACLTuple{
			AddMember: len(r.Form.Get("addMember")) > 0,
			DeleteMember: len(r.Form.Get("deleteMember")) > 0,
			EditMember: len(r.Form.Get("editMember")) > 0,
			AddRepository: len(r.Form.Get("addRepo")) > 0,
			EditInfo: len(r.Form.Get("editInfo")) > 0,
			PushToRepository: len(r.Form.Get("pushToRepo")) > 0,
			ArchiveRepository: len(r.Form.Get("archiveRepo")) > 0,
			DeleteRepository: len(r.Form.Get("deleteRepo")) > 0,
			EditHooks: len(r.Form.Get("editHooks")) > 0,
		}
		err = ctx.DatabaseInterface.SetNamespaceACL(namespaceName, targetUsername, t)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		FoundAt(w, fmt.Sprintf("/s/%s/member", namespaceName))
	}))
}


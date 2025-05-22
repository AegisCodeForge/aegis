package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/bctnry/aegis/pkg/aegis/model"
	auxfuncs "github.com/bctnry/aegis/pkg/auxfuncs"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindNamespaceController(ctx *RouterContext) {
	if !ctx.Config.UseNamespace { return }
	http.HandleFunc("GET /s/{namespace}", WithLog(func(w http.ResponseWriter, r *http.Request) {
		namespaceName := r.PathValue("namespace")
		var ns *model.Namespace
		var ok bool
		var err error
		if ctx.Config.PlainMode {
			ns, ok = ctx.GitNamespaceList[namespaceName]
			if !ok {
				err = ctx.SyncAllNamespacePlain()
				if err != nil {
					ctx.ReportInternalError(err.Error(), w, r)
					return
				}
				ns, ok = ctx.GitNamespaceList[namespaceName]
				if !ok {
					ctx.ReportNotFound(namespaceName, "Namespace", ctx.Config.DepotName, w, r)
					return
				}
			}
		} else {
			ns, err = ctx.DatabaseInterface.GetNamespaceByName(namespaceName)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			s, err := ctx.DatabaseInterface.GetAllRepositoryFromNamespace(ns.Name)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			ns.RepositoryList = s
		}
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		v := ns.ACL.GetUserPrivilege(loginInfo.UserName).HasSettingPrivilege()
		loginInfo.IsSettingMember = v
		LogTemplateError(ctx.LoadTemplate("namespace").Execute(w, templates.NamespaceTemplateModel{
			Namespace: ns,
			LoginInfo: loginInfo,
			Config: ctx.Config,
		}))
	}))

	// ========================================
	// settings
	http.HandleFunc("GET /s/{namespace}/setting", WithLog(func(w http.ResponseWriter, r *http.Request){
		namespaceName := r.PathValue("namespace")
		namespacePath := fmt.Sprintf("/s/%s", namespaceName)
		// NOTE: we don't support editing namespace from web ui when in plain mode.
		if ctx.Config.PlainMode { FoundAt(w, namespacePath); return }
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
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
			ctx.ReportForbidden("Not enough permission", w, r)
			return
		}
		LogTemplateError(ctx.LoadTemplate("namespace-setting/change-info").Execute(w, templates.NamespaceSettingTemplateModel{
			Namespace: ns,
			LoginInfo: loginInfo,
			Config: ctx.Config,
		}))
	}))

	http.HandleFunc("POST /s/{namespace}/setting", WithLog(func(w http.ResponseWriter, r *http.Request) {
		namespaceName := r.PathValue("namespace")
		namespacePath := fmt.Sprintf("/s/%s", namespaceName)
		if ctx.Config.PlainMode { FoundAt(w, namespacePath); return }
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
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
			ctx.ReportForbidden("Not enough permission", w, r)
			return
		}
		loginInfo.IsSettingMember = isSettingMember
		err = r.ParseForm()
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		anythingChanged := false
		if ns.Title != strings.TrimSpace(r.Form.Get("title")) {
			anythingChanged = true
		}
		ns.Title = r.Form.Get("title")
		if ns.Description != strings.TrimSpace(r.Form.Get("description")) {
			anythingChanged = true
		}
		ns.Description = r.Form.Get("description")
		if ns.Email != strings.TrimSpace(r.Form.Get("email")) {
			anythingChanged = true
		}
		ns.Email = r.Form.Get("email")
		i, err := strconv.Atoi(r.Form.Get("status"))
		if err != nil {
			LogTemplateError(ctx.LoadTemplate("namespace-setting/change-info").Execute(w, &templates.NamespaceSettingTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				Namespace: ns,
				ErrorMsg: struct{Type string; Message string}{
					Type: r.Form.Get("type"),
					Message: fmt.Sprintf("Invalid status value. Please try again.", err.Error()),
				},
			}))
			return
		}
		if ns.Status != model.AegisNamespaceStatus(i) {
			anythingChanged = true
		}
		ns.Status = model.AegisNamespaceStatus(i)
		userPrivilege := ns.ACL.GetUserPrivilege(loginInfo.UserName)
		if anythingChanged && (userPrivilege == nil || !userPrivilege.EditInfo) {
			LogTemplateError(ctx.LoadTemplate("namespace-setting/change-info").Execute(w, &templates.NamespaceSettingTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				Namespace: ns,
				ErrorMsg: struct{Type string; Message string}{
					Type: "",
					Message: "Not enough privilege.",
				},
			}))
			return
		}
		err = ctx.DatabaseInterface.UpdateNamespaceInfo(namespaceName, ns)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		LogTemplateError(ctx.LoadTemplate("namespace-setting/change-info").Execute(w, &templates.NamespaceSettingTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			Namespace: ns,
			ErrorMsg: struct{Type string; Message string}{
				Type: r.Form.Get("type"),
				Message: fmt.Sprintf("Updated."),
			},
		}))
	}))

	http.HandleFunc("GET /s/{namespace}/delete", WithLog(func(w http.ResponseWriter, r *http.Request) {
		namespaceName := r.PathValue("namespace")
		namespacePath := fmt.Sprintf("/s/%s", namespaceName)
		if ctx.Config.PlainMode { FoundAt(w, namespacePath); return }
		userInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !userInfo.LoggedIn { FoundAt(w, namespacePath); return }
		ns, err := ctx.DatabaseInterface.GetNamespaceByName(namespaceName)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if ns.Owner != userInfo.UserName {
			ctx.ReportForbidden("Not owner", w, r)
			return
		}
		err = ctx.DatabaseInterface.HardDeleteNamespaceByName(namespaceName)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		FoundAt(w, "/")
	}))

	http.HandleFunc("GET /s/{namespace}/member", WithLog(func(w http.ResponseWriter, r *http.Request) {
		namespaceName := r.PathValue("namespace")
		namespacePath := fmt.Sprintf("/s/%s", namespaceName)
		if ctx.Config.PlainMode { FoundAt(w, namespacePath); return }
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
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
			ctx.ReportForbidden("Not enough privilege", w, r)
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
		namespacePath := fmt.Sprintf("/s/%s", namespaceName)
		if ctx.Config.PlainMode { FoundAt(w, namespacePath); return }
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
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
			ctx.ReportForbidden("Not enough privilege", w, r)
			return
		}
		err = r.ParseForm()
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		username := strings.TrimSpace(r.Form.Get("username"))
		if len(username) <= 0 {		
			LogTemplateError(ctx.LoadTemplate("namespace-setting/_redirect-with-message").Execute(w, templates.RedirectWithMessageModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				Timeout: 3,
				RedirectUrl: fmt.Sprintf("/s/%s/member", namespaceName),
				MessageTitle: "Invalid request",
				MessageText: "User name cannot be empty.",
			}))
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
		}
		err = ctx.DatabaseInterface.SetNamespaceACL(namespaceName, username, t)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		FoundAt(w, fmt.Sprintf("/s/%s/member", namespaceName))
	}))

	http.HandleFunc("GET /s/{namespace}/member/{username}/delete", WithLog(func(w http.ResponseWriter, r *http.Request) {
		namespaceName := r.PathValue("namespace")
		namespacePath := fmt.Sprintf("/s/%s", namespaceName)
		if ctx.Config.PlainMode { FoundAt(w, namespacePath); return }
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
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
			ctx.ReportForbidden("Not enough privilege", w, r)
			return
		}
		fmt.Println("user priv ", priv)
		if !priv.DeleteMember {
			LogTemplateError(ctx.LoadTemplate("namespace-setting/_redirect-with-message").Execute(w, templates.RedirectWithMessageModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				Timeout: 3,
				RedirectUrl: fmt.Sprintf("/s/%s/member", namespaceName),
				MessageTitle: "Not enough privilege",
				MessageText: "You don't have enough privilege in this namespace.",
			}))
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
		namespacePath := fmt.Sprintf("/s/%s", namespaceName)
		if ctx.Config.PlainMode { FoundAt(w, namespacePath); return }
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
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
		if !loginInfo.IsAdmin && !isOwner {
			if !isSettingMember || !priv.EditMember {
				LogTemplateError(ctx.LoadTemplate("namespace-setting/_redirect-with-message").Execute(w, templates.RedirectWithMessageModel{
					Config: ctx.Config,
					LoginInfo: loginInfo,
					Timeout: 3,
					RedirectUrl: fmt.Sprintf("/s/%s/member", namespaceName),
					MessageTitle: "Not enough privilege",
					MessageText: "You don't have enough privilege in this namespace.",
				}))
				return
			}
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
		namespacePath := fmt.Sprintf("/s/%s", namespaceName)
		if ctx.Config.PlainMode { FoundAt(w, namespacePath); return }
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
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
		if !loginInfo.IsAdmin && !isOwner {
			if !isSettingMember || !priv.EditMember {
				LogTemplateError(ctx.LoadTemplate("namespace-setting/_redirect-with-message").Execute(w, templates.RedirectWithMessageModel{
					Config: ctx.Config,
					LoginInfo: loginInfo,
					Timeout: 3,
					RedirectUrl: fmt.Sprintf("/s/%s/member", namespaceName),
					MessageTitle: "Not enough privilege",
					MessageText: "You don't have enough privilege in this namespace.",
				}))
				return
			}
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
		}
		err = ctx.DatabaseInterface.SetNamespaceACL(namespaceName, targetUsername, t)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		FoundAt(w, fmt.Sprintf("/s/%s/member", namespaceName))
	}))
}


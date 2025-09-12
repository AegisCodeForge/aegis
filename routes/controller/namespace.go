package controller

import (
	"fmt"
	"net/http"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/db"
	"github.com/bctnry/aegis/pkg/aegis/model"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindNamespaceController(ctx *RouterContext) {
	if !ctx.Config.UseNamespace { return }
	http.HandleFunc("GET /s/{namespace}", WithLog(func(w http.ResponseWriter, r *http.Request) {
		namespaceName := r.PathValue("namespace")
		if !model.ValidNamespaceName(namespaceName) {
			ctx.ReportNotFound(namespaceName, "Repository", "Depot", w, r)
			return
		}
		var ns *model.Namespace
		var ok bool
		var err error
		var loginInfo *templates.LoginInfoModel
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
			LogTemplateError(ctx.LoadTemplate("namespace").Execute(w, templates.NamespaceTemplateModel{
				Namespace: ns,
				Config: ctx.Config,
			}))
		} else {
			ns, err = ctx.DatabaseInterface.GetNamespaceByName(namespaceName)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			loginInfo.IsOwner = ns.Owner == loginInfo.UserName
			v := ns.ACL.GetUserPrivilege(loginInfo.UserName).HasSettingPrivilege()
			loginInfo.IsSettingMember = v
			s, err := ctx.DatabaseInterface.GetAllVisibleRepositoryFromNamespace(loginInfo.UserName, ns.Name)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			ns.RepositoryList = make(map[string]*model.Repository, 0)
			for _, k := range s {
				ns.RepositoryList[k.Name] = k
			}
			LogTemplateError(ctx.LoadTemplate("namespace").Execute(w, templates.NamespaceTemplateModel{
				Namespace: ns,
				LoginInfo: loginInfo,
				Config: ctx.Config,
			}))
		}
	}))
	
	http.HandleFunc("GET /s/{namespace}/new-repo", UseMiddleware(
		[]Middleware{Logged, LoginRequired, GlobalVisibility, ErrorGuard}, ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			nsName := r.PathValue("namespace")
			if !model.ValidNamespaceName(nsName) {
				rc.ReportNotFound(nsName, "Namespace", "Depot", w, r)
				return
			}
			ns, err := rc.DatabaseInterface.GetNamespaceByName(nsName)
			if err == db.ErrEntityNotFound {
				rc.ReportNotFound(nsName, "Namespace", "Depot", w, r)
				return
			}
			if err != nil {
				rc.ReportInternalError(fmt.Sprintf("Failed to retrieve namespace: %s", err), w, r)
				return
			}
			rc.LoginInfo.IsOwner = ns.Owner == rc.LoginInfo.UserName
			priv := ns.ACL.GetUserPrivilege(rc.LoginInfo.UserName)
			if !rc.LoginInfo.IsAdmin && !rc.LoginInfo.IsOwner && priv == nil {
				rc.ReportRedirect(fmt.Sprintf("/s/%s", nsName), 5, "Not Member", "You need to be a member of this namespace to create a repository under this namespace.", w, r)
				return
			}
			LogTemplateError(rc.LoadTemplate("new/repository").Execute(w, &templates.NewRepositoryTemplateModel{
				Config: rc.Config,
				LoginInfo: rc.LoginInfo,
				PredefinedNamespace: ns.Name,
			}))
		},
	))
	
	http.HandleFunc("POST /s/{namespace}/new-repo", UseMiddleware(
		[]Middleware{Logged, LoginRequired, GlobalVisibility, ErrorGuard}, ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			nsName := r.PathValue("namespace")
			if !model.ValidNamespaceName(nsName) {
				rc.ReportNotFound(nsName, "Namespace", "Depot", w, r)
				return
			}
			err := r.ParseForm()
			if err != nil {
				rc.ReportNormalError("Invalid request", w, r)
				return
			}
			ns, err := rc.DatabaseInterface.GetNamespaceByName(nsName)
			if err == db.ErrEntityNotFound {
				rc.ReportNotFound(nsName, "Namespace", "Depot", w, r)
				return
			}
			if err != nil {
				rc.ReportInternalError(fmt.Sprintf("Failed to retrieve namespace: %s", err), w, r)
				return
			}
			rc.LoginInfo.IsOwner = ns.Owner == rc.LoginInfo.UserName
			priv := ns.ACL.GetUserPrivilege(rc.LoginInfo.UserName)
			if !rc.LoginInfo.IsAdmin && !rc.LoginInfo.IsOwner && priv == nil {
				rc.ReportRedirect(fmt.Sprintf("/s/%s", nsName), 5, "Not Member", "You need to be a member of this namespace to create a repository under this namespace.", w, r)
				return
			}
			name := r.Form.Get("name")
			repo, err := rc.DatabaseInterface.CreateRepository(nsName, name, model.REPO_TYPE_GIT, rc.LoginInfo.UserName)
			if err != nil {
				rc.ReportInternalError(fmt.Sprintf("Failed to create repository: %s", err), w, r)
				return
			}
			rc.ReportRedirect(fmt.Sprintf("/repo/%s", repo.FullName()), 5, "Repository Created", fmt.Sprintf("A new repository named %s has been created under namespace %s.", name, nsName), w, r)
		},
	))
}


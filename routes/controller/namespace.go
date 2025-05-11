package controller

import (
	"fmt"
	"net/http"

	"github.com/bctnry/gitus/pkg/gitus/model"
	. "github.com/bctnry/gitus/routes"
	"github.com/bctnry/gitus/templates"
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
				err = ctx.SyncAllNamespace()
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
		userInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		LogTemplateError(ctx.LoadTemplate("namespace").Execute(w, templates.NamespaceTemplateModel{
			DepotName: ctx.Config.DepotName,
			Namespace: ns,
			LoginInfo: userInfo,
			Config: ctx.Config,
		}))
	}))

	http.HandleFunc("GET /s/{namespace}/setting", WithLog(func(w http.ResponseWriter, r *http.Request){
		namespaceName := r.PathValue("namespace")
		namespacePath := fmt.Sprintf("/s/%s", namespaceName)
		// NOTE: we don't support editing namespace from web ui when in plain mode.
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
		LogTemplateError(ctx.LoadTemplate("namespace-setting").Execute(w, templates.NamespaceSettingTemplateModel{
			DepotName: ctx.Config.DepotName,
			Namespace: ns,
			LoginInfo: userInfo,
			Config: ctx.Config,
		}))
	}))

	http.HandleFunc("POST /s/{namespace}/setting", WithLog(func(w http.ResponseWriter, r *http.Request) {
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
		err = r.ParseForm()
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		ns.Title = r.Form.Get("title")
		ns.Description = r.Form.Get("description")
		ns.Email = r.Form.Get("email")
		err = ctx.DatabaseInterface.UpdateNamespaceInfo(namespaceName, ns)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		FoundAt(w, namespacePath)
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
		err = ctx.DatabaseInterface.UpdateNamespaceStatus(namespaceName, model.NAMESPACE_DELETED)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		FoundAt(w, "/")
	}))
}

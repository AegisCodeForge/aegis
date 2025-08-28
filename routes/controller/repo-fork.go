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

func bindRepositoryForkController(ctx *RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/fork", WithLog(func(w http.ResponseWriter, r *http.Request) {
		rfn := r.PathValue("repoName")
		if !model.ValidRepositoryName(rfn) {
			ctx.ReportNotFound(rfn, "Repository", "Namespace", w, r)
			return
		}
		if ctx.Config.PlainMode {
			switch ctx.Config.GlobalVisibility {
			case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
				FoundAt(w, "/maintenance-notice")
				return
			case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
				FoundAt(w, "/shutdown-notice")
				return
			case aegis.GLOBAL_VISIBILITY_PRIVATE:
				FoundAt(w, "/private-notice")
				return
			}
			FoundAt(w, fmt.Sprintf("/repo/%s", rfn))
			return
		}
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
		_, _, _, s, err := ctx.ResolveRepositoryFullName(rfn)
		if err == ErrNotFound {
			ctx.ReportNotFound(rfn, "Repository", "Depot", w, r)
			return
		}
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !loginInfo.LoggedIn {
			ctx.ReportRedirect(fmt.Sprintf("/repo/%s", rfn), 5, "Not Logged In", "You must log in before forking this repository.", w, r)
			return
		}
		fr, err := ctx.DatabaseInterface.GetForkRepositoryOfUser(loginInfo.UserName, s.Namespace, s.Name)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if (!ctx.Config.UseNamespace) {
			if len(fr) > 0 {
				FoundAt(w, fmt.Sprintf("/repo/%s", fr[0].FullName()))
				return
			}
		}
		var l map[string]*model.Namespace = nil
		if (ctx.Config.UseNamespace) {
			l, err = ctx.DatabaseInterface.GetAllComprisingNamespace(loginInfo.UserName)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			for k := range l {
				if !loginInfo.IsAdmin && l[k].Owner != loginInfo.UserName {
					acl := l[k].ACL.GetUserPrivilege(loginInfo.UserName)
					if acl == nil || !acl.AddRepository {
						delete(l, k)
					}
				}
			}
		}
		LogTemplateError(ctx.LoadTemplate("fork").Execute(w, templates.ForkTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			SourceRepository: s,
			ForkedRepoList: fr,
			NamespaceList: l,
		}))
	}))
	
	http.HandleFunc("POST /repo/{repoName}/fork", WithLog(func(w http.ResponseWriter, r *http.Request) {
		rfn := r.PathValue("repoName")
		if !model.ValidRepositoryName(rfn) {
			ctx.ReportNotFound(rfn, "Repository", "Namespace", w, r)
			return
		}
		if ctx.Config.PlainMode {
			switch ctx.Config.GlobalVisibility {
			case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
				FoundAt(w, "/maintenance-notice")
				return
			case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
				FoundAt(w, "/shutdown-notice")
				return
			case aegis.GLOBAL_VISIBILITY_PRIVATE:
				FoundAt(w, "/private-notice")
				return
			}
			FoundAt(w, fmt.Sprintf("/repo/%s", rfn))
			return
		}
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
		originNs, originName, _, _, err := ctx.ResolveRepositoryFullName(rfn)
		if err == ErrNotFound {
			ctx.ReportNotFound(rfn, "Repository", "Depot", w, r)
			return
		}
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		err = r.ParseForm()
		if err != nil {
			ctx.ReportNormalError("Invalid request", w, r)
			return
		}
		namespace := r.Form.Get("namespace")
		name := r.Form.Get("name")
		if !model.ValidStrictRepositoryName(name){
			ctx.ReportRedirect(fmt.Sprintf("/repo/%s/fork", rfn), 0, "Invalid Repository Name", "Repository name must consists of only upper & lowercase letters (a-z, A-Z), 0-9, underscore and hyphen.", w, r)
			return
		}
		rp, err := ctx.DatabaseInterface.SetUpCloneRepository(originNs, originName, namespace, name, loginInfo.UserName)
		if err == db.ErrEntityAlreadyExists {
			ctx.ReportRedirect(fmt.Sprintf("/repo/%s/fork", rfn), 0, "Already Exists", fmt.Sprintf("The repository %s:%s already exists. Please choose a different name or namespace.", namespace, name), w, r)
			return
		}
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		FoundAt(w, fmt.Sprintf("/repo/%s", rp.FullName()))
	}))
}


package controller

import (
	"fmt"
	"net/http"

	"github.com/bctnry/aegis/pkg/aegis/db"
	"github.com/bctnry/aegis/pkg/aegis/model"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindRepositoryForkController(ctx *RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/fork", WithLog(func(w http.ResponseWriter, r *http.Request) {
		rfn := r.PathValue("repoName")
		if ctx.Config.PlainMode {
			FoundAt(w, fmt.Sprintf("/repo/%s", rfn))
			return
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
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
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
			if fr != nil && len(fr) > 0 {
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
		if ctx.Config.PlainMode {
			FoundAt(w, fmt.Sprintf("/repo/%s", rfn))
			return
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
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		/*
		fr, err := ctx.DatabaseInterface.GetForkRepositoryOfUser(loginInfo.UserName, s.Namespace, s.Name)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if fr != nil {
			FoundAt(w, fmt.Sprintf("/repo/%s", fr.FullName()))
			return
		}
		*/
		err = r.ParseForm()
		if err != nil {
			ctx.ReportNormalError("Invalid request", w, r)
			return
		}
		namespace := r.Form.Get("namespace")
		name := r.Form.Get("name")
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


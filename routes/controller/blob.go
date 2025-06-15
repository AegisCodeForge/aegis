package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/pkg/gitlib"
	"github.com/bctnry/aegis/routes"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)


func bindBlobController(ctx *RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/blob/{blobId}/", WithLog(func(w http.ResponseWriter, r *http.Request) {
		rfn := r.PathValue("repoName")
		_, _, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		if err == routes.ErrNotFound {
			ctx.ReportNotFound(rfn, "Repository", "Depot", w, r)
			return
		}
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		var loginInfo *templates.LoginInfoModel = nil
		if !ctx.Config.PlainMode {
			loginInfo, err = GenerateLoginInfoModel(ctx, r)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			loginInfo.IsOwner = (repo.Owner == loginInfo.UserName) || (ns.Owner == loginInfo.UserName)
		}
		if !ctx.Config.PlainMode && repo.Status == model.REPO_NORMAL_PRIVATE {
			t := repo.AccessControlList.GetUserPrivilege(loginInfo.UserName)
			if t == nil {
				t = ns.ACL.GetUserPrivilege(loginInfo.UserName)
			}
			if t == nil {
				LogTemplateError(ctx.LoadTemplate("error").Execute(w, templates.ErrorTemplateModel{
					LoginInfo: loginInfo,
					ErrorCode: 403,
					ErrorMessage: "Not enough privilege.",
				}))
				return
			}
		}
		
		blobId := r.PathValue("blobId")

		repoHeaderInfo := GenerateRepoHeader("blob", blobId)

		gobj, err := repo.Repository.ReadObject(blobId)
		if err != nil {
			ctx.ReportObjectReadFailure(blobId, err.Error(), w, r)
			return
		}
		if gobj.Type() != gitlib.BLOB {
			ctx.ReportObjectTypeMismatch(gobj.ObjectId(), "BLOB", gobj.Type().String(), w, r)
			return
		}

		// NOTE THAT we don't know the path with blob so we can't predict what kind of
		// file it is unless we look at its content and hope that we can make a good
		// assumption without calculating too much. the current behaviour is thus
		// intentional and we shall come back to this in the future...
		templateType := "file-text"
		bobj := gobj.(*gitlib.BlobObject)
		if r.URL.Query().Has("raw") || r.URL.Query().Has("snapshot") {
			w.Write(bobj.Data)
			return
		}
		str := string(bobj.Data)
		coloredStr, err := colorSyntax("", str)
		if err == nil { str = coloredStr }
		permaLink := fmt.Sprintf("/repo/%s/blob/%s", rfn, blobId)


		LogTemplateError(ctx.LoadTemplate(templateType).Execute(w, templates.FileTemplateModel{
			RepoHeaderInfo: *repoHeaderInfo,
			File: templates.BlobTextTemplateModel{
				FileLineCount: strings.Count(str, "\n"),
				FileContent: str,
			},
			Repository: repo,
			PermaLink: permaLink,
			TreePath: nil,
			CommitInfo: nil,
			TagInfo: nil,
			LoginInfo: loginInfo,
		}))

	}))
}



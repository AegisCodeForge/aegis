package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/pkg/gitlib"
	"github.com/bctnry/aegis/routes"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func handleTreeSnapshotRequest(repo *gitlib.LocalGitRepository, treeId string, obj gitlib.GitObject, w http.ResponseWriter) {
	filename := fmt.Sprintf(
		"%s-%s-tree-%s",
		repo.Namespace, repo.Name, treeId,
	)
	responseWithTreeZip(repo, obj, filename, w)
}

func bindTreeHandler(ctx *RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/tree/{treeId}/{treePath...}", WithLog(func(w http.ResponseWriter, r *http.Request) {
		rfn := r.PathValue("repoName")
		var err error
		var loginInfo *templates.LoginInfoModel = nil
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
		_, repoName, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		if err == routes.ErrNotFound {
			ctx.ReportNotFound(rfn, "Repository", "Depot", w, r)
			return
		}
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !ctx.Config.PlainMode {
			loginInfo.IsOwner = repo.Owner == loginInfo.UserName || ns.Owner == loginInfo.UserName
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
		
		treeId := r.PathValue("treeId")
		treePath := r.PathValue("treePath")
		repo, ok := ctx.GitRepositoryList[repoName]
		if !ok {
			ctx.ReportNotFound(repoName, "Repository", "depot", w, r)
			return
		}
		repoHeaderInfo := templates.RepoHeaderTemplateModel{
			TypeStr: "tree",
			NodeName: treeId,
		}

		gobj, err := repo.Repository.ReadObject(treeId)
		if err != nil {
			ctx.ReportObjectReadFailure(treeId, err.Error(), w, r)
			return
		}
		if gobj.Type() != gitlib.TREE {
			ctx.ReportObjectTypeMismatch(gobj.ObjectId(), "TREE", gobj.Type().String(), w, r)
			return
		}

		rootFullName := fmt.Sprintf("%s@%s:%s", rfn, "tree", treeId)
		rootPath := fmt.Sprintf("/repo/%s/%s/%s", rfn, "tree", treeId)
		permaLink := fmt.Sprintf("/repo/%s/tree/%s/%s", rfn, treeId, treePath)
		tobj := gobj.(*gitlib.TreeObject)
		var commitInfo *templates.CommitInfoTemplateModel = nil
		target, err := repo.Repository.ResolveTreePath(tobj, treePath)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
		}

		if target.Type() == gitlib.TREE && len(treePath) > 0 && !strings.HasSuffix(treePath, "/") {
			FoundAt(w, fmt.Sprintf("%s/%s/", rootPath, treePath))
			return
		}

		isSnapshotRequest :=  r.URL.Query().Has("snapshot")
		if isSnapshotRequest {
			handleTreeSnapshotRequest(repo.Repository, treeId, target, w)
			return
		}
		
		tp1 := make([]string, 0)
		treePathSegmentList := make([]struct{Name string;RelPath string}, 0)
		for item := range strings.SplitSeq(treePath, "/") {
			if len(item) <= 0 { continue }
			tp1 = append(tp1, item)
			treePathSegmentList = append(treePathSegmentList, struct{
				Name string; RelPath string
			}{
				Name: item, RelPath: strings.Join(tp1, "/"),
			})
		}
		treePathModelValue := &templates.TreePathTemplateModel{
			RootFullName: rootFullName,
			RootPath: rootPath,
			TreePath: treePath,
			TreePathSegmentList: treePathSegmentList,
		}
		LogTemplateError(ctx.LoadTemplate("tree").Execute(w, templates.TreeTemplateModel{
			Repository: repo,
			RepoHeaderInfo: repoHeaderInfo,
			TreeFileList: &templates.TreeFileListTemplateModel{
				ShouldHaveParentLink: len(treePath) > 0,
				RootPath: rootPath,
				TreePath: treePath,
				FileList: target.(*gitlib.TreeObject).ObjectList,
			},
			PermaLink: permaLink,
			TreePath: treePathModelValue,
			CommitInfo: commitInfo,
			TagInfo: nil,
			LoginInfo: loginInfo,
			Config: ctx.Config,
		}))
	}))
}


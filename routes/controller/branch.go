package controller

import (
	"fmt"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/pkg/gitlib"
	"github.com/bctnry/aegis/routes"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func handleBranchSnapshotRequest(repo *gitlib.LocalGitRepository, branchName string, obj gitlib.GitObject, w http.ResponseWriter) {
	filename := fmt.Sprintf(
		"%s-%s-branch-%s",
		repo.Namespace, repo.Name, branchName,
	)
	responseWithTreeZip(repo, obj, filename, w)
}

func bindBranchController(ctx *RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/branch/{branchName}/{treePath...}", WithLog(func(w http.ResponseWriter, r *http.Request) {
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
		rfn := r.PathValue("repoName")
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
		
		branchName := r.PathValue("branchName")
		repoHeaderInfo := GenerateRepoHeader("branch", branchName)
		
		treePath := r.PathValue("treePath")

		err = repo.Repository.SyncAllBranchList()
		if err != nil {
			ctx.ReportInternalError(
				fmt.Sprintf(
					"Cannot sync branch list for %s: %s",
					repoName,
					err.Error(),
				), w, r,
			)
			return
		}
		br, ok := repo.Repository.BranchIndex[branchName]
		if !ok {
			ctx.ReportNotFound(branchName, "Branch", repoName, w, r)
			return
		}
		gobj, err := repo.Repository.ReadObject(br.HeadId)
		if err != nil {
			ctx.ReportObjectReadFailure(br.HeadId, err.Error(), w, r)
			return
		}
		if gobj.Type() != gitlib.COMMIT {
			ctx.ReportObjectTypeMismatch(gobj.ObjectId(), "COMMIT", gobj.Type().String(), w, r)
			return
		}

		cobj := gobj.(*gitlib.CommitObject)
		commitInfo := &templates.CommitInfoTemplateModel{
			RootPath: fmt.Sprintf("/repo/%s", rfn),
			Commit: cobj,
		}
		gobj, err = repo.Repository.ReadObject(cobj.TreeObjId)
		if err != nil { ctx.ReportInternalError(err.Error(), w, r) }
		target, err := repo.Repository.ResolveTreePath(gobj.(*gitlib.TreeObject), treePath)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}

		// if it's a query for snapshot of a tree we directly output
		// the tree object as a .zip file
		isSnapshotRequest :=  r.URL.Query().Has("snapshot")
		if isSnapshotRequest {
			if target.Type() == gitlib.BLOB {
				mime := mime.TypeByExtension(path.Ext(treePath))
				if len(mime) <= 0 { mime = "application/octet-stream" }
				w.Header().Add("Content-Type", mime)
				w.Write((target.(*gitlib.BlobObject)).Data)
				return
			} else {
				handleBranchSnapshotRequest(repo.Repository, branchName, target, w)
				return
			}
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
		rootFullName := fmt.Sprintf("%s@%s:%s", rfn, "branch", branchName)
		rootPath := fmt.Sprintf("/repo/%s/%s/%s", rfn, "branch", branchName)
		treePathModelValue := &templates.TreePathTemplateModel{
			RootFullName: rootFullName,
			RootPath: rootPath,
			TreePath: treePath,
			TreePathSegmentList: treePathSegmentList,
		}
		permaLink := fmt.Sprintf("/repo/%s/commit/%s/%s", rfn, cobj.Id, treePath)
		
		switch target.Type() {
		case gitlib.TREE:
			if len(treePath) > 0 && !strings.HasSuffix(treePath, "/") {
				FoundAt(w, fmt.Sprintf("%s/%s/", rootPath, treePath))
				return
			}
			// NOTE: this is intentional. by the time we've reached here
			// `treePath` would end with a slash `/`, and the first `path.Dir`
			// call would only remove that slash, whose result is not the path
			// of the parent directory.
			var parentTreeFileList *templates.TreeFileListTemplateModel = nil
			if treePath != "" {
				dirPath := path.Dir(path.Dir(treePath)) + "/"
				dirObj, err := repo.Repository.ResolveTreePath(gobj.(*gitlib.TreeObject), dirPath)
				if err != nil {
					ctx.ReportInternalError(err.Error(), w, r)
					return
				}
				parentTreeFileList = &templates.TreeFileListTemplateModel{
					ShouldHaveParentLink: len(treePath) > 0,
					RootPath: rootPath,
					TreePath: dirPath,
					FileList: dirObj.(*gitlib.TreeObject).ObjectList,
				}
			}
			LogTemplateError(ctx.LoadTemplate("tree").Execute(w, templates.TreeTemplateModel{
				Repository: repo,
				RepoHeaderInfo: *repoHeaderInfo,
				TreeFileList: &templates.TreeFileListTemplateModel{
					ShouldHaveParentLink: len(treePath) > 0,
					RootPath: rootPath,
					TreePath: treePath,
					FileList: target.(*gitlib.TreeObject).ObjectList,
				},
				ParentTreeFileList: parentTreeFileList,
				PermaLink: permaLink,
				TreePath: treePathModelValue,
				CommitInfo: commitInfo,
				TagInfo: nil,
				LoginInfo: loginInfo,
				Config: ctx.Config,
			}))
		case gitlib.BLOB:
			dirPath := path.Dir(treePath) + "/"
			dirObj, err := repo.Repository.ResolveTreePath(gobj.(*gitlib.TreeObject), dirPath)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			mime := mime.TypeByExtension(path.Ext(treePath))
			if len(mime) <= 0 { mime = "application/octet-stream" }
			templateType := "file-text"
			if strings.HasPrefix(mime, "image/") {
				templateType = "file-image"
			}
			bobj := target.(*gitlib.BlobObject)
			if r.URL.Query().Has("raw") {
				w.Header().Add("Content-Type", mime)
				w.Write(bobj.Data)
				return
			}
			str := string(bobj.Data)
			filename := path.Base(treePath)
			coloredStr, err := colorSyntax(filename, str)
			if err == nil { str = coloredStr }
			LogTemplateError(ctx.LoadTemplate(templateType).Execute(w, templates.FileTemplateModel{
				Repository: repo,
				RepoHeaderInfo: *repoHeaderInfo,
				File: templates.BlobTextTemplateModel{
					FileLineCount: strings.Count(str, "\n"),
					FileContent: str,
				},
				PermaLink: permaLink,
				TreeFileList: &templates.TreeFileListTemplateModel{
					ShouldHaveParentLink: len(treePath) > 0,
					RootPath: rootPath,
					TreePath: dirPath,
					FileList: dirObj.(*gitlib.TreeObject).ObjectList,
				},
				TreePath: treePathModelValue,
				CommitInfo: commitInfo,
				TagInfo: nil,
				LoginInfo: loginInfo,
				Config: ctx.Config,
			}))
		default:
			ctx.ReportInternalError("", w, r)
		}

	}))
}



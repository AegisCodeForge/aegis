package controller

import (
	"fmt"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/bctnry/gitus/pkg/gitlib"
	"github.com/bctnry/gitus/routes"
	. "github.com/bctnry/gitus/routes"
	"github.com/bctnry/gitus/templates"
)

func handleBranchSnapshotRequest(repo *gitlib.LocalGitRepository, branchName string, obj gitlib.GitObject, w http.ResponseWriter, r *http.Request) {
	filename := fmt.Sprintf(
		"%s-%s-branch-%s",
		repo.Namespace, repo.Name, branchName,
	)
	responseWithTreeZip(repo, obj, filename, w, r)
}

func bindBranchController(ctx *RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/branch/{branchName}/{treePath...}", WithLog(func(w http.ResponseWriter, r *http.Request) {
		rfn := r.PathValue("repoName")
		namespaceName, repoName, repo, err := ctx.ResolveRepositoryFullName(rfn)
		if err != nil {
			errCode := 500
			if routes.IsRouteError(err) {
				if err.(*RouteError).ErrorType == NOT_FOUND {
					errCode = 404
				}
			}
			LogTemplateError(ctx.LoadTemplate("error").Execute(w, templates.ErrorTemplateModel{
				ErrorCode: errCode,
				ErrorMessage: err.Error(),
			}))
			return
		}
		branchName := r.PathValue("branchName")
		treePath := r.PathValue("treePath")
		repoHeaderInfo := templates.RepoHeaderTemplateModel{
			NamespaceName: namespaceName,
			RepoName: rfn,
			RepoDescription: repo.Description,
			TypeStr: "branch",
			NodeName: branchName,
			RepoLabelList: nil,
			RepoURL: fmt.Sprintf("%s/repo/%s", ctx.Config.HttpHostName, rfn),
		}

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
			RepoName: rfn,
			Commit: cobj,
		}
		gobj, err = repo.Repository.ReadObject(cobj.TreeObjId)
		if err != nil { ctx.ReportInternalError(err.Error(), w, r) }
		target, err := repo.Repository.ResolveTreePath(gobj.(*gitlib.TreeObject), treePath)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
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
				handleBranchSnapshotRequest(repo.Repository, branchName, target, w, r)
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
		
		var loginInfo *templates.LoginInfoModel = nil
		if !ctx.Config.PlainMode {
			loginInfo, err = GenerateLoginInfoModel(ctx, r)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
		}
		
		switch target.Type() {
		case gitlib.TREE:
			if len(treePath) > 0 && !strings.HasSuffix(treePath, "/") {
				FoundAt(w, fmt.Sprintf("%s/%s/", rootPath, treePath))
				return
			}
			LogTemplateError(ctx.LoadTemplate("tree").Execute(w, templates.TreeTemplateModel{
				RepoHeaderInfo: repoHeaderInfo,
				TreeFileList: templates.TreeFileListTemplateModel{
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
			}))
		case gitlib.BLOB:
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
			LogTemplateError(ctx.LoadTemplate(templateType).Execute(w, templates.FileTemplateModel{
				RepoHeaderInfo: repoHeaderInfo,
				File: templates.BlobTextTemplateModel{
					FileLineCount: strings.Count(str, "\n"),
					FileContent: str,
				},
				PermaLink: permaLink,
				TreePath: treePathModelValue,
				CommitInfo: commitInfo,
				TagInfo: nil,
				LoginInfo: loginInfo,
			}))
		default:
			ctx.ReportInternalError("", w, r)
		}

	}))
}



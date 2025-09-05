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

func handleCommitSnapshotRequest(repo *gitlib.LocalGitRepository, commitId string, obj gitlib.GitObject, w http.ResponseWriter) {
	filename := fmt.Sprintf(
		"%s-%s-commit-%s",
		repo.Namespace, repo.Name, commitId,
	)
	responseWithTreeZip(repo, obj, filename, w)
}

func bindCommitController(ctx *RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/commit/{commitId}/{treePath...}", WithLog(func(w http.ResponseWriter, r *http.Request) {
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
		if !model.ValidRepositoryName(rfn) {
			ctx.ReportNotFound(rfn, "Repository", "Depot", w, r)
			return
		}
		_, _, ns, repo, err := ctx.ResolveRepositoryFullName(rfn)
		if err == routes.ErrNotFound {
			ctx.ReportNotFound(rfn, "Repository", "Depot", w, r)
			return
		}
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if repo.Type != model.REPO_TYPE_GIT {
			ctx.ReportNormalError("The repository you have requested isn't a Git repository.", w, r)
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
		
		commitId := r.PathValue("commitId")
		treePath := r.PathValue("treePath")

		repoHeaderInfo := GenerateRepoHeader("commit", commitId)

		rr := repo.Repository.(*gitlib.LocalGitRepository)
		gobj, err := rr.ReadObject(commitId)
		if err != nil {
			ctx.ReportObjectReadFailure(commitId, err.Error(), w, r)
			return
		}
		if gobj.Type() != gitlib.COMMIT {
			ctx.ReportObjectTypeMismatch(gobj.ObjectId(), "COMMIT", gobj.Type().String(), w, r)
			return
		}

		cobj := gobj.(*gitlib.CommitObject)
		m := make(map[string]string, 0)
		m[cobj.AuthorInfo.AuthorEmail] = ""
		m[cobj.CommitterInfo.AuthorEmail] = ""
		_, err = ctx.DatabaseInterface.ResolveMultipleEmailToUsername(m)
		// NOTE: we don't check, we just assume the emails are not verified
		// to anyone if an error occur.
		commitInfo := &templates.CommitInfoTemplateModel{
			RootPath: fmt.Sprintf("/repo/%s", rfn),
			Commit: cobj,
			EmailUserMapping: m,
		}
		gobj, err = rr.ReadObject(cobj.TreeObjId)
		if err != nil { ctx.ReportInternalError(err.Error(), w, r) }
		target, err := rr.ResolveTreePath(gobj.(*gitlib.TreeObject), treePath)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
		}

		isSnapshotRequest :=  r.URL.Query().Has("snapshot")
		if isSnapshotRequest {
			if target.Type() == gitlib.BLOB {
				mime := mime.TypeByExtension(path.Ext(treePath))
				if len(mime) <= 0 { mime = "application/octet-stream" }
				w.Header().Add("Content-Type", mime)
				w.Write((target.(*gitlib.BlobObject)).Data)
				return
			} else {
				handleCommitSnapshotRequest(rr, commitId, target, w)
				return
			}
		}

		
		isBlameRequest := r.URL.Query().Has("blame")
		if isBlameRequest {
			if target.Type() == gitlib.BLOB {
				mime := mime.TypeByExtension(path.Ext(treePath))
				if len(mime) <= 0 { mime = "application/octet-stream" }
				if !strings.HasPrefix(mime, "image/") {
					dirPath := path.Dir(treePath) + "/"
					dirObj, err := rr.ResolveTreePath(gobj.(*gitlib.TreeObject), dirPath)
					if err != nil {
						ctx.ReportInternalError(err.Error(), w, r)
						return
					}
					blame, err := rr.Blame(cobj, treePath)
					if err != nil {
						ctx.ReportInternalError(fmt.Sprintf("Failed to run git-blame: %s.", err), w, r)
						return
					}
					LogTemplateError(ctx.LoadTemplate("git-blame").Execute(w, &templates.GitBlameTemplateModel{
						Repository: repo,
						RepoHeaderInfo: *repoHeaderInfo,
						TreeFileList: &templates.TreeFileListTemplateModel{
							ShouldHaveParentLink: len(treePath) > 0,
							RepoPath: fmt.Sprintf("/repo/%s", rfn),
							RootPath: fmt.Sprintf("/repo/%s/%s/%s", rfn, "commit", cobj.Id),
							TreePath: dirPath,
							FileList: dirObj.(*gitlib.TreeObject).ObjectList,
						},
						Blame: blame,
						CommitInfo: commitInfo,
						TagInfo: nil,
						LoginInfo: loginInfo,
						Config: ctx.Config,
					}))
					return
				}
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
		rootFullName := fmt.Sprintf("%s@%s:%s", rfn, "commit", commitId)
		repoPath := fmt.Sprintf("/repo/%s", rfn)
		rootPath := fmt.Sprintf("/repo/%s/%s/%s", rfn, "commit", commitId)
		permaLink := fmt.Sprintf("/repo/%s/commit/%s/%s", rfn, commitId, treePath)
		treePathModelValue := &templates.TreePathTemplateModel{
			RootFullName: rootFullName,
			RootPath: rootPath,
			TreePath: treePath,
			TreePathSegmentList: treePathSegmentList,
		}

		switch target.Type() {
		case gitlib.TREE:
			if len(treePath) > 0 && !strings.HasSuffix(treePath, "/") {
				FoundAt(w, fmt.Sprintf("%s/%s/", rootPath, treePath))
				return
			}
			// TODO: find a better way to do this...
			for i, k := range target.(*gitlib.TreeObject).ObjectList {
				cid, err := rr.ResolvePathLastCommitId(cobj, treePath + k.Name)
				if err != nil { continue }
				o, err := rr.ReadObject(strings.TrimSpace(cid))
				if err != nil { continue }
				target.(*gitlib.TreeObject).ObjectList[i].LastCommit = o.(*gitlib.CommitObject)
			}
			LogTemplateError(ctx.LoadTemplate("tree").Execute(w, templates.TreeTemplateModel{
				Repository: repo,
				RepoHeaderInfo: *repoHeaderInfo,
				TreeFileList: &templates.TreeFileListTemplateModel{
					ShouldHaveParentLink: len(treePath) > 0,
					RepoPath: repoPath,
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
		case gitlib.BLOB:
			dirPath := path.Dir(treePath)
			dirObj, err := rr.ResolveTreePath(gobj.(*gitlib.TreeObject), dirPath)
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
					RepoPath: repoPath,
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

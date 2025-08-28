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

func handleTagSnapshotRequest(repo *gitlib.LocalGitRepository, branchName string, obj gitlib.GitObject, w http.ResponseWriter) error {
	// would resolve tags that point to tags.
	subj := obj
	var err error = nil
	for subj.Type() == gitlib.TAG {
		subj, err = repo.ReadObject((subj.(*gitlib.TagObject)).TaggedObjId)
		if err != nil { return err }
	}
	filename := fmt.Sprintf(
		"%s-%s-branch-%s",
		repo.Namespace, repo.Name, branchName,
	)
	if subj.Type() == gitlib.TREE {
		return responseWithTreeZip(repo, obj, filename, w)
	} else {
		w.Write(subj.RawData())
		return nil
	}
}

func bindTagController(ctx *RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/tag/{tagId}/{treePath...}", WithLog(func(w http.ResponseWriter, r *http.Request) {
		rfn := r.PathValue("repoName")
		if !model.ValidRepositoryName(rfn) {
			ctx.ReportNotFound(rfn, "Repository", "Depot", w, r)
			return
		}
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

		rr := repo.Repository.(*gitlib.LocalGitRepository)
		tagName := r.PathValue("tagId")
		treePath := r.PathValue("treePath")

		// the logic goes as follows:
		// + the subject for resolving would be a tag.
		// + resolve suject into tagInfo w/ one of commit / tree / blob / tag
		// + if the subject is commit, resolves it to commitInfo w/ tree.
		// + if the subject is tree and len(treePath) > 0, resolve tree with treePath
		//   into a tree or a blob.
		// + by now we have a tagInfo/nil, a commitInfo/nil and a tree/blob/tag.
		//   we thus display them accordingly.
		repoHeaderInfo := GenerateRepoHeader("tag", tagName)
		
		err = rr.SyncAllTagList()
		if err != nil {
			ctx.ReportInternalError(
				fmt.Sprintf("Failed to sync tag list: %s", err.Error()),
				w, r,
			)
			return
		}
		t, ok := rr.TagIndex[tagName]
		if !ok {
			ctx.ReportNotFound(tagName, "Tag", rfn, w, r)
			return
		}
		tobj, err := rr.ReadObject(t.HeadId)
		if err != nil {
			ctx.ReportObjectReadFailure(t.HeadId, err.Error(), w, r)
			return
		}

		if r.URL.Query().Has("snapshot") {
			handleTagSnapshotRequest(rr, tagName, tobj, w)
			return
		}

		// NOTE: the part about permalink is slightly tricky.
		// + if a tag points to a tag, then the permalink is the same as the link.
		// + if a tag points to a commit, then the permalink should be that of
		//   the commit.
		// + if a tag points to anything else then the permalink should be that
		//   of those.
		permaLink := ""
		
		var subject gitlib.GitObject = nil
		var tagInfo *templates.TagInfoTemplateModel = nil

		if tobj.Type() == gitlib.TAG {
			to := tobj.(*gitlib.TagObject)
			tagInfo = &templates.TagInfoTemplateModel{
				Annotated: true,
				RepoName: rfn,
				Tag: to,
			}
			subject, err = rr.ReadObject(to.TaggedObjId)
			if err != nil {
				ctx.ReportObjectReadFailure(to.TaggedObjId, err.Error(), w, r)
				return
			}
		} else {
			subject = tobj
		}

		var commitInfo *templates.CommitInfoTemplateModel = nil

		if subject.Type() == gitlib.COMMIT {
			cobj, ok := subject.(*gitlib.CommitObject)
			if !ok {
				ctx.ReportInternalError(
					fmt.Sprintf(
						"Shouldn't happen - object with COMMIT type but not parsed as commit object. ObjId: %s", subject.ObjectId(),
					),
					w, r,
				)
				return
			}
			commitInfo = &templates.CommitInfoTemplateModel{
				RootPath: fmt.Sprintf("/repo/%s", rfn),
				Commit: cobj,
			}
			subject, err = rr.ReadObject(cobj.TreeObjId)
			if err != nil {
				ctx.ReportObjectReadFailure(
					cobj.TreeObjId,
					fmt.Sprintf("%s (commit %s)", err.Error(), cobj.Id),
					w, r,
				)
				return
			}
			permaLink = fmt.Sprintf("/repo/%s/commit/%s/%s", rfn, cobj.Id, treePath)
		}

		if subject.Type() == gitlib.TREE {
			subject, err = rr.ResolveTreePath(subject.(*gitlib.TreeObject), treePath)
			if subject.Type() == gitlib.TREE && len(treePath) > 0 && !strings.HasSuffix(treePath, "/") {
				FoundAt(w, fmt.Sprintf("/repo/%s/tag/%s/%s/", rfn, tagName, treePath))
				return
			}
			if len(permaLink) <= 0 {
				permaLink = fmt.Sprintf("/repo/%s/tree/%s/%s", rfn, subject.ObjectId(), treePath)
			}
		}

		switch subject.Type() {
		case gitlib.TAG:
			tobj, ok := subject.(*gitlib.TagObject)
			if !ok {
				ctx.ReportInternalError(
					fmt.Sprintf(
						"Shouldn't happen - object with TAG type but not parsed as tag object. ObjId: %s", subject.ObjectId(),
					),
					w, r,
				)
				return
			}
			LogTemplateError(ctx.LoadTemplate("tag").Execute(w, templates.TagTemplateModel{
				Repository: repo,
				RepoHeaderInfo: *repoHeaderInfo,
				Tag: tobj,
				TagInfo: tagInfo,
				LoginInfo: loginInfo,
			}))
			return
		case gitlib.BLOB:
			mime := mime.TypeByExtension(path.Ext(treePath))
			if len(mime) <= 0 { mime = "application/octet-stream" }
			templateType := "file-text"
			if strings.HasPrefix(mime, "image/") {
				templateType = "file-image"
			}
			bobj := subject.(*gitlib.BlobObject)
			if r.URL.Query().Has("raw") {
				w.Header().Add("Content-Type", mime)
				w.Write(bobj.Data)
				return
			}
			if len(permaLink) <= 0 {
				permaLink = fmt.Sprintf("/repo/%s/blob/%s", rfn, bobj.Id)
			}
			str := string(bobj.Data)
			coloredStr, err := colorSyntax("", str)
			if err == nil { str = coloredStr }
			LogTemplateError(ctx.LoadTemplate(templateType).Execute(w, templates.FileTemplateModel{
				Repository: repo,
				RepoHeaderInfo: *repoHeaderInfo,
				File: templates.BlobTextTemplateModel{
					FileLineCount: strings.Count(str, "\n"),
					FileContent: str,
				},
				PermaLink: permaLink,
				TreePath: nil,
				CommitInfo: commitInfo,
				TagInfo: tagInfo,
				LoginInfo: loginInfo,
				Config: ctx.Config,
			}))

		case gitlib.TREE:
			rootFullName := fmt.Sprintf("%s@%s:%s", rfn, "tag", tagName)
			rootPath := fmt.Sprintf("/repo/%s/%s/%s", rfn, "tag", tagName)
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
			target, err := rr.ResolveTreePath(subject.(*gitlib.TreeObject), treePath)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
			}
			treePathModelValue := &templates.TreePathTemplateModel{
				RootFullName: rootFullName,
				RootPath: rootPath,
				TreePath: treePath,
				TreePathSegmentList: treePathSegmentList,
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
				PermaLink: permaLink,
				TreePath: treePathModelValue,
				CommitInfo: commitInfo,
				TagInfo: nil,
				LoginInfo: loginInfo,
				Config: ctx.Config,
			}))
			return
		default:
			ctx.ReportInternalError(
				fmt.Sprintf("Shouldn't happen: object type expected to be one of tag/blob/tree but it's %s instead.", subject.Type().String()),
				w, r,
			)
		}
	}))
}

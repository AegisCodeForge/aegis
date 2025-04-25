package controller

import (
	"fmt"
	"path"
	"mime"
	"strings"
	"net/http"
	
	. "github.com/bctnry/gitus/routes"
	"github.com/bctnry/gitus/templates"
	"github.com/bctnry/gitus/pkg/gitlib"
)

func bindTagController(ctx RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/tag/{tagId}/{treePath...}", WithLog(func(w http.ResponseWriter, r *http.Request) {
		repoName := r.PathValue("repoName")
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
		repo, ok := ctx.GitRepositoryList[repoName]
		if !ok {
			ctx.ReportNotFound(repoName, "Repository", "depot", w, r)
			return
		}
		repoHeaderInfo := templates.RepoHeaderTemplateModel{
			RepoName: repoName,
			RepoDescription: repo.Description,
			TypeStr: "tag",
			NodeName: tagName,
			RepoLabelList: nil,
		}
		
		err := repo.SyncAllTagList()
		if err != nil {
			ctx.ReportInternalError(
				fmt.Sprintf("Failed to sync tag list: %s", err.Error()),
				w, r,
			)
			return
		}
		t, ok := repo.TagIndex[tagName]
		if !ok {
			ctx.ReportNotFound(tagName, "Tag", repoName, w, r)
			return
		}
		tobj, err := repo.ReadObject(t.HeadId)
		if err != nil {
			ctx.ReportObjectReadFailure(t.HeadId, err.Error(), w, r)
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
				RepoName: repoName,
				Tag: to,
			}
			subject, err = repo.ReadObject(to.TaggedObjId)
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
				RepoName: repoName,
				Commit: cobj,
			}
			subject, err = repo.ReadObject(cobj.TreeObjId)
			if err != nil {
				ctx.ReportObjectReadFailure(
					cobj.TreeObjId,
					fmt.Sprintf("%s (commit %s)", err.Error(), cobj.Id),
					w, r,
				)
				return
			}
			permaLink = fmt.Sprintf("/repo/%s/commit/%s/%s", repoName, cobj.Id, treePath)
		}

		if subject.Type() == gitlib.TREE {
			subject, err = repo.ResolveTreePath(subject.(*gitlib.TreeObject), treePath)
			if subject.Type() == gitlib.TREE && len(treePath) > 0 && !strings.HasSuffix(treePath, "/") {
				FoundAt(w, fmt.Sprintf("/repo/%s/tag/%s/%s/", repoName, tagName, treePath))
				return
			}
			if len(permaLink) <= 0 {
				permaLink = fmt.Sprintf("/repo/%s/tree/%s/%s", repoName, subject.ObjectId(), treePath)
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
				RepoHeaderInfo: repoHeaderInfo,
				Tag: tobj,
				TagInfo: tagInfo,
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
				permaLink = fmt.Sprintf("/repo/%s/blob/%s", repoName, bobj.Id)
			}
			str := string(bobj.Data)
			LogTemplateError(ctx.LoadTemplate(templateType).Execute(w, templates.FileTemplateModel{
				RepoHeaderInfo: repoHeaderInfo,
				File: templates.BlobTextTemplateModel{
					FileLineCount: strings.Count(str, "\n"),
					FileContent: str,
				},
				PermaLink: permaLink,
				TreePath: nil,
				CommitInfo: commitInfo,
				TagInfo: tagInfo,
			}))

		case gitlib.TREE:
			rootFullName := fmt.Sprintf("%s@%s:%s", repoName, "tag", tagName)
			rootPath := fmt.Sprintf("/repo/%s/%s/%s", repoName, "tag", tagName)
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
			target, err := repo.ResolveTreePath(subject.(*gitlib.TreeObject), treePath)
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

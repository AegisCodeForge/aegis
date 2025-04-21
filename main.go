package main

//go:generate go run devtools/generate-template.go templates

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/bctnry/gitus/pkg/gitlib"
	"github.com/bctnry/gitus/templates"
)

// go don't have ufcs so i'll have to suffer.
func withLog(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(fmt.Sprintf(" %s %s", r.Method, r.URL.Path))
		f(w, r)
	}
}
func withLogHandler(f http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(fmt.Sprintf(" %s %s", r.Method, r.URL.Path))
		f.ServeHTTP(w, r)
	}
}

func foundAt(w http.ResponseWriter, p string) {
	w.Header().Add("Content-Length", "0")
	w.Header().Add("Location", p)
	w.WriteHeader(302)
}

func (ctx RouterContext) loadTemplate(name string) *template.Template {
	return ctx.MasterTemplate.Lookup(name)
}

func logTemplateError(e error) {
	if e != nil { log.Print(e) }
}

func (ctx RouterContext) reportNotFound(objName string, objType string, namespace string, w http.ResponseWriter, r *http.Request) {
	logTemplateError(ctx.loadTemplate("error").Execute(w,
		templates.ErrorTemplateModel{
			ErrorCode: 404,
			ErrorMessage: fmt.Sprintf(
				"%s %s not found in %s",
				objType, objName, namespace,
			),
		},
	))
}

func (ctx RouterContext) reportInternalError(msg string, w http.ResponseWriter, r *http.Request) {
	logTemplateError(ctx.loadTemplate("error").Execute(w,
		templates.ErrorTemplateModel{
			ErrorCode: 500,
			ErrorMessage: fmt.Sprintf(
				"Internal error: %s",
				msg,
			),
		},
	))
}

func (ctx RouterContext) reportObjectReadFailure(objid string, msg string, w http.ResponseWriter, r *http.Request) {
	ctx.reportInternalError(
		fmt.Sprintf(
			"Fail to read object %s: %s",
			objid, msg,
		), w, r,
	)
}

func (ctx RouterContext) reportObjectTypeMismatch(objid string, expectedType string, actualType string, w http.ResponseWriter, r *http.Request) {
	ctx.reportInternalError(
		fmt.Sprintf(
			"Object type mismatch for %s: %s expected but %s found",
			objid, expectedType, actualType,
		), w, r,
	)
}

func (ctx RouterContext) branchHandler(repoName string, branchName string, treePath string, w http.ResponseWriter, r *http.Request) {
	repo, ok := ctx.GitRepositoryList[repoName]
	if !ok {
		ctx.reportNotFound(repoName, "Repository", "depot", w, r)
		return
	}
	repoHeaderInfo := templates.RepoHeaderTemplateModel{
		RepoName: repoName,
		RepoDescription: repo.Description,
		TypeStr: "branch",
		NodeName: branchName,
	}

	err := repo.SyncAllBranchList()
	if err != nil {
		ctx.reportInternalError(
			fmt.Sprintf(
				"Cannot sync branch list for %s: %s",
				repoName,
				err.Error(),
			), w, r,
		)
		return
	}
	br, ok := repo.BranchIndex[branchName]
	if !ok {
		ctx.reportNotFound(branchName, "Branch", repoName, w, r)
		return
	}
	gobj, err := repo.ReadObject(br.HeadId)
	if err != nil {
		ctx.reportObjectReadFailure(br.HeadId, err.Error(), w, r)
		return
	}
	if gobj.Type() != gitlib.COMMIT {
		ctx.reportObjectTypeMismatch(gobj.ObjectId(), "COMMIT", gobj.Type().String(), w, r)
		return
	}

	cobj := gobj.(*gitlib.CommitObject)
	commitInfo := &templates.CommitInfoTemplateModel{
		RepoName: repoName,
		Commit: cobj,
	}
	gobj, err = repo.ReadObject(cobj.TreeObjId)
	if err != nil { ctx.reportInternalError(err.Error(), w, r) }
	target, err := repo.ResolveTreePath(gobj.(*gitlib.TreeObject), treePath)
	if err != nil {
		ctx.reportInternalError(err.Error(), w, r)
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
	rootFullName := fmt.Sprintf("%s@%s:%s", repoName, "branch", branchName)
	rootPath := fmt.Sprintf("/repo/%s/%s/%s", repoName, "branch", branchName)
	treePathModelValue := &templates.TreePathTemplateModel{
		RootFullName: rootFullName,
		RootPath: rootPath,
		TreePath: treePath,
		TreePathSegmentList: treePathSegmentList,
	}
	switch target.Type() {
	case gitlib.TREE:
		if len(treePath) > 0 && !strings.HasSuffix(treePath, "/") {
			foundAt(w, fmt.Sprintf("%s/%s/", rootPath, treePath))
			return
		}
		logTemplateError(ctx.loadTemplate("tree").Execute(w, templates.TreeTemplateModel{
			RepoHeaderInfo: repoHeaderInfo,
			TreeFileList: templates.TreeFileListTemplateModel{
				ShouldHaveParentLink: len(treePath) > 0,
				RootPath: rootPath,
				TreePath: treePath,
				FileList: target.(*gitlib.TreeObject).ObjectList,
			},
			TreePath: treePathModelValue,
			CommitInfo: commitInfo,
			TagInfo: nil,
		}))
	case gitlib.BLOB:
		bobj := target.(*gitlib.BlobObject)
		str := string(bobj.Data)
		logTemplateError(ctx.loadTemplate("file-text").Execute(w, templates.FileTextTemplateModel{
			RepoHeaderInfo: repoHeaderInfo,
			File: templates.BlobTextTemplateModel{
				FileLineCount: strings.Count(str, "\n"),
				FileContent: str,
			},
			TreePath: treePathModelValue,
			CommitInfo: commitInfo,
			TagInfo: nil,
		}))
	default:
		ctx.reportInternalError("", w, r)
	}
}

func (ctx RouterContext) commitHandler(repoName string, commitId string, treePath string, w http.ResponseWriter, r *http.Request) {
	repo, ok := ctx.GitRepositoryList[repoName]
	if !ok {
		ctx.reportNotFound(repoName, "Repository", "depot", w, r)
		return
	}
	repoHeaderInfo := templates.RepoHeaderTemplateModel{
		RepoName: repoName,
		RepoDescription: repo.Description,
		TypeStr: "commit",
		NodeName: commitId,
	}

	gobj, err := repo.ReadObject(commitId)
	if err != nil {
		ctx.reportObjectReadFailure(commitId, err.Error(), w, r)
		return
	}
	if gobj.Type() != gitlib.COMMIT {
		ctx.reportObjectTypeMismatch(gobj.ObjectId(), "COMMIT", gobj.Type().String(), w, r)
		return
	}

	cobj := gobj.(*gitlib.CommitObject)
	commitInfo := &templates.CommitInfoTemplateModel{
		RepoName: repoName,
		Commit: cobj,
	}
	gobj, err = repo.ReadObject(cobj.TreeObjId)
	if err != nil { ctx.reportInternalError(err.Error(), w, r) }
	target, err := repo.ResolveTreePath(gobj.(*gitlib.TreeObject), treePath)
	if err != nil {
		ctx.reportInternalError(err.Error(), w, r)
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
	rootFullName := fmt.Sprintf("%s@%s:%s", repoName, "commit", commitId)
	rootPath := fmt.Sprintf("/repo/%s/%s/%s", repoName, "commit", commitId)
	treePathModelValue := &templates.TreePathTemplateModel{
		RootFullName: rootFullName,
		RootPath: rootPath,
		TreePath: treePath,
		TreePathSegmentList: treePathSegmentList,
	}
	switch target.Type() {
	case gitlib.TREE:
		if len(treePath) > 0 && !strings.HasSuffix(treePath, "/") {
			foundAt(w, fmt.Sprintf("%s/%s/", rootPath, treePath))
			return
		}
		logTemplateError(ctx.loadTemplate("tree").Execute(w, templates.TreeTemplateModel{
			RepoHeaderInfo: repoHeaderInfo,
			TreeFileList: templates.TreeFileListTemplateModel{
				ShouldHaveParentLink: len(treePath) > 0,
				RootPath: rootPath,
				TreePath: treePath,
				FileList: target.(*gitlib.TreeObject).ObjectList,
			},
			TreePath: treePathModelValue,
			CommitInfo: commitInfo,
			TagInfo: nil,
		}))
	case gitlib.BLOB:
		bobj := target.(*gitlib.BlobObject)
		str := string(bobj.Data)
		logTemplateError(ctx.loadTemplate("file-text").Execute(w, templates.FileTextTemplateModel{
			RepoHeaderInfo: repoHeaderInfo,
			File: templates.BlobTextTemplateModel{
				FileLineCount: strings.Count(str, "\n"),
				FileContent: str,
			},
			TreePath: treePathModelValue,
			CommitInfo: commitInfo,
			TagInfo: nil,
		}))
	default:
		ctx.reportInternalError("", w, r)
	}
}

func (ctx RouterContext) treeHandler(repoName string, treeId string, treePath string, w http.ResponseWriter, r *http.Request) {
	repo, ok := ctx.GitRepositoryList[repoName]
	if !ok {
		ctx.reportNotFound(repoName, "Repository", "depot", w, r)
		return
	}
	repoHeaderInfo := templates.RepoHeaderTemplateModel{
		RepoName: repoName,
		RepoDescription: repo.Description,
		TypeStr: "tree",
		NodeName: treeId,
	}

	gobj, err := repo.ReadObject(treeId)
	if err != nil {
		ctx.reportObjectReadFailure(treeId, err.Error(), w, r)
		return
	}
	if gobj.Type() != gitlib.TREE {
		ctx.reportObjectTypeMismatch(gobj.ObjectId(), "TREE", gobj.Type().String(), w, r)
		return
	}

	rootFullName := fmt.Sprintf("%s@%s:%s", repoName, "tree", treeId)
	rootPath := fmt.Sprintf("/repo/%s/%s/%s", repoName, "tree", treeId)
	tobj := gobj.(*gitlib.TreeObject)
	var commitInfo *templates.CommitInfoTemplateModel = nil
	target, err := repo.ResolveTreePath(tobj, treePath)
	if err != nil {
		ctx.reportInternalError(err.Error(), w, r)
	}
	if target.Type() == gitlib.TREE && len(treePath) > 0 && !strings.HasSuffix(treePath, "/") {
		foundAt(w, fmt.Sprintf("%s/%s/", rootPath, treePath))
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
	logTemplateError(ctx.loadTemplate("tree").Execute(w, templates.TreeTemplateModel{
		RepoHeaderInfo: repoHeaderInfo,
		TreeFileList: templates.TreeFileListTemplateModel{
			ShouldHaveParentLink: len(treePath) > 0,
			RootPath: rootPath,
			TreePath: treePath,
			FileList: target.(*gitlib.TreeObject).ObjectList,
		},
		TreePath: treePathModelValue,
		CommitInfo: commitInfo,
		TagInfo: nil,
	}))
}

func (ctx RouterContext) blobHandler(repoName string, blobId string, w http.ResponseWriter, r *http.Request) {
	repo, ok := ctx.GitRepositoryList[repoName]
	if !ok {
		ctx.reportNotFound(repoName, "Repository", "depot", w, r)
		return
	}
	repoHeaderInfo := templates.RepoHeaderTemplateModel{
		RepoName: repoName,
		RepoDescription: repo.Description,
		TypeStr: "blob",
		NodeName: blobId,
	}

	gobj, err := repo.ReadObject(blobId)
	if err != nil {
		ctx.reportObjectReadFailure(blobId, err.Error(), w, r)
		return
	}
	if gobj.Type() != gitlib.BLOB {
		ctx.reportObjectTypeMismatch(gobj.ObjectId(), "BLOB", gobj.Type().String(), w, r)
		return
	}

	bobj := gobj.(*gitlib.BlobObject)
	str := string(bobj.Data)
	logTemplateError(ctx.loadTemplate("file-text").Execute(w, templates.FileTextTemplateModel{
		RepoHeaderInfo: repoHeaderInfo,
		File: templates.BlobTextTemplateModel{
			FileLineCount: strings.Count(str, "\n"),
			FileContent: str,
		},
		TreePath: nil,
		CommitInfo: nil,
		TagInfo: nil,
	}))
}


func (ctx RouterContext) tagHandler(repoName string, tagName string, treePath string, w http.ResponseWriter, r *http.Request) {
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
		ctx.reportNotFound(repoName, "Repository", "depot", w, r)
		return
	}
	repoHeaderInfo := templates.RepoHeaderTemplateModel{
		RepoName: repoName,
		RepoDescription: repo.Description,
		TypeStr: "tag",
		NodeName: tagName,
	}
	
	err := repo.SyncAllTagList()
	if err != nil {
		ctx.reportInternalError(
			fmt.Sprintf("Failed to sync tag list: %s", err.Error()),
			w, r,
		)
		return
	}
	t, ok := repo.TagIndex[tagName]
	if !ok {
		ctx.reportNotFound(tagName, "Tag", repoName, w, r)
		return
	}
	tobj, err := repo.ReadObject(t.HeadId)
	if err != nil {
		ctx.reportObjectReadFailure(t.HeadId, err.Error(), w, r)
		return
	}
	
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
			ctx.reportObjectReadFailure(to.TaggedObjId, err.Error(), w, r)
			return
		}
	} else {
		subject = tobj
	}

	var commitInfo *templates.CommitInfoTemplateModel = nil

	if subject.Type() == gitlib.COMMIT {
		cobj, ok := subject.(*gitlib.CommitObject)
		if !ok {
			ctx.reportInternalError(
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
			ctx.reportObjectReadFailure(
				cobj.TreeObjId,
				fmt.Sprintf("%s (commit %s)", err.Error(), cobj.Id),
				w, r,
			)
			return
		}
	}

	if subject.Type() == gitlib.TREE {
		subject, err = repo.ResolveTreePath(subject.(*gitlib.TreeObject), treePath)
		if subject.Type() == gitlib.TREE && len(treePath) > 0 && !strings.HasSuffix(treePath, "/") {
			foundAt(w, fmt.Sprintf("/repo/%s/tag/%s/%s/", repoName, tagName, treePath))
			return
		}
	}

	switch subject.Type() {
	case gitlib.TAG:
		tobj, ok := subject.(*gitlib.TagObject)
		if !ok {
			ctx.reportInternalError(
				fmt.Sprintf(
					"Shouldn't happen - object with TAG type but not parsed as tag object. ObjId: %s", subject.ObjectId(),
				),
				w, r,
			)
			return
		}
		logTemplateError(ctx.loadTemplate("tag").Execute(w, templates.TagTemplateModel{
			RepoHeaderInfo: repoHeaderInfo,
			Tag: tobj,
			TagInfo: tagInfo,
		}))
		return
	case gitlib.BLOB:
		bobj := subject.(*gitlib.BlobObject)
		str := string(bobj.Data)
		logTemplateError(ctx.loadTemplate("file-text").Execute(w, templates.FileTextTemplateModel{
			RepoHeaderInfo: repoHeaderInfo,
			File: templates.BlobTextTemplateModel{
				FileLineCount: strings.Count(str, "\n"),
				FileContent: str,
			},
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
			ctx.reportInternalError(err.Error(), w, r)
		}
		treePathModelValue := &templates.TreePathTemplateModel{
			RootFullName: rootFullName,
			RootPath: rootPath,
			TreePath: treePath,
			TreePathSegmentList: treePathSegmentList,
		}
		logTemplateError(ctx.loadTemplate("tree").Execute(w, templates.TreeTemplateModel{
			RepoHeaderInfo: repoHeaderInfo,
			TreeFileList: templates.TreeFileListTemplateModel{
				ShouldHaveParentLink: len(treePath) > 0,
				RootPath: rootPath,
				TreePath: treePath,
				FileList: target.(*gitlib.TreeObject).ObjectList,
			},
			TreePath: treePathModelValue,
			CommitInfo: commitInfo,
			TagInfo: nil,
		}))
		return
	default:
		ctx.reportInternalError(
			fmt.Sprintf("Shouldn't happen: object type expected to be one of tag/blob/tree but it's %s instead.", subject.Type().String()),
			w, r,
		)
	}
}


const gitPath = "/home/bctnry/workspace/"

func loadTemplate(t *template.Template, name string) *template.Template {
	res := t.Lookup(name)
	if res == nil { log.Fatal(fmt.Sprintf("Failed to find template \"%s\"", name)) }
	return res
}

func getAllGitRepository() (map[string]gitlib.LocalGitRepository, error) {
	res := make(map[string]gitlib.LocalGitRepository, 0)
	l, err := os.ReadDir(gitPath)
	if err != nil { return nil, err }
	for _, item := range l {
		p := path.Join(gitPath, item.Name())
		if gitlib.IsValidGitDirectory(p) {
			res[item.Name()] = gitlib.NewLocalGitRepository(p)
			continue
		}
		p2 := path.Join(p, ".git")
		if gitlib.IsValidGitDirectory(p2) {
			res[item.Name() + ".git"] = gitlib.NewLocalGitRepository(p2)
			continue
		}
	}
	return res, nil
}

func check(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

func logIfError(err error) {
	if err != nil {
		log.Print(err.Error())
	}
}

type RouterContext struct {
	MasterTemplate *template.Template
	GitRepositoryList map[string]gitlib.LocalGitRepository
}

func main() {
	masterTemplate := templates.LoadTemplate()
	t := loadTemplate(masterTemplate, "index")
	tRepository := loadTemplate(masterTemplate, "repository")

	grlist, err := getAllGitRepository()
	check(err)
	grmodel := make([]struct{RelPath string; Description string}, 0)
	for key, item := range grlist {
		grmodel = append(grmodel, struct{
			RelPath string
			Description string
		}{
			RelPath: key,
			Description: item.Description,
		})
	}

	context := RouterContext{
		MasterTemplate: masterTemplate,
		GitRepositoryList: grlist,
	}

	var tErr error = nil

	http.HandleFunc("GET /", withLog(func(w http.ResponseWriter, r *http.Request) {
		t.Execute(w, templates.IndexModel{
			RepositoryList: grmodel,
		})
	}))

	http.HandleFunc("GET /repo/{repoName}/", withLog(func(w http.ResponseWriter, r *http.Request) {
		rn := r.PathValue("repoName")
		s, ok := grlist[rn]
		if !ok {
			tErr = loadTemplate(masterTemplate, "error").Execute(w, templates.ErrorTemplateModel{
				ErrorCode: 404,
				ErrorMessage: fmt.Sprintf("Repository %s not found.", rn),
			})
			logIfError(tErr)
			return
		}

		err := s.SyncAllBranchList()
		check(err)
		err = s.SyncAllTagList()
		check(err)

		tRepository.Execute(w, templates.RepositoryModel{
			RepoName: rn,
			RepoObj: s,
			BranchList: s.BranchIndex,
			TagList: s.TagIndex,
		})
	}))

	
	http.HandleFunc("GET /repo/{repoName}/branch/{branchName}/{treePath...}", withLog(func(w http.ResponseWriter, r *http.Request) {
		context.branchHandler(r.PathValue("repoName"), r.PathValue("branchName"), r.PathValue("treePath"), w, r)
	}))
	http.HandleFunc("GET /repo/{repoName}/commit/{commitId}/{treePath...}", withLog(func(w http.ResponseWriter, r *http.Request) {
		context.commitHandler(r.PathValue("repoName"), r.PathValue("commitId"), r.PathValue("treePath"), w, r)
	}))
	http.HandleFunc("GET /repo/{repoName}/tree/{treeId}/{treePath...}", withLog(func(w http.ResponseWriter, r *http.Request) {
		context.treeHandler(r.PathValue("repoName"), r.PathValue("treeId"), r.PathValue("treePath"), w, r)
	}))
	http.HandleFunc("GET /repo/{repoName}/blob/{blobId}/", withLog(func(w http.ResponseWriter, r *http.Request) {
		context.blobHandler(r.PathValue("repoName"), r.PathValue("blobId"), w, r)
	}))
	http.HandleFunc("GET /repo/{repoName}/tag/{tagId}/{treePath...}", withLog(func(w http.ResponseWriter, r *http.Request) {
		context.tagHandler(r.PathValue("repoName"), r.PathValue("tagId"), r.PathValue("treePath"), w, r)
	}))
	

	http.HandleFunc("GET /repo/{repoName}/history/{nodeName}", withLog(func(w http.ResponseWriter, r *http.Request) {
		rn := r.PathValue("repoName")
		repo, ok := grlist[rn]
		if !ok {
			tErr = loadTemplate(masterTemplate, "error").Execute(w, templates.ErrorTemplateModel{
				ErrorCode: 404,
				ErrorMessage: fmt.Sprintf("Repository %s not found.", rn),
			})
			logIfError(tErr)
			return
		}
		nodeName := r.PathValue("nodeName")
		nodeNameElem := strings.Split(nodeName, ":")
		typeStr := string(nodeNameElem[0])
		cid := string(nodeNameElem[1])
		if string(nodeNameElem[0]) == "branch" {
			err := repo.SyncAllBranchList()
			if err != nil {
				tErr = loadTemplate(masterTemplate, "error").Execute(w, templates.ErrorTemplateModel{
					ErrorCode: 500,
					ErrorMessage: fmt.Sprintf("Failed at syncing branch list for repository %s: %s", rn, err.Error()),
				})
				logIfError(tErr)
				return
			}
			br, ok := repo.BranchIndex[string(nodeNameElem[1])]
			if !ok {
				tErr = loadTemplate(masterTemplate, "error").Execute(w, templates.ErrorTemplateModel{
					ErrorCode: 404,
					ErrorMessage: fmt.Sprintf("Repository %s not found.", rn),
				})
				logIfError(tErr)
				return
			}
			cid = br.HeadId
		}
		cobj, err := repo.ReadObject(cid)
		if err != nil {
			tErr = loadTemplate(masterTemplate, "error").Execute(w, templates.ErrorTemplateModel{
				ErrorCode: 500,
				ErrorMessage: fmt.Sprintf(
					"Failed to read commit object %s: %s",
					cid,
					err,
				),
			})
			logIfError(tErr)
			return
		}
		h, err := repo.GetCommitHistory(cid)
		if err != nil {
			tErr = loadTemplate(masterTemplate, "error").Execute(w, templates.ErrorTemplateModel{
				ErrorCode: 500,
				ErrorMessage: fmt.Sprintf(
					"Failed to read commit history of object %s: %s",
					cid,
					err,
				),
			})
			logIfError(tErr)
			return
		}
		check(err)
		tErr = loadTemplate(masterTemplate, "commit-history").Execute(
			w,
			templates.CommitHistoryModel{
				RepoHeaderInfo: templates.RepoHeaderTemplateModel{
					RepoName: rn,
					RepoDescription: repo.Description,
					TypeStr: typeStr,
					NodeName: nodeNameElem[1],
				},
				Commit: *(cobj.(*gitlib.CommitObject)),
				CommitHistory: h,
			},
		)
		logIfError(tErr)
	}))

	http.HandleFunc("GET /repo/{repoName}/diff/{commitId}/", withLog(func(w http.ResponseWriter, r *http.Request){
		rn := r.PathValue("repoName")
		repo, ok := grlist[rn]
		if !ok {
			tErr = loadTemplate(masterTemplate, "error").Execute(w, templates.ErrorTemplateModel{
				ErrorCode: 404,
				ErrorMessage: fmt.Sprintf("Repository %s not found.", rn),
			})
			logIfError(tErr)
			return
		}
		commitId := r.PathValue("commitId")
		cobj, err := repo.ReadObject(commitId)
		if err != nil {
			tErr = loadTemplate(masterTemplate, "error").Execute(w, templates.ErrorTemplateModel{
				ErrorCode: 500,
				ErrorMessage: fmt.Sprintf("Failed to read commit %s: $s", commitId, err),
			})
			logIfError(tErr)
			return
		}
		diff, err := repo.GetDiff(commitId)
		if err != nil {
			tErr = loadTemplate(masterTemplate, "error").Execute(w, templates.ErrorTemplateModel{
				ErrorCode: 500,
				ErrorMessage: fmt.Sprintf("Failed to read diff %s: $s", commitId, err),
			})
			logIfError(tErr)
			return
		}
		tErr = loadTemplate(masterTemplate, "diff").Execute(w, templates.DiffTemplateModel{
			RepoHeaderInfo: templates.RepoHeaderTemplateModel{
				RepoName: rn,
				RepoDescription: repo.Description,
				TypeStr: "commit",
				NodeName: commitId,
			},
			CommitInfo: templates.CommitInfoTemplateModel{
				RepoName: rn,
				Commit: cobj.(*gitlib.CommitObject),
			},
			Diff: diff,
		})

	}))

	var fs = http.FileServer(http.Dir("static/"))
	http.Handle("GET /favicon.ico", withLogHandler(fs))
	http.Handle("GET /static/", http.StripPrefix("/static/", withLogHandler(fs)))

	log.Println("Serve at :8000")
	http.ListenAndServe(":8000", nil)
}

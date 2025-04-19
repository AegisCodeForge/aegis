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

func main() {
	masterTemplate := templates.LoadTemplate()
	var check = func(err error) {
		if err != nil {
			log.Fatal(err.Error())
		}
	}

	var logIfError = func(err error) {
		if err != nil {
			log.Print(err.Error())
		}
	}

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

	
	http.HandleFunc("GET /repo/{repoName}/{typeStr}/{nodeName}/{treePath...}", withLog(func(w http.ResponseWriter, r *http.Request) {
		// NOTE: typeStr can have the following value:
		//     branch, commit, tree, blob, tag
		// the logic is as follows:
		// + branch: read branch,
		//           read commit from branch,
		//           read tree from commit,
		//           display tree / file depending on the result.
		// + commit: read tree from commit,
		//           display tree / file depending on the result.
		// + tree: read tree,
		//         display tree / file depending on the result.
		// + blob: read blog,
		//         display file,
		// + tag: read tag.
		//        if it's tag object:
		//            prepare tagInfo as "annotated tag".
		//            read pointed object.
		//            if pointed object not tag: follow above
		//            if pointed object is tag: display tag.
		//        if it's *not* tag object
		//            prepare tagInfo as "lightweight tag".
		//            try the rest 3 types of object.
		//            if tag: display tag.
		//            if not tag: follow above.
		// for this reason the order of things is as follows:
		// 1.  prepare `tagInfo` and `subject`. if typeStr is tag, setup
		//     `repoHeaderInfo`, `tagInfo` and `subject` accordingly. or
		//     else, setup `repoHeaderInfo` to be "tag", `tagInfo` to be nil.
		//     `subject` must be a GitObject.
		// 2.  check if typeStr is branch. if yes, setup `repoHeaderInfo` to
		//     be branch and resolve the branch into id to a commit object.
		//     setup `subject` to be that commit object.
		// 3.  check if typeStr is blob or tree. if yes, setup `repoHeaderInfo`
		//     and `subject` accordingly.
		// 4.  at this point, if `subject` is commit, setup `commitInfo` and
		//     resolve `subject` to a tree.
		// 5.  at this point, `subject` is either a tag, a blob or a tree.
		//     display it accordingly.
		// commit technicall should have two views: a tree view and a diff view.
		// we currently only have a tree view. it's easy to change if we later
		// add a diff view - just remove step 4 and change step 5.
		// the main source of complexity comes from the fact that there's two
		// kinds of tags and tagging non-commit objects has a limited but
		// legitimate real-world use.
		rn := r.PathValue("repoName")
		repo, ok := grlist[rn]
		if !ok {
			tErr = loadTemplate(masterTemplate, "error").Execute(w, templates.ErrorTemplateModel{
				ErrorCode: 500,
				ErrorMessage: fmt.Sprintf(
					"Repository %s not found.",
					rn,
					err.Error(),
				),
			})
			logIfError(tErr)
			return
		}
		typeStr := r.PathValue("typeStr")
		nodeName := r.PathValue("nodeName")
		repoHeaderInfo := templates.RepoHeaderTemplateModel{
			RepoName: rn,
			RepoDescription: repo.Description,
			TypeStr: typeStr,
			NodeName: nodeName,
		}
		
		var subject gitlib.GitObject = nil
		var err error = nil

		var tagInfo *templates.TagInfoTemplateModel = nil
		if typeStr == "tag" {
			err = repo.SyncAllTagList()
			if err != nil {
				tErr = loadTemplate(masterTemplate, "error").Execute(w, templates.ErrorTemplateModel{
					ErrorCode: 500,
					ErrorMessage: fmt.Sprintf(
						"Failed to sync tag list: %s",
						err.Error(),
					),
				})
				logIfError(tErr)
				return
			}
			t, ok := repo.TagIndex[nodeName]
			if !ok {
				loadTemplate(masterTemplate, "error").Execute(w, templates.ErrorTemplateModel{
					ErrorCode: 404,
					ErrorMessage: fmt.Sprintf(
						"Tag %s not found in repository %s",
						nodeName, rn,
					),
				})
				return
			}
			tobj, err := repo.ReadObject(t.HeadId)
			if err != nil {
				loadTemplate(masterTemplate, "error").Execute(w, templates.ErrorTemplateModel{
					ErrorCode: 500,
					ErrorMessage: fmt.Sprintf(
						"Failed to read object %s: %s", t.HeadId, err.Error(),
					),
				})
				return
			}
			if tobj.Type() == gitlib.TAG {
				to := tobj.(*gitlib.TagObject)
				tagInfo = &templates.TagInfoTemplateModel{
					Annotated: true,
					RepoName: rn,
					Tag: to,
				}
				subject, err = repo.ReadObject(to.TaggedObjId)
				if err != nil {
					loadTemplate(masterTemplate, "error").Execute(w, templates.ErrorTemplateModel{
						ErrorCode: 500,
						ErrorMessage: fmt.Sprintf(
							"Failed to read object %s: %s", to.TaggedObjId, err.Error(),
						),
					})
					return
				}
			} else {
				subject = tobj
			}
		}

		if typeStr == "blob" {
			subject, err = repo.ReadObject(nodeName)
		}

		if typeStr == "tree" {
			subject, err = repo.ReadObject(nodeName)
		}

		if typeStr == "commit" {
			subject, err = repo.ReadObject(nodeName)
		}

		if typeStr == "branch" {
			err = repo.SyncAllBranchList()
			if err != nil {
				loadTemplate(masterTemplate, "error").Execute(w, templates.ErrorTemplateModel{
					ErrorCode: 500,
					ErrorMessage: fmt.Sprintf(
						"Failed to sync branch list: %s", err.Error(),
					),
				})
				return
			}
			br, ok := repo.BranchIndex[nodeName]
			if !ok {
				loadTemplate(masterTemplate, "error").Execute(w, templates.ErrorTemplateModel{
					ErrorCode: 404,
					ErrorMessage: fmt.Sprintf(
						"Branch %s not found", nodeName,
					),
				})
				return
			}
			subject, err = repo.ReadObject(br.HeadId)
			if err != nil {
				loadTemplate(masterTemplate, "error").Execute(w, templates.ErrorTemplateModel{
					ErrorCode: 404,
					ErrorMessage: fmt.Sprintf(
						"Failed to read object %s: %s",
						br.HeadId, err.Error(),
					),
				})
				return
			}
		}
		var commitInfo *templates.CommitInfoTemplateModel = nil

		if subject.Type() == gitlib.COMMIT {
			cobj, ok := subject.(*gitlib.CommitObject)
			if !ok {
				loadTemplate(masterTemplate, "error").Execute(w, templates.ErrorTemplateModel{
					ErrorCode: 500,
					ErrorMessage: fmt.Sprintf(
						"Shouldn't happen - object with COMMIT type but not parsed as commit object. ObjId: %s", subject.ObjectId(),
					),
				})
				return
			}
			commitInfo = &templates.CommitInfoTemplateModel{
				RepoName: rn,
				Commit: cobj,
			}
			subject, err = repo.ReadObject(cobj.TreeObjId)
			if err != nil {
				loadTemplate(masterTemplate, "error").Execute(w, templates.ErrorTemplateModel{
					ErrorCode: 500,
					ErrorMessage: fmt.Sprintf(
						"Failed while reading tree object (id: %s)  of commit %s", cobj.TreeObjId, cobj.Id,
					),
				})
				return
			}
		}
		fmt.Println("subj", subject)
		if subject.Type() == gitlib.TAG {
			tobj, ok := subject.(*gitlib.TagObject)
			fmt.Println("tobj", tobj)
			if !ok {
				loadTemplate(masterTemplate, "error").Execute(w, templates.ErrorTemplateModel{
					ErrorCode: 500,
					ErrorMessage: fmt.Sprintf(
						"Shouldn't happen - object with TAG type but not parsed as tag object. ObjId: %s", subject.ObjectId(),
					),
				})
				return
			}
			loadTemplate(masterTemplate, "tag").Execute(w, templates.TagTemplateModel{
				RepoHeaderInfo: repoHeaderInfo,
				Tag: tobj,
				TagInfo: tagInfo,
			})
			return
		} else {
			treePath := r.PathValue("treePath")
			var treePathModelValue *templates.TreePathTemplateModel = nil
			rootFullName := fmt.Sprintf("%s@%s:%s", rn, typeStr, nodeName)
			rootPath := fmt.Sprintf("/repo/%s/%s/%s", rn, typeStr, nodeName)
			if len(treePath) > 0 {
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
					treeobj := subject.(*gitlib.TreeObject)
					fmt.Println(treeobj)
					found := false
					for _, sub := range treeobj.ObjectList {
						if sub.Name == item {
							found = true;
							subject, err = repo.ReadObject(sub.Hash)
							if err != nil {
								loadTemplate(masterTemplate, "error").Execute(w, templates.ErrorTemplateModel{
									ErrorCode: 500,
									ErrorMessage: fmt.Sprintf("Failed to read object %s: %s", sub.Hash, err.Error()),
								})
								return
							}
							break
						}
					}
					if !found {
						loadTemplate(masterTemplate, "error").Execute(w, templates.ErrorTemplateModel{
							ErrorCode: 404,
							ErrorMessage: fmt.Sprintf("Cannot find path %s", treePath),
						})
						return
					}
				}
				treePathModelValue = &templates.TreePathTemplateModel{
					RootFullName: rootFullName,
					RootPath: rootPath,
					TreePath: treePath,
					TreePathSegmentList: treePathSegmentList,
				}
			} else {
				treePathModelValue = &templates.TreePathTemplateModel{
					RootFullName: rootFullName,
					RootPath: rootPath,
					TreePath: treePath,
					TreePathSegmentList: nil,
				}
			}
			fmt.Println(treePathModelValue)
			if gitlib.IsTreeObj(subject) {
				tobj := subject.(*gitlib.TreeObject)
				err = loadTemplate(masterTemplate, "tree").Execute(w, templates.TreeTemplateModel{
					RepoHeaderInfo: repoHeaderInfo,
					TreeFileList: templates.TreeFileListTemplateModel{
						ShouldHaveParentLink: len(treePath) > 0,
						RootPath: fmt.Sprintf("/repo/%s/%s/%s", rn, typeStr, nodeName),
						TreePath: treePath,
						FileList: tobj.ObjectList,
					},
					TreePath: treePathModelValue,
					CommitInfo: commitInfo,
					TagInfo: tagInfo,
				})
				log.Println(err)
				return
			}

			if gitlib.IsBlobObject(subject) {
				bobj := subject.(*gitlib.BlobObject)
				str := string(bobj.Data)
				loadTemplate(masterTemplate, "file-text").Execute(w, templates.FileTextTemplateModel{
					RepoHeaderInfo: repoHeaderInfo,
					File: templates.BlobTextTemplateModel{
						FileLineCount: strings.Count(str, "\n")+1,
						FileContent: str,
					},
					TreePath: treePathModelValue,
					CommitInfo: commitInfo,
					TagInfo: tagInfo,
				})
				return
			}

			if gitlib.IsTagObject(subject) {
				tobj := subject.(*gitlib.TagObject)
				tErr = loadTemplate(masterTemplate, "tag").Execute(w, templates.TagTemplateModel{
					RepoHeaderInfo: repoHeaderInfo,
					Tag: tobj,
					TagInfo: tagInfo,
				})
				logIfError(tErr)
				return
			}

			tErr = loadTemplate(masterTemplate, "error").Execute(w, templates.ErrorTemplateModel{
				ErrorCode: 500,
				ErrorMessage: fmt.Sprintf("Failed to discern object id %s.", subject.ObjectId()),
			})
			logIfError(tErr)
		}
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
		
		h, err := repo.GetCommitHistory(cid)
		check(err)
		his := make([]templates.CommitHistoryItemModel, 0)
		for _, item := range h {
			his = append(his, templates.CommitHistoryItemModel{
				ShortId: item.Id[:6],
				Id: item.Id,
				Message: item.CommitMessage,
				AuthorInfo: item.AuthorInfo,
				CommitterInfo: item.CommitterInfo,
			})
		}
		loadTemplate(masterTemplate, "commit-history").Execute(
			w,
			templates.CommitHistoryModel{
				RepoName: rn,
				RepoDescription: repo.Description,
				TypeStr: typeStr,
				NodeName: string(nodeNameElem[1]),
				CommitHistory: his,
			},
		)
	}))

	var fs = http.FileServer(http.Dir("static/"))
	http.Handle("GET /favicon.ico", withLogHandler(fs))
	http.Handle("GET /static/", http.StripPrefix("/static/", withLogHandler(fs)))

	log.Println("Serve at :8000")
	http.ListenAndServe(":8000", nil)
}

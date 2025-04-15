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
			log.Fatal(err)
		}
	}

	t := loadTemplate(masterTemplate, "index")
	tBranch := loadTemplate(masterTemplate, "branch")
	tRepository := loadTemplate(masterTemplate, "repository")
	tRepositoryNotFound := loadTemplate(masterTemplate, "repository-not-found")

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

	http.HandleFunc("GET /", withLog(func(w http.ResponseWriter, r *http.Request) {
		t.Execute(w, templates.IndexModel{
			RepositoryList: grmodel,
		})
	}))

	http.HandleFunc("GET /repo/{repoName}/obj/{objId}", withLog(func(w http.ResponseWriter, r *http.Request){
		rn := r.PathValue("repoName")
		s, ok := grlist[rn]
		if !ok {
			tRepositoryNotFound.Execute(w, struct{
				RepositoryName string
			}{ RepositoryName: rn })
			return
		}
		obj, err := s.ReadObjectNoResolve(r.PathValue("objId"))
		check(err)
		loadTemplate(masterTemplate, "object").Execute(w, struct{
			ObjectId string
			ObjectType int
			ObjectData []byte
		}{
			ObjectId: obj.ObjectId(),
			ObjectType: obj.Type(),
			ObjectData: obj.RawData(),
		})
	}))

	http.HandleFunc("GET /repo/{repoName}", withLog(func(w http.ResponseWriter, r *http.Request) {
		rn := r.PathValue("repoName")
		s, ok := grlist[rn]
		if !ok {
			tRepositoryNotFound.Execute(w, struct{
				RepositoryName string
			}{ RepositoryName: rn })
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
		rn := r.PathValue("repoName")
		repo, ok := grlist[rn]
		if !ok {
			tRepositoryNotFound.Execute(w, struct{
				RepositoryName string
			}{ RepositoryName: rn })
			return
		}

		typeStr := r.PathValue("typeStr")
		nodeName := r.PathValue("nodeName")
		annotated := false
		taggerInfo := gitlib.AuthorTime{}
		tagMessage := ""
		tagSignature := ""
		
		cid := nodeName
		if typeStr == "branch" {
			err := repo.SyncAllBranchList()
			if err != nil {
				loadTemplate(masterTemplate, "internal-error").Execute(w, templates.InternalErrorModel{
					ErrorCode: 500,
					ErrorMessage: err.Error(),
				})
				return
			}

			check(err)
			br, ok := repo.BranchIndex[nodeName]
			if !ok {
				tRepositoryNotFound.Execute(w, struct{
					RepositoryName string
				}{ RepositoryName: rn })
				return
			}
			cid = br.HeadId
		} else if typeStr == "tag" {
			tl, err := repo.GetAllTagList()
			if err != nil {
				loadTemplate(masterTemplate, "not-found").Execute(w, nil)
				return
			}
			node, ok := tl[nodeName]
			if !ok {
				loadTemplate(masterTemplate, "not-found").Execute(w, nil)
				return
			}
			gobj, err := repo.ReadObject(node.HeadId)
			check(err)
			// NOTE: two types of tag exist: lightweight tag & annotated tag.
			// a lightweight tag is simply a direct reference to a commit,
			// but an annotated tag is an independent object.
			if gobj.Type() == gitlib.TAG {
				annotated = true
				tobj := gobj.(*gitlib.TagObject)
				taggerInfo = tobj.TaggerInfo
				tagMessage = tobj.TagMessage
				tagSignature = tobj.Signature
				cid = tobj.TaggedObjId
			} else if gobj.Type() == gitlib.COMMIT {
				cid = gobj.ObjectId()
			}
		}

		treePath := r.PathValue("treePath")
		tp1 := make([]string, 0)
		treePathSegmentList := make([]struct{Name string;RelPath string}, 0)

		// NOTE: in go, if we strings.Split an empty string with a non-empty
		// separator we will have a slice of length 1 (the item within is
		// of course an empty string). should take this into consideration in
		// future endeavours...
		gobj, err := repo.ReadObject(cid)
		check(err)
		cobj := gobj.(*gitlib.CommitObject)
		subj, err := repo.ReadObject(cobj.TreeObjId)
		check(err)
		for item := range strings.SplitSeq(treePath, "/") {
			if len(item) <= 0 { continue }
			tp1 = append(tp1, item)
			treePathSegmentList = append(treePathSegmentList, struct{
				Name string; RelPath string
			}{
				Name: item, RelPath: strings.Join(tp1, "/"),
			})
			if !gitlib.IsTreeObj(subj) { break }
			treeobj := subj.(*gitlib.TreeObject)
			found := false
			for _, sub := range treeobj.ObjectList {
				if sub.Name == item {
					found = true;
					subj, err = repo.ReadObject(sub.Hash)
					if err != nil {
						loadTemplate(masterTemplate, "internal-error").Execute(w, templates.InternalErrorModel{
							ErrorCode: 500,
							ErrorMessage: err.Error(),
						})
						return
					}
					break
				}
			}
			if !found {
				tRepositoryNotFound.Execute(w, struct{
					RepositoryName string
				}{ RepositoryName: rn })
				return
			}
		}
		resobj, ok := subj.(*gitlib.TreeObject)
		if !ok {
			resobj2, ok := subj.(*gitlib.BlobObject)
			if !ok {
				tRepositoryNotFound.Execute(w, struct{
					RepositoryName string
				}{ RepositoryName: rn })
				return
			}
			loadTemplate(masterTemplate, "branch-single-file").Execute(
				w, templates.BranchSingleFileTemplateModel{
					RepoName: rn,
					RepoDescription: repo.Description,
					TypeStr: typeStr,
					NodeName: nodeName,
					CommitId: cid,
					AuthorInfo: cobj.AuthorInfo,
					CommitterInfo: cobj.CommitterInfo,
					CommitMessage: cobj.CommitMessage,
					CommitSignature: cobj.Signature,
					TreePath: treePath,
					FileContent: string(resobj2.Data),
					FileLineNumber: len(strings.Split(string(resobj2.Data), "\n")),
					TreePathSegmentList: treePathSegmentList,
					Annotated: annotated,
					TaggerInfo: taggerInfo,
					TagMessage: tagMessage,
					TagSignature: tagSignature,
				},
			)
		} else {
			tBranch.Execute(w, templates.BranchTemplateModel{
				RepoName: rn,
				RepoDescription: repo.Description,
				TypeStr: typeStr,
				NodeName: nodeName,
				CommitId: cid,
				AuthorInfo: cobj.AuthorInfo,
				CommitterInfo: cobj.CommitterInfo,
				CommitMessage: cobj.CommitMessage,
				CommitSignature: cobj.Signature,
				TreePath: treePath,
				FileList: resobj.ObjectList,
				TreePathSegmentList: treePathSegmentList,
				Annotated: annotated,
				TaggerInfo: taggerInfo,
				TagMessage: tagMessage,
				TagSignature: tagSignature,
			})
		}
	}))
	
	http.HandleFunc("GET /repo/{repoName}/history/{nodeName}", withLog(func(w http.ResponseWriter, r *http.Request) {
		rn := r.PathValue("repoName")
		repo, ok := grlist[rn]
		if !ok {
			tRepositoryNotFound.Execute(w, struct{
				RepositoryName string
			}{ RepositoryName: rn })
			return
		}

		nodeName := r.PathValue("nodeName")
		nodeNameElem := strings.Split(nodeName, ":")
		typeStr := string(nodeNameElem[0])
		cid := string(nodeNameElem[1])
		if string(nodeNameElem[0]) == "branch" {
			err := repo.SyncAllBranchList()
			check(err)
			br, ok := repo.BranchIndex[string(nodeNameElem[1])]
			if !ok {
				tRepositoryNotFound.Execute(w, struct{
					RepositoryName string
				}{ RepositoryName: rn })
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

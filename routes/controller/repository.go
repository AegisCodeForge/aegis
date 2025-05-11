package controller

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/bctnry/gitus/pkg/gitlib"
	"github.com/bctnry/gitus/routes"
	. "github.com/bctnry/gitus/routes"
	"github.com/bctnry/gitus/templates"
	"github.com/gomarkdown/markdown"
	"github.com/microcosm-cc/bluemonday"
	"github.com/niklasfasching/go-org/org"
)

func bindRepositoryController(ctx *RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/", WithLog(func(w http.ResponseWriter, r *http.Request) {
		rfn := r.PathValue("repoName")
		namespaceName, _, s, err := ctx.ResolveRepositoryFullName(rfn)
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
		
		err = s.Repository.SyncAllBranchList()
		if err != nil {
			LogTemplateError(ctx.LoadTemplate("error").Execute(w, templates.ErrorTemplateModel{
				ErrorCode: 500,
				ErrorMessage: fmt.Sprintf("Failed to sync branch list: %s", err.Error()),
			}))
			return
		}
		err = s.Repository.SyncAllTagList()
		if err != nil {
			LogTemplateError(ctx.LoadTemplate("error").Execute(w, templates.ErrorTemplateModel{
				ErrorCode: 500,
				ErrorMessage: fmt.Sprintf("Failed to sync tag list: %s", err.Error()),
			}))
			return
		}

		readmeString := ""
		readmeType := ""
		// now we try to read the README file.
		// the order would be: README - README.txt
		//                     - any file that starts with "README."
		// the branch order would be: master - main
		// if any of the two cannot be found, it's considered without a readme.
		var obj gitlib.GitObject
		br, ok := s.Repository.BranchIndex["master"]
		if !ok { br, ok = s.Repository.BranchIndex["main"] }
		if !ok { goto findingReadmeDone; }
		obj, err = s.Repository.ReadObject(br.HeadId)
		if err != nil { goto findingReadmeDone; }
		// i don't know if it would ever happen that a branch head would point to
		// anything that's not a commit, but if we can't find it we treat it as
		// no readme.
		if !gitlib.IsCommitObject(obj) { goto findingReadmeDone; }
		obj, err = s.Repository.ReadObject(obj.(*gitlib.CommitObject).TreeObjId)
		if err != nil { goto findingReadmeDone; }
		for _, item := range obj.(*gitlib.TreeObject).ObjectList {
			if item.Name == "README" || item.Name == "README.txt" || strings.HasPrefix(item.Name, "README.") {
				obj, err = s.Repository.ReadObject(item.Hash)
				if err != nil { continue }
				if !gitlib.IsBlobObject(obj) { continue }
				obj, err = s.Repository.ReadObject(item.Hash)
				readmeType = path.Ext(item.Name)
				readmeString = string(obj.(*gitlib.BlobObject).Data)
				goto renderReadme
			}
		}
		
	renderReadme:
		switch readmeType {
		case ".md":
			// NOTE: markdown is tricky. because people uses html in
			// markdown file for things markdown does not support
			// (e.g. <detail>) so it's a good idea to allow them
			// instead of escaping all html (and also we're going to
			// embed raw html anyway doe to how we currently do
			// things), but if we allow arbitrary html then people are
			// gonna do XSS attacks. i'd like to keep dependencies as
			// minimal as possible, but i'm not ready to make a whole
			// markdown-to-html renderer, at least not yet.
			readmeString = string(markdown.ToHTML([]byte(readmeString), nil, nil))
			readmeString = bluemonday.UGCPolicy().Sanitize(readmeString)
		case ".org":
			// NOTE: due to go-org having no documentations and I can't see
			// a way to inject prefix into the rendering process, we might
			// have to either make a better form or create an org-mode
			// parser on our own.
			doc := org.New().Parse(strings.NewReader(readmeString), "")
			out, err := doc.Write(org.NewHTMLWriter())
			if err != nil {
				readmeString = bluemonday.UGCPolicy().Sanitize(readmeString)
				readmeString = fmt.Sprintf("<pre>%s</pre>", readmeString)
			} else {
				readmeString = bluemonday.UGCPolicy().Sanitize(out)
			}
		default:
			readmeString = bluemonday.UGCPolicy().Sanitize(readmeString)
			readmeString = fmt.Sprintf("<pre>%s</pre>", readmeString)
		}
		
	findingReadmeDone:

		LogTemplateError(ctx.LoadTemplate("repository").Execute(w, templates.RepositoryModel{
			RepoName: rfn,
			RepoObj: s.Repository,
			RepoHeaderInfo: templates.RepoHeaderTemplateModel{
				NamespaceName: namespaceName,
				RepoName: rfn,
				TypeStr: "",
				NodeName: "",
				RepoDescription: s.Description,
				RepoLabelList: nil,
				RepoURL: fmt.Sprintf("%s/repo/%s", ctx.Config.HttpHostName, rfn),
			},
			BranchList: s.Repository.BranchIndex,
			TagList: s.Repository.TagIndex,
			ReadmeString: readmeString,
		}))
	}))
}


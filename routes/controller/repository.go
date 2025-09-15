package controller

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/pkg/gitlib"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
	"github.com/gomarkdown/markdown"
	"github.com/microcosm-cc/bluemonday"
	"github.com/niklasfasching/go-org/org"
)

func bindRepositoryController(ctx *RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/", UseMiddleware(
		[]Middleware{Logged, ValidRepositoryNameRequired("repoName"),
			UseLoginInfo, GlobalVisibility, ErrorGuard,
		}, ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			rfn := r.PathValue("repoName")
			branch := strings.TrimSpace(r.URL.Query().Get("branch"))
			if len(branch) > 0 {
				k := strings.SplitN(branch, ":", 2)
				switch k[0] {
				case "branch":
					FoundAt(w, fmt.Sprintf("/repo/%s/branch/%s", rfn, k[1]))
					return
				case "tag":
					FoundAt(w, fmt.Sprintf("/repo/%s/tag/%s", rfn, k[1]))
					return
				}
			}
			_, _, ns, s, err := rc.ResolveRepositoryFullName(rfn)
			if err == ErrNotFound {
				rc.ReportNotFound(rfn, "Repository", "Depot", w, r)
				return
			}
			if err != nil {
				rc.ReportInternalError(err.Error(), w, r)
				return
			}
			
			if !rc.Config.PlainMode {
				rc.LoginInfo.IsOwner = s.Owner == rc.LoginInfo.UserName || ns.Owner == rc.LoginInfo.UserName
			}
			if !rc.Config.PlainMode && s.Status == model.REPO_NORMAL_PRIVATE {
				t := s.AccessControlList.GetUserPrivilege(rc.LoginInfo.UserName)
				if t == nil {
					t = ns.ACL.GetUserPrivilege(rc.LoginInfo.UserName)
				}
				if t == nil {
					rc.ReportForbidden("Not enough privilege", w, r)
					return
				}
			}

			rr := s.Repository.(*gitlib.LocalGitRepository)
			err = rr.SyncAllBranchList()
			if err != nil {
				LogTemplateError(rc.LoadTemplate("error").Execute(w, templates.ErrorTemplateModel{
					ErrorCode: 500,
					ErrorMessage: fmt.Sprintf("Failed to sync branch list: %s", err.Error()),
				}))
				return
			}
			err = rr.SyncAllTagList()
			if err != nil {
				LogTemplateError(rc.LoadTemplate("error").Execute(w, templates.ErrorTemplateModel{
					ErrorCode: 500,
					ErrorMessage: fmt.Sprintf("Failed to sync tag list: %s", err.Error()),
				}))
				return
			}

			readmeString := ""
			readmeType := ""
			readmeTier := 0
			// now we try to read the README file.
			// the order would be: README - README.txt
			//                     - any file that starts with "README."
			// the branch order would be: master - main
			// if any of the two cannot be found, it's considered without a readme.
			var obj gitlib.GitObject
			br, ok := rr.BranchIndex["master"]
			if !ok { br, ok = rr.BranchIndex["main"] }
			if !ok { goto findingReadmeDone; }
			obj, err = rr.ReadObject(br.HeadId)
			if err != nil { goto findingReadmeDone; }
			// i don't know if it would ever happen that a branch head would point to
			// anything that's not a commit, but if we can't find it we treat it as
			// no readme.
			if !gitlib.IsCommitObject(obj) { goto findingReadmeDone; }
			obj, err = rr.ReadObject(obj.(*gitlib.CommitObject).TreeObjId)
			if err != nil { goto findingReadmeDone; }
			for _, item := range obj.(*gitlib.TreeObject).ObjectList {
				if item.Name == "README" || strings.HasPrefix(item.Name, "README.") {
					// NOTE: this is to make sure that README.md and the like will
					// always have a higher priority than other README file; some repo
					// put platform-specific README in files like `README.{plat}.md`
					// and you can't have them getting selected when a more general
					// readme exists.
					thisTier := 2
					if item.Name == "README" || item.Name == "README.txt" || item.Name == "README.org" || item.Name == "README.md" { thisTier = 1 }
					if readmeTier > 0 && thisTier > readmeTier { continue }
					readmeTier = thisTier
					obj, err = rr.ReadObject(item.Hash)
					if err != nil { continue }
					if !gitlib.IsBlobObject(obj) { continue }
					obj, err = rr.ReadObject(item.Hash)
					readmeType = path.Ext(item.Name)
					readmeString = string(obj.(*gitlib.BlobObject).Data)
					if thisTier == 1 { goto renderReadme }
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
					readmeString = fmt.Sprintf("<pre class=\"repo-readme\">>%s</pre>", readmeString)
				} else {
					readmeString = bluemonday.UGCPolicy().Sanitize(out)
				}
			default:
				readmeString = bluemonday.UGCPolicy().Sanitize(readmeString)
				readmeString = fmt.Sprintf("<pre class=\"repo-readme\">%s</pre>", readmeString)
			}
			
		findingReadmeDone:

			repoHeaderInfo := GenerateRepoHeader("", "")

			LogTemplateError(rc.LoadTemplate("repository").Execute(w, templates.RepositoryModel{
				Config: rc.Config,
				Repository: s,
				RepoHeaderInfo: *repoHeaderInfo,
				BranchList: rr.BranchIndex,
				TagList: rr.TagIndex,
				ReadmeString: readmeString,
				LoginInfo: rc.LoginInfo,
			}))
		},
	))

	if !ctx.Config.PlainMode {
		bindRepositoryForkController(ctx)
		bindRepositoryPullRequestController(ctx)
	}
}


package controller

import (
	"fmt"
	"net/http"
	"path"
	"strings"

	"github.com/bctnry/aegis/pkg/aegis/db"
	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/pkg/gitlib"
	"github.com/bctnry/aegis/routes"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
	"github.com/gomarkdown/markdown"
	"github.com/microcosm-cc/bluemonday"
	"github.com/niklasfasching/go-org/org"
)

func bindRepositoryController(ctx *RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/", WithLog(func(w http.ResponseWriter, r *http.Request) {
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
		_, _, ns, s, err := ctx.ResolveRepositoryFullName(rfn)
		if err == routes.ErrNotFound {
			ctx.ReportNotFound(rfn, "Repository", "Depot", w, r)
			return
		}
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		
		var loginInfo *templates.LoginInfoModel = nil
		if !ctx.Config.PlainMode {
			loginInfo, err = GenerateLoginInfoModel(ctx, r)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			loginInfo.IsOwner = s.Owner == loginInfo.UserName || ns.Owner == loginInfo.UserName
		}
		if !ctx.Config.PlainMode && s.Status == model.REPO_NORMAL_PRIVATE {
			t := s.AccessControlList.GetUserPrivilege(loginInfo.UserName)
			if t == nil {
				t = ns.ACL.GetUserPrivilege(loginInfo.UserName)
			}
			if t == nil {
				ctx.ReportForbidden("Not enough privilege", w, r)
				return
			}
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
		readmeTier := 0
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
				obj, err = s.Repository.ReadObject(item.Hash)
				if err != nil { continue }
				if !gitlib.IsBlobObject(obj) { continue }
				obj, err = s.Repository.ReadObject(item.Hash)
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

		LogTemplateError(ctx.LoadTemplate("repository").Execute(w, templates.RepositoryModel{
			Config: ctx.Config,
			Repository: s,
			RepoHeaderInfo: *repoHeaderInfo,
			BranchList: s.Repository.BranchIndex,
			TagList: s.Repository.TagIndex,
			ReadmeString: readmeString,
			LoginInfo: loginInfo,
		}))
	}))

	http.HandleFunc("GET /repo/{repoName}/fork", WithLog(func(w http.ResponseWriter, r *http.Request) {
		rfn := r.PathValue("repoName")
		if ctx.Config.PlainMode {
			FoundAt(w, fmt.Sprintf("/repo/%s", rfn))
			return
		}
		_, _, _, s, err := ctx.ResolveRepositoryFullName(rfn)
		if err == routes.ErrNotFound {
			ctx.ReportNotFound(rfn, "Repository", "Depot", w, r)
			return
		}
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		l, err := ctx.DatabaseInterface.GetAllComprisingNamespace(loginInfo.UserName)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		for k := range l {
			if !loginInfo.IsAdmin && l[k].Owner != loginInfo.UserName {
				acl := l[k].ACL.GetUserPrivilege(loginInfo.UserName)
				if acl == nil || !acl.AddRepository {
					delete(l, k)
				}
			}
		}
		LogTemplateError(ctx.LoadTemplate("fork").Execute(w, templates.ForkTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			SourceRepository: s,
			NamespaceList: l,
		}))
	}))
	
	http.HandleFunc("POST /repo/{repoName}/fork", WithLog(func(w http.ResponseWriter, r *http.Request) {
		rfn := r.PathValue("repoName")
		if ctx.Config.PlainMode {
			FoundAt(w, fmt.Sprintf("/repo/%s", rfn))
			return
		}
		originNs, originName, _, _, err := ctx.ResolveRepositoryFullName(rfn)
		if err == routes.ErrNotFound {
			ctx.ReportNotFound(rfn, "Repository", "Depot", w, r)
			return
		}
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		err = r.ParseForm()
		if err != nil {
			ctx.ReportNormalError("Invalid request", w, r)
			return
		}
		namespace := r.Form.Get("namespace")
		name := r.Form.Get("name")
		rp, err := ctx.DatabaseInterface.SetUpCloneRepository(originNs, originName, namespace, name, loginInfo.UserName)
		if err == db.ErrEntityAlreadyExists {
			ctx.ReportRedirect(fmt.Sprintf("/repo/%s/fork", rfn), 0, "Already Exists", fmt.Sprintf("The repository %s:%s already exists. Please choose a different name or namespace.", namespace, name), w, r)
			return
		}
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		FoundAt(w, fmt.Sprintf("/repo/%s", rp.FullName()))
	}))
}


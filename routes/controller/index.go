package controller

import (
	"fmt"
	"html"
	"net/http"
	"strings"

	"github.com/bctnry/aegis/pkg/aegis"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
	"github.com/gomarkdown/markdown"
	"github.com/microcosm-cc/bluemonday"
	"github.com/niklasfasching/go-org/org"
)

func bindIndexController(ctx *RouterContext) {
	http.HandleFunc("GET /", WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if !CheckGlobalVisibleToUser(ctx, loginInfo) {
			switch ctx.Config.GlobalVisibility {
			case aegis.GLOBAL_VISIBILITY_MAINTENANCE: FoundAt(w, "/maintenance-notice")
			case aegis.GLOBAL_VISIBILITY_SHUTDOWN: FoundAt(w, "/shutdown-notice")
			case aegis.GLOBAL_VISIBILITY_PRIVATE: FoundAt(w, "/login")
			}
			return
		}
		if strings.HasPrefix(ctx.Config.FrontPageType, "static/") {
			frontPageContentType := ctx.Config.FrontPageType[len("static/"):]
			f := ctx.Config.FrontPageContent
			var frontPageHtml string
			switch frontPageContentType {
			case "text":
				frontPageHtml = fmt.Sprintf("<pre>%s</pre>", html.EscapeString(f))
			case "markdown":
				frontPageHtml = string(markdown.ToHTML([]byte(f), nil, nil))
			case "org":
				out, err := org.New().Parse(strings.NewReader(f), "").Write(org.NewHTMLWriter())
				if err != nil {
					frontPageHtml = fmt.Sprintf("<pre>%s</pre>", f)
				} else {
					frontPageHtml = out
				}
			case "html":
				frontPageHtml = f
			}
			frontPageHtml = bluemonday.UGCPolicy().Sanitize(frontPageHtml)
			LogTemplateError(ctx.LoadTemplate("index-static").Execute(w, templates.IndexStaticTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				FrontPage: frontPageHtml,
			}))
			return
		} else if ctx.Config.FrontPageType == "all/namespace" {
			FoundAt(w, "/all/namespace")
		} else if ctx.Config.FrontPageType == "all/repository" {
			FoundAt(w, "/all/repo")
		} else if strings.HasPrefix(ctx.Config.FrontPageType, "namespace/") {
			if !ctx.Config.UseNamespace {
				frontPageHtml := "<p>Misconfiguration: a namespace is used for the front page, but the depot itself is configured to not support namespaces. Please contact the site owner about this issue.</p>"
				LogTemplateError(ctx.LoadTemplate("index-static").Execute(w, templates.IndexStaticTemplateModel{
					Config: ctx.Config,
					LoginInfo: loginInfo,
					FrontPage: frontPageHtml,
				}))
				return
			}
			FoundAt(w, fmt.Sprintf("/s/%s", ctx.Config.FrontPageType[len("namespace/"):]))
		} else if strings.HasPrefix(ctx.Config.FrontPageType, "repository/") {
			FoundAt(w, fmt.Sprintf("/repo/%s", ctx.Config.FrontPageType[len("repository/"):]))
		} else {
			if ctx.Config.UseNamespace {
				FoundAt(w, "/all/namespace")
			} else {
				FoundAt(w, "/all/repo")
			}
		}
	}))
}


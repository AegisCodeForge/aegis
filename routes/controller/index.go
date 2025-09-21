package controller

import (
	"fmt"
	"html"
	"net/http"
	"strings"

	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
	"github.com/gomarkdown/markdown"
	"github.com/microcosm-cc/bluemonday"
	"github.com/niklasfasching/go-org/org"
)

func bindIndexController(ctx *RouterContext) {
	http.HandleFunc("GET /", UseMiddleware(
		[]Middleware{Logged, UseLoginInfo, GlobalVisibility, ErrorGuard}, ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			fmt.Println(r.Header.Get("Host"))
			fmt.Println(r.URL.Path)
				
			if strings.HasPrefix(rc.Config.FrontPageType, "static/") {
				frontPageContentType := rc.Config.FrontPageType[len("static/"):]
				f := rc.Config.FrontPageContent
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
				LogTemplateError(rc.LoadTemplate("index-static").Execute(w, templates.IndexStaticTemplateModel{
					Config: rc.Config,
					LoginInfo: rc.LoginInfo,
					FrontPage: frontPageHtml,
				}))
				return
			} else if rc.Config.FrontPageType == "all/namespace" {
				FoundAt(w, "/all/namespace")
			} else if rc.Config.FrontPageType == "all/repository" {
				FoundAt(w, "/all/repo")
			} else if strings.HasPrefix(rc.Config.FrontPageType, "namespace/") {
				if !rc.Config.UseNamespace {
					frontPageHtml := "<p>Misconfiguration: a namespace is used for the front page, but the depot itself is configured to not support namespaces. Please contact the site owner about this issue.</p>"
					LogTemplateError(rc.LoadTemplate("index-static").Execute(w, templates.IndexStaticTemplateModel{
						Config: rc.Config,
						LoginInfo: rc.LoginInfo,
						FrontPage: frontPageHtml,
					}))
					return
				}
				FoundAt(w, fmt.Sprintf("/s/%s", rc.Config.FrontPageType[len("namespace/"):]))
			} else if strings.HasPrefix(rc.Config.FrontPageType, "repository/") {
				FoundAt(w, fmt.Sprintf("/repo/%s", rc.Config.FrontPageType[len("repository/"):]))
			} else {
				if rc.Config.UseNamespace {
					FoundAt(w, "/all/namespace")
				} else {
					FoundAt(w, "/all/repo")
				}
			}
		},
	))
}


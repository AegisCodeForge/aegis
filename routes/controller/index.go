package controller

import (
	"fmt"
	"html"
	"net/http"
	"os"
	"path"
	"strings"

	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
	"github.com/gomarkdown/markdown"
	"github.com/microcosm-cc/bluemonday"
)

func bindIndexController(ctx *RouterContext) {
	http.HandleFunc("GET /", WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		if strings.HasPrefix(ctx.Config.FrontPageConfig, "/") {
			p := path.Join(ctx.Config.StaticAssetDirectory, ctx.Config.FrontPageConfig[1:])
			var frontPageHtml string
			f, err := os.ReadFile(p)
			if err != nil {
				frontPageHtml = fmt.Sprintf("<p>Failed to read preset index page: %s. Please contact the site owner about this.</p>", err.Error())
			} else {
				switch path.Ext(p) {
				case ".txt":
					frontPageHtml = fmt.Sprintf("<pre>%s</pre>", html.EscapeString(string(f)))
				case ".md":
					rs := string(markdown.ToHTML(f, nil, nil))
					rs = bluemonday.UGCPolicy().Sanitize(rs)
					frontPageHtml = rs
				case ".htm": fallthrough
				case ".html":
					frontPageHtml = bluemonday.UGCPolicy().Sanitize(string(f))
				}
			}
			LogTemplateError(ctx.LoadTemplate("index-static").Execute(w, templates.IndexStaticTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				FrontPage: frontPageHtml,
			}))
			return
		} else if ctx.Config.FrontPageConfig == "all/namespace" {
			FoundAt(w, "/all/namespace")
		} else if ctx.Config.FrontPageConfig == "all/repository" {
			FoundAt(w, "/all/repo")
		} else if strings.HasPrefix(ctx.Config.FrontPageConfig, "namespace/") {
			if !ctx.Config.UseNamespace {
				frontPageHtml := "<p>Misconfiguration: a namespace is used for the front page, but the depot itself is configured to not support namespaces. Please contact the site owner about this issue.</p>"
				LogTemplateError(ctx.LoadTemplate("index-static").Execute(w, templates.IndexStaticTemplateModel{
					Config: ctx.Config,
					LoginInfo: loginInfo,
					FrontPage: frontPageHtml,
				}))
				return
			}
			FoundAt(w, fmt.Sprintf("/s/%s", ctx.Config.FrontPageConfig[len("namespace/"):]))
		} else if strings.HasPrefix(ctx.Config.FrontPageConfig, "repository/") {
			FoundAt(w, fmt.Sprintf("/repo/%s", ctx.Config.FrontPageConfig[len("repository/"):]))
		} else {
			FoundAt(w, "/all/namespace")
		}
	}))
}


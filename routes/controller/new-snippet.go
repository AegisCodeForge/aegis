package controller

import (
	"fmt"
	"net/http"
	"strconv"

	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindNewSnippetController(ctx *RouterContext) {
	http.HandleFunc("GET /new/snippet", UseMiddleware(
		[]Middleware{Logged, LoginRequired, GlobalVisibility, ErrorGuard}, ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			LogTemplateError(ctx.LoadTemplate("new/snippet").Execute(w, templates.NewRepositoryTemplateModel{
				Config: rc.Config,
				LoginInfo: rc.LoginInfo,
			}))
		},
	))
	http.HandleFunc("POST /new/snippet", UseMiddleware(
		[]Middleware{Logged, ValidPOSTRequestRequired,
			UseLoginInfo, LoginRequired,
			GlobalVisibility, ErrorGuard,
		}, ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			if err != nil {
				rc.ReportNormalError("Invalid request", w, r)
				return
			}
			username := rc.LoginInfo.UserName
			name := r.Form.Get("name")
			filename := r.Form.Get("filename")
			statusStr := r.Form.Get("status")
			status, err := strconv.Atoi(statusStr)
			if err != nil {
				rc.ReportNormalError("Invalid request", w, r)
				return
			}
			content := r.Form.Get("content")
			sn, err := rc.DatabaseInterface.NewSnippet(username, name, uint8(status))
			if err != nil {
				rc.ReportInternalError(fmt.Sprintf("Failed to create new snippet: %s", err), w, r)
				return
			}
			if sn.FileList == nil {
				sn.FileList = make(map[string]string, 0)
			}
			sn.FileList[filename] = content
			err = sn.SyncFile(rc.Config.SnippetRoot, filename)
			if err != nil {
				rc.ReportInternalError(fmt.Sprintf("Failed to create new snippet: %s", err), w, r)
				return
			}
			rc.ReportRedirect(fmt.Sprintf("/snippet/%s/%s", username, name), 5, "Snippet Created", "Snippet created.", w, r)
		},
	))
}


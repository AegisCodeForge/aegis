package admin

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindAdminIndexConfigController(ctx *RouterContext) {
	http.HandleFunc("GET /admin/index-config", UseMiddleware(
		[]Middleware{Logged, LoginRequired, AdminRequired,
			GlobalVisibility, ErrorGuard,
		}, ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			iType := rc.Config.FrontPageType
			var namespace, repository, fileContent string
			if strings.HasPrefix(iType, "static/") {
				fileContent = rc.Config.FrontPageContent
			} else if strings.HasPrefix(iType, "namespace/") {
				namespace = iType[len("namespace/"):]
				iType = "namespace"
			} else if strings.HasPrefix(iType, "repository/") {
				k := strings.Split(iType, ":")
				if len(k) <= 1 {
					namespace = ""
					repository = k[0]
				} else {
					namespace = k[0]
					repository = k[1]
				}
				iType = "repository"
			}
			LogTemplateError(rc.LoadTemplate("admin/index-config").Execute(w, &templates.AdminIndexConfigTemplateModel{
				Config: rc.Config,
				LoginInfo: rc.LoginInfo,
				ErrorMsg: "",
				IndexType: iType,
				IndexNamespace: namespace,
				IndexRepository: repository,
				IndexFileContent: fileContent,
			}))
		},
	))
	
	http.HandleFunc("POST /admin/index-config", UseMiddleware(
		[]Middleware{Logged, ValidPOSTRequestRequired,
			LoginRequired, AdminRequired,
			GlobalVisibility, ErrorGuard,
		}, ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			indexType := r.Form.Get("index-type")
			namespace := r.Form.Get("namespace")
			repository := r.Form.Get("repository")
			fileName := r.Form.Get("file-name")
			if !strings.HasPrefix(fileName, "/") {
				fileName = "/" + fileName
			}
			fileContent := r.Form.Get("file-content")
			rc.Config.LockForSync()
			defer rc.Config.Unlock()
			switch indexType {
			case "all/namespace":
				rc.Config.FrontPageType = "all/namespace"
			case "all/repository":
				rc.Config.FrontPageType = "all/repository"
			case "namespace":
				rc.Config.FrontPageType = fmt.Sprintf("namespace/%s", namespace)
			case "repository":
				rc.Config.FrontPageType = fmt.Sprintf("repository/%s:%s", namespace, repository)
			case "static":
				rc.Config.FrontPageType = fileName
				p := path.Join(rc.Config.StaticAssetDirectory, fileName)
				f, err := os.OpenFile(p, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0664)
				if err != nil {
					rc.ReportInternalError(fmt.Sprintf("Failed to open static file: %s", err), w, r)
					return
				}
				defer f.Close()
				_, err = f.Write([]byte(fileContent))
				if err != nil {
					rc.ReportInternalError(fmt.Sprintf("Failed to save static file: %s", err), w, r)
					return
				}
				err = rc.Config.Sync()
				if err != nil {
					rc.ReportInternalError(fmt.Sprintf("Failed to save config: %s", err), w, r)
					return
				}
				
				LogTemplateError(rc.LoadTemplate("admin/index-config").Execute(w, templates.AdminIndexConfigTemplateModel{
					Config: rc.Config,
					LoginInfo: rc.LoginInfo,
					ErrorMsg: "Updated.",
					IndexType: indexType,
					IndexNamespace: namespace,
					IndexRepository: repository,
					IndexFileContent: fileContent,
				}))
			}
		},
	))
}


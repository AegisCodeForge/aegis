package admin

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/bctnry/aegis/pkg/aegis/model"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

// /admin/namespace-list?p={pagenum}&s={pagesize}&q={query}
func bindAdminNamespaceListController(ctx *RouterContext) {
	http.HandleFunc("GET /admin/namespace-list", UseMiddleware(
		[]Middleware{Logged, LoginRequired, AdminRequired,
			GlobalVisibility, ErrorGuard,
		}, ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			i, err := rc.DatabaseInterface.CountAllUser()
			p := r.URL.Query().Get("p")
			if len(p) <= 0 { p = "1" }
			s := r.URL.Query().Get("s")
			if len(s) <= 0 { s = "50" }
			q := strings.TrimSpace(r.URL.Query().Get("q"))
			pageNum, err := strconv.ParseInt(p, 10, 32)
			pageSize, err := strconv.ParseInt(s, 10, 32)
			totalPage := i / pageSize
			if i % pageSize != 0 { totalPage += 1 }
			if pageNum > totalPage { pageNum = totalPage }
			if pageNum <= 1 { pageNum = 1 }
			var namespaceList map[string]*model.Namespace
			if len(q) > 0 {
				namespaceList, err = rc.DatabaseInterface.SearchForNamespace(q, int(pageNum-1), int(pageSize))
			} else {
				namespaceList, err = rc.DatabaseInterface.GetAllNamespaces(int(pageNum-1), int(pageSize))
			}
			if err != nil {
				rc.ReportInternalError(fmt.Sprintf("Failed to load namespace list: %s", err), w, r)
				return
			}
			LogTemplateError(rc.LoadTemplate("admin/namespace-list").Execute(w, &templates.AdminNamespaceListTemplateModel{
				Config: rc.Config,
				LoginInfo: rc.LoginInfo,
				ErrorMsg: "",
				NamespaceList: namespaceList,
				Query: q,
				PageInfo: &templates.PageInfoModel{
					PageNum: int(pageNum),
					PageSize: int(pageSize),
					TotalPage: int(totalPage),
				},
			}))
			
		},
	))

	http.HandleFunc("GET /admin/namespace/{name}/delete", UseMiddleware(
		[]Middleware{Logged, LoginRequired, AdminRequired,
			GlobalVisibility, ErrorGuard,
		}, ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			nsn := r.PathValue("name")
			if !model.ValidUserName(nsn) { FoundAt(w, "/") }
			err := rc.DatabaseInterface.HardDeleteNamespaceByName(nsn)
			if err != nil {
				rc.ReportInternalError(fmt.Sprintf("Failed to delete namespace by name: %s", err), w, r)
				return
			}
			FoundAt(w, "/admin/namespace-list")
		},
	))
}


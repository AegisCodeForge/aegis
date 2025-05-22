package admin

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

// /admin/namespace-list?p={pagenum}&s={pagesize}&q={query}
func bindAdminNamespaceListController(ctx *routes.RouterContext) {
	http.HandleFunc("GET /admin/namespace-list", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		i, err := ctx.DatabaseInterface.CountAllUser()
		p := r.URL.Query().Get("p")
		if len(p) <= 0 { p = "1" }
		s := r.URL.Query().Get("s")
		if len(s) <= 0 { s = "50" }
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		pageNum, err := strconv.ParseInt(p, 10, 32)
		pageSize, err := strconv.ParseInt(s, 10, 32)
		totalPage := i / pageSize
		fmt.Println(pageNum, pageSize, totalPage)
		if i % pageSize != 0 { totalPage += 1 }
		if pageNum > totalPage { pageNum = totalPage }
		if pageNum <= 1 { pageNum = 1 }
		var namespaceList map[string]*model.Namespace
		if len(q) > 0 {
			namespaceList, err = ctx.DatabaseInterface.SearchForNamespace(q, int(pageNum-1), int(pageSize))
		} else {
			namespaceList, err = ctx.DatabaseInterface.GetAllNamespaces(int(pageNum-1), int(pageSize))
		}
		if err != nil {
			routes.LogTemplateError(ctx.LoadTemplate("admin/namespace-list").Execute(w, &templates.AdminNamespaceListTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: fmt.Sprintf("Failed to load namespace list: %s", err.Error()),
				NamespaceList: nil,
			}))
			return
		}
		routes.LogTemplateError(ctx.LoadTemplate("admin/namespace-list").Execute(w, &templates.AdminNamespaceListTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			ErrorMsg: "",
			NamespaceList: namespaceList,
			PageInfo: &templates.PageInfoModel{
				PageNum: int(pageNum),
				PageSize: int(pageSize),
				TotalPage: int(totalPage),
			},
		}))
	}))

	http.HandleFunc("GET /admin/namespace/{name}/delete", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		nsn := r.PathValue("name")
		err = ctx.DatabaseInterface.HardDeleteNamespaceByName(nsn)
		if err != nil {
			routes.LogTemplateError(ctx.LoadTemplate("admin/_redirect-with-message").Execute(w, &templates.AdminRedirectWithMessageModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: "",
				Timeout: 3,
				RedirectUrl: "/admin/namespace-list",
				MessageTitle: "Error",
				MessageText: fmt.Sprintf("Failed to hard delete namespace: %s", err.Error()),
			}))
			return
		}
		routes.FoundAt(w, "/admin/namespace-list")
	}))
}


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

// /admin/repo-list?p={pagenum}&s={pagesize}&q={query}
func bindAdminRepositoryListController(ctx *routes.RouterContext) {
	http.HandleFunc("GET /admin/repo-list", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
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
		var repoList []*model.Repository
		if len(q) > 0 {
			repoList, err = ctx.DatabaseInterface.SearchForRepository(q, int(pageNum-1), int(pageSize))
		} else {
			repoList, err = ctx.DatabaseInterface.GetAllRepositories(int(pageNum-1), int(pageSize))
		}
		if err != nil {
			routes.LogTemplateError(ctx.LoadTemplate("admin/repo-list").Execute(w, &templates.AdminRepositoryListTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				Query: q,
				ErrorMsg: fmt.Sprintf("Failed to load repository list: %s", err.Error()),
				RepositoryList: nil,
			}))
			return
		}
		routes.LogTemplateError(ctx.LoadTemplate("admin/repo-list").Execute(w, &templates.AdminRepositoryListTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			ErrorMsg: "",
			RepositoryList: repoList,
			Query: q,
			PageInfo: &templates.PageInfoModel{
				PageNum: int(pageNum),
				PageSize: int(pageSize),
				TotalPage: int(totalPage),
			},
		}))
	}))

}


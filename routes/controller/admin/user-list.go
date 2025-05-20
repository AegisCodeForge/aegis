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

// /admin/user-list?p={pagenum}&s={pagesize}
func bindAdminUserListController(ctx *routes.RouterContext) {
	http.HandleFunc("GET /admin/user-list", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
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
		if i % pageSize != 0 { totalPage += 1 }
		if pageNum > totalPage { pageNum = totalPage }
		if pageNum <= 1 { pageNum = 1 }
		var userList []*model.AegisUser
		if len(q) > 0 {
			userList, err = ctx.DatabaseInterface.SearchForUser(q, int(pageNum-1), int(pageSize))
		} else {
			userList, err = ctx.DatabaseInterface.GetAllUsers(int(pageNum-1), int(pageSize))
		}
		if err != nil {
			userList := make([]*model.AegisUser, 0)
			routes.LogTemplateError(ctx.LoadTemplate("admin/user-list").Execute(w, &templates.AdminUserListTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: fmt.Sprintf("Failed to load user list: %s", err.Error()),
				UserList: userList,
			}))
			return
		}
		routes.LogTemplateError(ctx.LoadTemplate("admin/user-list").Execute(w, &templates.AdminUserListTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			ErrorMsg: "",
			UserList: userList,
			PageInfo: &templates.PageInfoModel{
				PageNum: int(pageNum),
				PageSize: int(pageSize),
				TotalPage: int(totalPage),
			},
		}))
	}))

	http.HandleFunc("GET /admin/user/{username}/delete", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		un := r.PathValue("username")
		u, err := ctx.DatabaseInterface.GetUserByName(un)
		if !loginInfo.IsSuperAdmin && u.Status == model.SUPER_ADMIN {
			routes.LogTemplateError(ctx.LoadTemplate("admin/_redirect-with-message").Execute(w, &templates.AdminRedirectWithMessageModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: "",
				Timeout: 3,
				RedirectUrl: "/admin/user-list",
				MessageTitle: "Error",
				MessageText: "Not enough permission.",
			}))
			return
		}
		err = ctx.DatabaseInterface.HardDeleteUserByName(un)
		if err != nil {
			routes.LogTemplateError(ctx.LoadTemplate("admin/_redirect-with-message").Execute(w, &templates.AdminRedirectWithMessageModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: "",
				Timeout: 3,
				RedirectUrl: "/admin/user-list",
				MessageTitle: "Error",
				MessageText: fmt.Sprintf("Failed to hard delete user: %s", err.Error()),
			}))
			return
		}
		routes.FoundAt(w, "/admin/user-list")
	}))
}


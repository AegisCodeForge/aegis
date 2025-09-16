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

// /admin/user-list?p={pagenum}&s={pagesize}
func bindAdminUserListController(ctx *RouterContext) {
	http.HandleFunc("GET /admin/user-list", UseMiddleware(
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
			var userList []*model.AegisUser
			if len(q) > 0 {
				userList, err = rc.DatabaseInterface.SearchForUser(q, int(pageNum-1), int(pageSize))
			} else {
				userList, err = rc.DatabaseInterface.GetAllUsers(int(pageNum-1), int(pageSize))
			}
			if err != nil {
				rc.ReportInternalError(fmt.Sprintf("Failed to load users: %s", err), w, r)
				return
			}
			LogTemplateError(rc.LoadTemplate("admin/user-list").Execute(w, &templates.AdminUserListTemplateModel{
				Config: rc.Config,
				LoginInfo: rc.LoginInfo,
				ErrorMsg: "",
				UserList: userList,
				PageInfo: &templates.PageInfoModel{
					PageNum: int(pageNum),
					PageSize: int(pageSize),
					TotalPage: int(totalPage),
				},
			}))
		},
	))

	http.HandleFunc("GET /admin/user/{username}/delete", UseMiddleware(
		[]Middleware{Logged, LoginRequired, AdminRequired,
			GlobalVisibility, ErrorGuard,
		}, ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			un := r.PathValue("username")
			u, err := rc.DatabaseInterface.GetUserByName(un)
			if !rc.LoginInfo.IsSuperAdmin && u.Status == model.SUPER_ADMIN {
				rc.ReportRedirect("/admin/user-list", 3, "Not Enough Privilege", "Your account does not have enough privilege to perform this action.", w, r)
				return
			}
			err = rc.DatabaseInterface.HardDeleteUserByName(un)
			if err != nil {
				rc.ReportInternalError(fmt.Sprintf("Failed to delete user: %s", err), w, r)
				return
			}
			FoundAt(w, "/admin/user-list")
		},
	))
}


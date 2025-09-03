package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/bctnry/aegis/pkg/aegis/db"
	"github.com/bctnry/aegis/pkg/aegis/model"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)


func bindUserController(ctx *RouterContext) {
	http.HandleFunc("GET /u/{userName}", UseMiddleware(
		[]Middleware{Logged, GlobalVisibility, ErrorGuard}, ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			un := r.PathValue("userName")
			if !model.ValidUserName(un) {
				rc.ReportNotFound(un, "User", "Depot", w, r)
				return
			}
			user, err := ctx.DatabaseInterface.GetUserByName(un)
			if err == db.ErrEntityNotFound {
				rc.ReportNotFound(un, "User", "depot", w, r)
				return
			}
			if err != nil {
				rc.ReportInternalError(err.Error(), w, r)
				return
			}
			viewingUser := ""
			if rc.LoginInfo != nil { viewingUser = rc.LoginInfo.UserName }
			nsList, err := rc.DatabaseInterface.GetAllBelongingNamespace(viewingUser, un)
			if err != nil {
				rc.ReportInternalError(err.Error(), w, r)
				return
			}

			pStr := r.URL.Query().Get("p")
			p, err := strconv.ParseInt(pStr, 10, 64)
			if err != nil || p <= 0 { p = 1 }
			sStr := r.URL.Query().Get("s")
			s, err := strconv.ParseInt(sStr, 10, 64)
			if err != nil || s <= 0 { s = 30 }
			q := strings.TrimSpace(r.URL.Query().Get("q"))
			count, err := rc.DatabaseInterface.CountAllBelongingRepository(viewingUser, un, q)
			if err != nil {
				rc.ReportInternalError(fmt.Sprintf("Failed to count all belonging repository: %s.\n", err.Error()), w, r)
				return
			}
			totalPage := count / s
			if totalPage <= 0 { totalPage = 1 }
			repoList, err := rc.DatabaseInterface.GetAllBelongingRepository(viewingUser, un, q, int(p-1), int(s))
			if err != nil {
				rc.ReportInternalError(err.Error(), w, r)
				return
			}
			LogTemplateError(ctx.LoadTemplate("user").Execute(w, templates.UserTemplateModel{
				User: user,
				RepositoryList: repoList,
				Config: rc.Config,
				LoginInfo: rc.LoginInfo,
				BelongingNamespaceList: nsList,
				PageInfo: &templates.PageInfoModel{
					PageNum: int(p),
					PageSize: int(s),
					TotalPage: int(totalPage),
				},
				Query: q,
			}))
		},
	))
}

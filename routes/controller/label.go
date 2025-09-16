package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindLabelController(ctx *RouterContext) {
	http.HandleFunc("GET /label/{label}", UseMiddleware(
		[]Middleware{
			Logged, UseLoginInfo, GlobalVisibility, ErrorGuard,
		}, ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			qStr := strings.TrimSpace(r.URL.Query().Get("q"))
			if len(qStr) > 0 {
				FoundAt(w, fmt.Sprintf("/label/%s", qStr));
				return
			}
			label := r.PathValue("label")
			pStr := r.URL.Query().Get("p")
			sStr := r.URL.Query().Get("s")
			p, err := strconv.ParseInt(pStr, 10, 64)
			if err != nil { p = 1 }
			s, err := strconv.ParseInt(sStr, 10, 64)
			if err != nil { s = 30}
			cnt, err := rc.DatabaseInterface.CountRepositoryWithLabel(rc.LoginInfo.UserName, label)
			if err != nil {
				rc.ReportInternalError(fmt.Sprintf("Failed while retrieving the number of repository with label: %s.", err), w, r)
				return
			}
			totalPage := cnt / s
			if totalPage <= 0 { totalPage = 1 }
			repoList, err := rc.DatabaseInterface.GetRepositoryWithLabelPaginated(rc.LoginInfo.UserName, label, p-1, s)
			LogTemplateError(rc.LoadTemplate("label").Execute(w, &templates.LabelModel{
				Config: rc.Config,
				RepositoryList: repoList,
				LoginInfo: rc.LoginInfo,
				PageInfo: &templates.PageInfoModel{
					PageNum: p,
					PageSize: s,
					TotalPage: totalPage,
				},
				Label: label,
			}))
		}))
}


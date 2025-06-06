package all

import (
	"net/http"
	"slices"
	"strings"

	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

// list of namespace if enabled,
// list of repo if namespace is not used,
func BindAllController(ctx *routes.RouterContext) {
	if ctx.Config.UseNamespace {
		http.HandleFunc("GET /all/namespace", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
			loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			q := strings.TrimSpace(r.URL.Query().Get("q"))
			var nsl map[string]*model.Namespace
			var nslCount int
			var pageInfo *templates.PageInfoModel
			if ctx.Config.PlainMode {
				if len(q) > 0 {
					nsl, err = ctx.Config.SearchAllNamespacePlain(q)
				} else {
					nsl, err = ctx.Config.GetAllNamespacePlain()
				}
				if err != nil {
					ctx.ReportInternalError(err.Error(), w, r)
					return
				}
				nslCount = len(nsl)
				pageInfo, err = routes.GeneratePageInfo(r, nslCount)
				if err != nil {
					ctx.ReportInternalError(err.Error(), w, r)
					return
				}
			} else {
				var nslCountv int64
				if len(q) > 0 {
					nslCountv, err = ctx.DatabaseInterface.CountAllVisibleNamespaceSearchResult(loginInfo.UserName, q)
					nslCount = int(nslCountv)
				} else {
					nslCountv, err = ctx.DatabaseInterface.CountAllVisibleNamespace(loginInfo.UserName)
					nslCount = int(nslCountv)
				}
				if err != nil {
					ctx.ReportInternalError(err.Error(), w, r)
					return
				}
				pageInfo, err = routes.GeneratePageInfo(r, nslCount)
				if err != nil {
					ctx.ReportInternalError(err.Error(), w, r)
					return
				}
				if len(q) > 0 {
					nsl, err = ctx.DatabaseInterface.SearchAllVisibleNamespacePaginated(loginInfo.UserName, q, pageInfo.PageNum-1, pageInfo.PageSize)
				} else {
					nsl, err = ctx.DatabaseInterface.GetAllVisibleNamespacePaginated(loginInfo.UserName, pageInfo.PageNum-1, pageInfo.PageSize)
				}
			}
			routes.LogTemplateError(ctx.LoadTemplate("all/namespace-list").Execute(w, templates.AllNamespaceListModel{
				DepotName: ctx.Config.DepotName,
				NamespaceList: nsl,
				Config: ctx.Config,
				LoginInfo: loginInfo,
				PageInfo: pageInfo,
				Query: q,
			}))
		}))
	}
	
	http.HandleFunc("GET /all/repo", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil {
			ctx.ReportInternalError(err.Error(), w, r)
			return
		}
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		var repol []*model.Repository
		var repolCount int
		var pageInfo *templates.PageInfoModel
		if ctx.Config.PlainMode {
			if len(q) > 0 {
				repol, err = ctx.Config.SearchAllRepositoryPlain(q)
			} else {
				repol, err = ctx.Config.GetAllRepositoryPlain()
			}
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			repolCount = len(repol)
			pageInfo, err = routes.GeneratePageInfo(r, repolCount)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			slices.SortFunc(repol, func(a, b *model.Repository) int {
				if a.Namespace < b.Namespace { return -1 }
				if a.Namespace > b.Namespace { return 1 }
				if a.Name < b.Name { return -1 }
				if a.Name > b.Name { return 1 }
				return 0
			})
		} else {
			var repolCountv int64
			if loginInfo.IsAdmin {
				if len(q) > 0 {
					repolCountv, err = ctx.DatabaseInterface.CountAllRepositoriesSearchResult(q)
				} else {
					repolCountv, err = ctx.DatabaseInterface.CountAllRepositories()
				}
			} else {
				if len(q) > 0 {
					repolCountv, err = ctx.DatabaseInterface.CountAllVisibleRepositoriesSearchResult(loginInfo.UserName, q)
				} else {
					repolCountv, err = ctx.DatabaseInterface.CountAllVisibleRepositories(loginInfo.UserName)
				}
			}
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			repolCount = int(repolCountv)
			pageInfo, err = routes.GeneratePageInfo(r, repolCount)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			if loginInfo.IsAdmin {
				if len(q) > 0 {
					repol, err = ctx.DatabaseInterface.SearchForRepository(q, pageInfo.PageNum - 1, pageInfo.PageSize)
				} else {
					repol, err = ctx.DatabaseInterface.GetAllRepositories(pageInfo.PageNum - 1, pageInfo.PageSize)
				}
			} else {
				if len(q) > 0 {
					repol, err = ctx.DatabaseInterface.SearchAllVisibleRepositoryPaginated(loginInfo.UserName, q, pageInfo.PageNum - 1, pageInfo.PageSize)
				} else {
					repol, err = ctx.DatabaseInterface.GetAllVisibleRepositoryPaginated(loginInfo.UserName, pageInfo.PageNum-1, pageInfo.PageSize)
				}
			}
		}
		routes.LogTemplateError(ctx.LoadTemplate("all/repository-list").Execute(w, templates.AllRepositoryListModel{
			RepositoryList: repol,
			DepotName: ctx.Config.DepotName,
			Config: ctx.Config,
			LoginInfo: loginInfo,
			PageInfo: pageInfo,
			Query: q,
		}))
	}))
}


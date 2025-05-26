package admin

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/bctnry/aegis/pkg/aegis/receipt"
	"github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

// /admin/receipt/{{.Id}}/edit
// /admin/receipt/{{.Id}}/confirm
// /admin/receipt/{{.Id}}/delete
func bindAdminReceiptListController(ctx *routes.RouterContext) {
	http.HandleFunc("GET /admin/receipt-list", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
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
		var receiptList []*receipt.Receipt
		if len(q) > 0 {
			receiptList, err = ctx.ReceiptSystem.SearchReceipt(q, int(pageNum-1), int(pageSize))
		} else {
			receiptList, err = ctx.ReceiptSystem.GetAllReceipt(int(pageNum-1), int(pageSize))
		}
		if err != nil {
			routes.LogTemplateError(ctx.LoadTemplate("admin/receipt-list").Execute(w, &templates.AdminReceiptListTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: fmt.Sprintf("Failed to load user list: %s", err.Error()),
				ReceiptList: nil,
			}))
			return
		}
		routes.LogTemplateError(ctx.LoadTemplate("admin/receipt-list").Execute(w, &templates.AdminReceiptListTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			ErrorMsg: "",
			ReceiptList: receiptList,
			Query: q,
			PageInfo: &templates.PageInfoModel{
				PageNum: int(pageNum),
				PageSize: int(pageSize),
				TotalPage: int(totalPage),
			},
		}))
	}))

	http.HandleFunc("GET /admin/receipt/{rid}/confirm", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		rid := r.PathValue("rid")
		routes.LogTemplateError(ctx.LoadTemplate("admin/_redirect-with-message").Execute(w, &templates.AdminRedirectWithMessageModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			ErrorMsg: "",
			Timeout: 0,
			RedirectUrl: fmt.Sprintf("/receipt?id=%s", rid),
			MessageTitle: "Ready to confirm receipt",
			MessageText: "",
		}))
		return
	}))

	http.HandleFunc("GET /admin/receipt/{rid}/delete", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		rid := r.PathValue("rid")
		err = ctx.ReceiptSystem.CancelReceipt(rid)
		if err != nil {
			routes.LogTemplateError(ctx.LoadTemplate("admin/_redirect-with-message").Execute(w, &templates.AdminRedirectWithMessageModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: "",
				Timeout: 3,
				RedirectUrl: "/admin/receipt-list",
				MessageTitle: "Failed to cancel receipt",
				MessageText: fmt.Sprintf("Failed to cancel receipt: %s", err.Error()),
			}))
		}
		routes.LogTemplateError(ctx.LoadTemplate("admin/_redirect-with-message").Execute(w, &templates.AdminRedirectWithMessageModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			ErrorMsg: "",
			Timeout: 5,
			RedirectUrl: "/admin/receipt-list",
			MessageTitle: "Receipt cancelled.",
			MessageText: "",
		}))
	}))
}

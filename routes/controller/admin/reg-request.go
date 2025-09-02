package admin

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/pkg/aegis/receipt"
	"github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

// /admin/reg-request?p={pagenum}&s={pagesize}
func bindAdminRegistrationRequestController(ctx *routes.RouterContext) {
	http.HandleFunc("GET /admin/reg-request", routes.UseMiddleware(
		[]routes.Middleware{
			routes.Logged, routes.LoginRequired, routes.AdminRequired, routes.ErrorGuard,
		}, ctx,
		func(rc *routes.RouterContext, w http.ResponseWriter, r *http.Request) {
			p := r.URL.Query().Get("p")
			if len(p) <= 0 { p = "1" }
			s := r.URL.Query().Get("s")
			if len(s) <= 0 { s = "50" }
			q := strings.TrimSpace(r.URL.Query().Get("q"))
			pageNum, err := strconv.ParseInt(p, 10, 32)
			if err != nil { rc.ReportNormalError("Invalid request", w, r); return }
			pageSize, err := strconv.ParseInt(s, 10, 32)
			if err != nil { rc.ReportNormalError("Invalid request", w, r); return }
			i, err := rc.DatabaseInterface.CountRegistrationRequest(q)
			totalPage := i / pageSize
			if totalPage <= 0 { totalPage = 1 }
			if pageNum > totalPage { pageNum = totalPage }
			if pageNum <= 1 { pageNum = 1 }
			var regreqList []*model.RegistrationRequest
			if len(q) > 0 {
				regreqList, err = ctx.DatabaseInterface.SearchRegistrationRequestPaginated(q, int(pageNum-1), int(pageSize))
			} else {
				regreqList, err = ctx.DatabaseInterface.GetRegistrationRequestPaginated(int(pageNum-1), int(pageSize))
			}
			routes.LogTemplateError(rc.LoadTemplate("admin/registration-request").Execute(w, &templates.AdminRegistrationRequestTemplateModel{
				Config: rc.Config,
				LoginInfo: rc.LoginInfo,
				ErrorMsg: "",
				RequestList: regreqList,
				PageInfo: &templates.PageInfoModel{
					PageNum: int(pageNum),
					PageSize: int(pageSize),
					TotalPage: int(totalPage),
				},
				Query: q,
			}))
		},
	))

	http.HandleFunc("GET /admin/reg-request/{absid}/approve", routes.UseMiddleware(
		[]routes.Middleware{
			routes.Logged, routes.LoginRequired, routes.AdminRequired, routes.ErrorGuard,
		}, ctx,
		func(rc *routes.RouterContext, w http.ResponseWriter, r *http.Request) {
			absidStr := r.PathValue("absid")
			absid, err := strconv.ParseInt(absidStr, 10, 64)
			if err != nil {
				rc.ReportNormalError("Invalid request", w, r)
				return
			}
			regreq, err := rc.DatabaseInterface.GetRegistrationRequestByAbsId(absid)
			if err != nil {
				rc.ReportInternalError(fmt.Sprintf("Failed to get registration request: %s.", err), w, r)
				return
			}
			err = rc.DatabaseInterface.ApproveRegistrationRequest(absid)
			if err != nil {
				rc.ReportInternalError(fmt.Sprintf("Failed to approve registration request: %s.", err), w, r)
				return
			}
			if ctx.Config.EmailConfirmationRequired {
				email := regreq.Email
				command := make([]string, 4)
				command[0] = receipt.CONFIRM_REGISTRATION
				command[1] = regreq.Username
				command[2] = email
				command[3] = regreq.PasswordHash
				rid, err := ctx.ReceiptSystem.IssueReceipt(24*60, command)
				if err != nil {
					rc.ReportInternalError(err.Error(), w, r)
					return
				}
				title := fmt.Sprintf("Confirmation of registering on %s", rc.Config.DepotName)
				body := fmt.Sprintf(`
This email is used to register on %s, a code repository hosting platform.

If this isn't you, you don't need to do anything about it, as the registration
request expires after 24 hours; but if this is you, please copy & open the
following link to confirm your registration:

    %s/receipt?id=%s

We wish you all the best in your future endeavours.

%s
`, ctx.Config.DepotName, rc.Config.ProperHTTPHostName(), rid, rc.Config.DepotName)
				err = rc.Mailer.SendPlainTextMail(email, title, body)
				// NOTE: the call to the registration wouldn't occur in this case
				// until the receipt link is visited.
			} else {
				if ctx.Config.UseNamespace {
					_, err = ctx.DatabaseInterface.RegisterNamespace(regreq.Username, regreq.Username)
					if err != nil {
						ctx.ReportInternalError(
							fmt.Sprintf("Failed at registering namespace %s. Please contact site admin for this issue.", err.Error()),
							w, r,
						)
						return
					}
				}
			}
			
			rc.ReportRedirect("/admin/reg-request", 0, "Approved", "User registration approved.", w, r)
		},
	))
	
	http.HandleFunc("GET /admin/reg-request/{absid}/disapprove", routes.UseMiddleware(
		[]routes.Middleware{
			routes.Logged, routes.LoginRequired, routes.AdminRequired, routes.ErrorGuard,
		}, ctx,
		func(rc *routes.RouterContext, w http.ResponseWriter, r *http.Request) {
			
			absidStr := r.PathValue("absid")
			absid, err := strconv.ParseInt(absidStr, 10, 64)
			if err != nil {
				rc.ReportNormalError("Invalid request", w, r)
				return
			}
			err = rc.DatabaseInterface.DisapproveRegistrationRequest(absid)
			if err != nil {
				rc.ReportInternalError(fmt.Sprintf("Failed while disapproving registration request: %s.", err), w, r)
				return
			}
			rc.ReportRedirect("/admin/reg-request", 0, "Disapproved", "User registration disapproved.", w, r)
		},
	))
}


package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/GitusCodeForge/Gitus/pkg/gitus"
	"github.com/GitusCodeForge/Gitus/pkg/gitus/model"
	"github.com/GitusCodeForge/Gitus/pkg/gitus/receipt"
	. "github.com/GitusCodeForge/Gitus/routes"
	"github.com/GitusCodeForge/Gitus/templates"
)

func bindConfirmRegistrationController(ctx *RouterContext) {
	http.HandleFunc("GET /confirm-registration", UseMiddleware(
		[]Middleware{Logged, UseLoginInfo, ErrorGuard}, ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			if ctx.Config.GlobalVisibility == gitus.GLOBAL_VISIBILITY_SHUTDOWN {
				FoundAt(w, "/shutdown-notice")
			}
			if ctx.Config.GlobalVisibility == gitus.GLOBAL_VISIBILITY_MAINTENANCE {
				FoundAt(w, "/maintenance-notice")
			}
			rid := r.URL.Query().Get("id")
			re, err := ctx.ReceiptSystem.RetrieveReceipt(rid)
			if err != nil { FoundAt(w, "/"); return }
			if (time.Now().Unix() - re.IssueTime) >= re.TimeoutMinute*60 {
				ctx.ReceiptSystem.CancelReceipt(rid)
				FoundAt(w, "/");
				return
			}
			if len(re.Command) <= 0 || re.Command[0] != receipt.CONFIRM_REGISTRATION {
				// invalid receipt command...
				FoundAt(w, "/")
			}
			username := re.Command[1]
			email := re.Command[2]
			passwordHash := re.Command[3]
			status := model.NORMAL_USER
			if ctx.Config.ManualApproval { status = model.NORMAL_USER_APPROVAL_NEEDED }
			ctx.ReceiptSystem.CancelReceipt(rid)
			_, err = ctx.DatabaseInterface.RegisterUser(username, email, passwordHash, status)
			if err != nil {
				ctx.ReportInternalError(
					fmt.Sprintf("Failed at registering: %s. Please contact site admin for this issue.", err.Error()),
					w, r,
				)
				return
			}
			if ctx.Config.UseNamespace {
				_, err = ctx.DatabaseInterface.RegisterNamespace(username, username)
				if err != nil {
					ctx.ReportInternalError(
						fmt.Sprintf("Failed at registering namespace %s. Please contact site admin for this issue.", err.Error()),
						w, r,
					)
					return
				}
			}
			LogTemplateError(ctx.LoadTemplate("error").Execute(w, &templates.ErrorTemplateModel{
				Config: ctx.Config,
				ErrorCode: 200,
				ErrorMessage: "Registration complete. You can try to login now.",
				LoginInfo: nil,
			}))
		},
	))
}


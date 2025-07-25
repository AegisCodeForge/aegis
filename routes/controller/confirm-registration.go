package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/pkg/aegis/receipt"
	"github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
)

func bindConfirmRegistrationController(ctx *routes.RouterContext) {
	http.HandleFunc("GET /confirm-registration", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		if ctx.Config.GlobalVisibility == aegis.GLOBAL_VISIBILITY_SHUTDOWN {
			routes.FoundAt(w, "/shutdown-notice")
		}
		if ctx.Config.GlobalVisibility == aegis.GLOBAL_VISIBILITY_MAINTENANCE {
			routes.FoundAt(w, "/maintenance-notice")
		}
		rid := r.URL.Query().Get("id")
		re, err := ctx.ReceiptSystem.RetrieveReceipt(rid)
		if err != nil { routes.FoundAt(w, "/"); return }
		if (time.Now().Unix() - re.IssueTime) >= re.TimeoutMinute*60 {
			ctx.ReceiptSystem.CancelReceipt(rid)
			routes.FoundAt(w, "/");
			return
		}
		if len(re.Command) <= 0 || re.Command[0] != receipt.CONFIRM_REGISTRATION {
			// invalid receipt command...
			routes.FoundAt(w, "/")
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
		routes.LogTemplateError(ctx.LoadTemplate("error").Execute(w, &templates.ErrorTemplateModel{
			Config: ctx.Config,
			ErrorCode: 200,
			ErrorMessage: "Registration complete. You can try to login now.",
			LoginInfo: nil,
		}))
	}))
}


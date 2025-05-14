package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bctnry/gitus/pkg/gitus/model"
	"github.com/bctnry/gitus/pkg/gitus/receipt"
	"github.com/bctnry/gitus/routes"
	"github.com/bctnry/gitus/templates"
)

func bindConfirmRegistrationController(ctx *routes.RouterContext) {
	http.HandleFunc("GET /confirm-registration", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
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
		_, err = ctx.DatabaseInterface.RegisterUser(username, email, passwordHash, status)
		if err != nil {
			ctx.ReportInternalError(
				fmt.Sprintf("Failed at registering: %s. Please contact site admin for this issue.", err.Error()),
				w, r,
			)
			return
		}
		routes.LogTemplateError(ctx.LoadTemplate("error").Execute(w, &templates.ErrorTemplateModel{
			Config: ctx.Config,
			ErrorCode: 200,
			ErrorMessage: "Registration complete. You can try to login now.",
			LoginInfo: nil,
		}))
	}))
}


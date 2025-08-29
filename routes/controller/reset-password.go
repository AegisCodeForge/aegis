// /reset-password/request
// /reset-password/update-password

package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/pkg/aegis/receipt"
	"github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
	"golang.org/x/crypto/bcrypt"
)

func bindResetPasswordController(ctx *routes.RouterContext) {
	http.HandleFunc("GET /reset-password/request", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		switch ctx.Config.GlobalVisibility {
		case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
			routes.FoundAt(w, "/maintenance-notice")
			return
		case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
			routes.FoundAt(w, "/shutdown-notice")
			return
		}
		loginInfo, _ := routes.GenerateLoginInfoModel(ctx, r)
		routes.LogTemplateError(ctx.LoadTemplate("reset-password-request").Execute(w, struct{
			Config *aegis.AegisConfig
			ErrorMsg string
			LoginInfo *templates.LoginInfoModel
		}{
			Config: ctx.Config,
			ErrorMsg: "",
			LoginInfo: loginInfo,
		}))
	}))

	http.HandleFunc("POST /reset-password/request", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		switch ctx.Config.GlobalVisibility {
		case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
			routes.FoundAt(w, "/maintenance-notice")
			return
		case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
			routes.FoundAt(w, "/shutdown-notice")
			return
		}
		err := r.ParseForm()
		if err != nil {
			ctx.ReportRedirect("/reset-password/request", 0, "Invalid Request", "Failed to parse request. Please try again.", w, r)
			return
		}
		targetUserName := strings.TrimSpace(r.Form.Get("username"))
		if !model.ValidUserName(targetUserName) {
			ctx.ReportNotFound(targetUserName, "User", "Depot", w, r)
			return
		}
		targetEmail := strings.TrimSpace(r.Form.Get("email"))
		user, err := ctx.DatabaseInterface.GetUserByName(targetUserName)
		if err != nil {
			ctx.ReportRedirect("/reset-password/request", 0,
				"Reset Failed",
				fmt.Sprintf("Failed to initiate password reset request: %s.", err.Error()),
				w, r,
			)
			return
		}
		if user.Email == targetEmail {
			iid, err := ctx.ReceiptSystem.IssueReceipt(24*60, []string{
				"reset-password",
				strings.TrimSpace(targetUserName),
			})
			if err != nil {
				ctx.ReportRedirect("/reset-password/request", 0,
					"Internal Error",
					fmt.Sprintf("Failed to initiate password reset request: %s. Please contact the site owner for this...", err.Error()),
					w, r,
				)
				return
			}
			ctx.Mailer.SendPlainTextMail(
				targetEmail,
				fmt.Sprintf("Reset password instructions from %s", ctx.Config.DepotName),
				fmt.Sprintf(`Dear user,

%s has received your request for a password reset. Please visit the following link to proceed with the process:

    %s/receipt?id=%s

This link would become invalid after 24 hours.

If this isn't you, you can simply ignore this message.`,
					ctx.Config.DepotName,
					ctx.Config.ProperHTTPHostName(),
					iid,
				),
			)
		}
		ctx.ReportRedirect("/", 0, "Request Recieved", "Your request of password reset has been received. If the info matches, an email would be sent to your email address; please proceed from there.", w, r)
	}))

	http.HandleFunc("GET /reset-password/update-password", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		switch ctx.Config.GlobalVisibility {
		case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
			routes.FoundAt(w, "/maintenance-notice")
			return
		case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
			routes.FoundAt(w, "/shutdown-notice")
			return
		}
		rid := strings.TrimSpace(r.URL.Query().Get("id"))
		if len(rid) <= 0 { routes.FoundAt(w, "/reset-password/request"); return }
		re, err := ctx.ReceiptSystem.RetrieveReceipt(rid)
		if err != nil {
			ctx.ReportInternalError(
				fmt.Sprintf("Failed while retrieving receipt: %s", err.Error()),
				w, r,
			)
			return
		}
		if re.Expired() {
			ctx.ReceiptSystem.CancelReceipt(rid)
			ctx.ReportRedirect("/", 5, "Receipt Expired", "The receipt you've received has passed its validity time limit. Please go through the process again.", w, r)
			return
		}
		if re.Command[0] != receipt.RESET_PASSWORD && len(re.Command) != 1 {
			ctx.ReceiptSystem.CancelReceipt(rid)
			ctx.ReportRedirect("/", 5, "Invalid Receipt", "The receipt you've provided is invalid. Please try again.", w, r)
			return
		}
		routes.LogTemplateError(ctx.LoadTemplate("reset-password-update").Execute(w, struct {
			Config *aegis.AegisConfig
			ReceiptId string
			LoginInfo *templates.LoginInfoModel
		}{
			Config: ctx.Config,
			ReceiptId: rid,
			LoginInfo: nil,
		}))
	}))

	http.HandleFunc("POST /reset-password/update-password", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		switch ctx.Config.GlobalVisibility {
		case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
			routes.FoundAt(w, "/maintenance-notice")
			return
		case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
			routes.FoundAt(w, "/shutdown-notice")
			return
		}
		err := r.ParseForm()
		if err != nil {
			ctx.ReportRedirect("/reset-password/request", 3, "Invalid Request", "Invalid request.", w, r)
			return
		}
		rid := strings.TrimSpace(r.Form.Get("rid"))
		re, err := ctx.ReceiptSystem.RetrieveReceipt(rid)
		if err != nil {
			ctx.ReportNormalError(
				fmt.Sprintf("Internal error: %s", err.Error()),
				w, r,
			)
			return
		}
		if re.Expired() {
			ctx.ReceiptSystem.CancelReceipt(rid)
			ctx.ReportRedirect("/", 5, "Receipt Expired", "The receipt you've received has passed its validity time limit. Please go through the process again.", w, r)
			return
		}
		if re.Command[0] != receipt.RESET_PASSWORD && len(re.Command) != 1 {
			ctx.ReceiptSystem.CancelReceipt(rid)
			ctx.ReportRedirect("/", 5, "Invalid Receipt", "The receipt you've provided is invalid. Please try again.", w, r)
			return
		}
		targetUserName := re.Command[1]
		if !model.ValidUserName(targetUserName) {
			ctx.ReportNotFound(targetUserName, "User", "Depot", w, r)
			return
		}
		newPassword := strings.TrimSpace(r.Form.Get("password"))
		confirm := strings.TrimSpace(r.Form.Get("confirm"))
		if newPassword != confirm {
			ctx.ReportRedirect(
				fmt.Sprintf("/reset-password/update-password?id=%s", rid),
				3,
				"Password Not Match",
				"The password you entered does not match. Please enter the *same* password in both of the password fields.",
				w, r,
			)
			return
		}
		newPassHashBytes, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			ctx.ReportNormalError(
				fmt.Sprintf("Internal error: %s", err.Error()),
				w, r,
			)
			return
		}
		err = ctx.DatabaseInterface.UpdateUserPassword(targetUserName, string(newPassHashBytes))
		if err != nil {
			ctx.ReportNormalError(
				fmt.Sprintf("Internal error: %s", err.Error()),
				w, r,
			)
			return
		}
		ctx.ReceiptSystem.CancelReceipt(rid)
		ctx.ReportRedirect("/login", 3, "Password Updated", "Your password has been updated.", w, r)
	}))
	
}


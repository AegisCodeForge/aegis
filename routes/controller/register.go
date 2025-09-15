package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/pkg/aegis/receipt"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
	"golang.org/x/crypto/bcrypt"
)


func bindRegisterController(ctx *RouterContext) {
	http.HandleFunc("GET /reg", UseMiddleware(
		[]Middleware{Logged, UseLoginInfo, ErrorGuard}, ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			switch rc.Config.GlobalVisibility {
			case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
				FoundAt(w, "/maintenance-notice")
				return
			case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
				FoundAt(w, "/shutdown-notice")
				return
			}
			if !rc.Config.AllowRegistration { FoundAt(w, "/"); return }
			LogTemplateError(rc.LoadTemplate("registration").Execute(w, templates.RegistrationTemplateModel{
				Config: rc.Config,
				ErrorMsg: "",
				LoginInfo: rc.LoginInfo,
			}))
		},
	))

	http.HandleFunc("POST /reg", UseMiddleware(
		[]Middleware{Logged, RateLimit, ValidPOSTRequestRequired, ErrorGuard}, ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			switch rc.Config.GlobalVisibility {
			case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
				FoundAt(w, "/maintenance-notice")
				return
			case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
				FoundAt(w, "/shutdown-notice")
				return
			}
			if !rc.Config.AllowRegistration {
				rc.ReportNormalError("Registration not allowed on this instance.", w, r)
				return
			}
			userName := r.Form.Get("username")
			if !model.ValidUserName(userName) {
				rc.ReportRedirect("/reg", 5, "Invalid User Name", "User name must consists of only upper & lowercase letters (a-z, A-Z), 0-9, underscore and hyphen.", w, r)
				return
			}
			email := r.Form.Get("email")

			// username & ns name check.
			_, err := rc.DatabaseInterface.GetUserByName(userName)
			if err == nil {
				LogTemplateError(rc.LoadTemplate("registration").Execute(w, &templates.RegistrationTemplateModel{
					Config: rc.Config,
					LoginInfo: nil,
					ErrorMsg: "Username/Namespace name already exists. Please try another name.",
				}))
				return
			}
			_, err = rc.DatabaseInterface.GetNamespaceByName(userName)
			if err == nil {
				LogTemplateError(rc.LoadTemplate("registration").Execute(w, &templates.RegistrationTemplateModel{
					Config: rc.Config,
					LoginInfo: nil,
					ErrorMsg: "Username/Namespace name already exists. Please try another name.",
				}))
				return
			}
			password := r.Form.Get("password")
			passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				rc.ReportInternalError(fmt.Sprintf("Failed to hash the provided password: %s. Please try again.", err.Error()), w, r)
				return
			}
			
			succeedMsg := "Registered. You can now log in."
			if rc.Config.ManualApproval {
				err = rc.DatabaseInterface.InsertRegistrationRequest(userName, email, string(passwordHash), strings.TrimSpace(r.Form.Get("reason")))
				if err != nil {
					rc.ReportInternalError(fmt.Sprintf("Failed to submit registration request: %s. Please contact the site owner.", err.Error()), w, r)
					return
				} else {
					msg := "Your registration request has been submitted. "
					if rc.Config.EmailConfirmationRequired {
						msg += " You will receive the confirmation email after the administrators approved your request."
					} else {
						msg += " Your account would be usable after the administrators approved your request."
					}
					rc.ReportRedirect("/", 0, "Request Submitted", msg, w, r)
					return
				}
			}
			
			if rc.Config.EmailConfirmationRequired {
				command := make([]string, 4)
				command[0] = receipt.CONFIRM_REGISTRATION
				command[1] = userName
				command[2] = email
				command[3] = string(passwordHash)
				rid, err := rc.ReceiptSystem.IssueReceipt(24*60, command)
				if err != nil {
					rc.ReportInternalError(fmt.Sprintf("Failed to issue receipt for registration: %s", err.Error()), w, r)
					return
				}
				go func() {
					email := r.Form.Get("email")
					title := fmt.Sprintf("Confirmation of registering on %s", rc.Config.DepotName)
					body := fmt.Sprintf(`
This email is used to register on %s, a code repository hosting platform.

If this isn't you, you don't need to do anything about it, as the registration
request expires after 24 hours; but if this is you, please copy & open the
following link to confirm your registration:

    %s/receipt?id=%s

We wish you all the best in your future endeavours.

%s
`, rc.Config.DepotName, rc.Config.ProperHTTPHostName(), rid, rc.Config.DepotName)
					err = rc.Mailer.SendPlainTextMail(email, title, body)
				}()
				succeedMsg = "A confirmation email has been sent to the email address you have specified. Please proceed from there."
			} else {
				status := model.NORMAL_USER
				_, err = rc.DatabaseInterface.RegisterUser(userName, email, string(passwordHash), status)
				if err != nil {
					LogTemplateError(rc.LoadTemplate("registration").Execute(w, &templates.RegistrationTemplateModel{
						Config: rc.Config,
						LoginInfo: nil,
						ErrorMsg: fmt.Sprintf("Error while registering: %s. Please try again.", err.Error()),
					}))
					return
				}
				if rc.Config.UseNamespace {
					_, err = rc.DatabaseInterface.RegisterNamespace(userName, userName)
					if err != nil {
						rc.ReportInternalError(
							fmt.Sprintf("Failed at registering namespace %s. Please contact site admin for this issue.", err.Error()),
							w, r,
						)
						return
					}
				}
			}
			loginInfo, _ := GenerateLoginInfoModel(ctx, r)
			LogTemplateError(rc.LoadTemplate("error").Execute(w, &templates.ErrorTemplateModel{
				Config: rc.Config,
				ErrorCode: 200,
				ErrorMessage: succeedMsg,
				LoginInfo: loginInfo,
			}))
		},
	))
}

	

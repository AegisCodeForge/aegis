package controller

import (
	"fmt"
	"net/http"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/pkg/aegis/receipt"
	"github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
	"golang.org/x/crypto/bcrypt"
)


func bindRegisterController(ctx *routes.RouterContext) {
	http.HandleFunc("GET /reg", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		switch ctx.Config.GlobalVisibility {
		case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
			routes.FoundAt(w, "/maintenance-notice")
			return
		case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
			routes.FoundAt(w, "/shutdown-notice")
			return
		}
		if !ctx.Config.AllowRegistration { routes.FoundAt(w, "/"); return }
		loginInfo, _ := routes.GenerateLoginInfoModel(ctx, r)
		routes.LogTemplateError(ctx.LoadTemplate("registration").Execute(w, templates.RegistrationTemplateModel{
			Config: ctx.Config,
			ErrorMsg: "",
			LoginInfo: loginInfo,
		}))
	}))
	// TODO: rate limit.
	http.HandleFunc("POST /reg", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		switch ctx.Config.GlobalVisibility {
		case aegis.GLOBAL_VISIBILITY_MAINTENANCE:
			routes.FoundAt(w, "/maintenance-notice")
			return
		case aegis.GLOBAL_VISIBILITY_SHUTDOWN:
			routes.FoundAt(w, "/shutdown-notice")
			return
		}
		if !ctx.Config.AllowRegistration {
			ctx.ReportNormalError("Registration not allowed on this instance.", w, r)
			return
		}
		err := r.ParseForm()
		if err != nil {
			ctx.ReportNormalError("Invalid request.", w, r)
			return
		}
		userName := r.Form.Get("username")
		if !model.ValidUserName(userName) {
			ctx.ReportRedirect("/reg", 5, "Invalid User Name", "User name must consists of only upper & lowercase letters (a-z, A-Z), 0-9, underscore and hyphen.", w, r)
			return
		}
		email := r.Form.Get("email")

		// username & ns name check.
		_, err = ctx.DatabaseInterface.GetUserByName(userName)
		if err == nil {
			routes.LogTemplateError(ctx.LoadTemplate("registration").Execute(w, &templates.RegistrationTemplateModel{
				Config: ctx.Config,
				LoginInfo: nil,
				ErrorMsg: "Username/Namespace name already exists. Please try another name.",
			}))
			return
		}
		_, err = ctx.DatabaseInterface.GetNamespaceByName(userName)
		if err == nil {
			routes.LogTemplateError(ctx.LoadTemplate("registration").Execute(w, &templates.RegistrationTemplateModel{
				Config: ctx.Config,
				LoginInfo: nil,
				ErrorMsg: "Username/Namespace name already exists. Please try another name.",
			}))
			return
		}
		password := r.Form.Get("password")
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			ctx.ReportInternalError(fmt.Sprintf("Failed to hash the provided password: %s. Please try again.", err.Error()), w, r)
			return
		}
		
		succeedMsg := "Registered. You can now log in."
		if ctx.Config.EmailConfirmationRequired {
			command := make([]string, 4)
			command[0] = receipt.CONFIRM_REGISTRATION
			command[1] = userName
			command[2] = email
			command[3] = string(passwordHash)
			rid, err := ctx.ReceiptSystem.IssueReceipt(24*60, command)
			if err != nil {
				ctx.ReportInternalError(err.Error(), w, r)
				return
			}
			email := r.Form.Get("email")
			title := fmt.Sprintf("Confirmation of registering on %s", ctx.Config.DepotName)
			body := fmt.Sprintf(`
This email is used to register on %s, a code repository hosting platform.

If this isn't you, you don't need to do anything about it, as the registration
request expires after 24 hours; but if this is you, please copy & open the
following link to confirm your registration:

    %s/receipt?id=%s

We wish you all the best in your future endeavours.

%s
`, ctx.Config.DepotName, ctx.Config.ProperHTTPHostName(), rid, ctx.Config.DepotName)
			err = ctx.Mailer.SendPlainTextMail(email, title, body)
			if err != nil {
				routes.LogTemplateError(ctx.LoadTemplate("registration").Execute(w, &templates.RegistrationTemplateModel{
					Config: ctx.Config,
					LoginInfo: nil,
					ErrorMsg: fmt.Sprintf("Error while registering: %s. Please try again.", err.Error()),
				}))
				return
			}
			succeedMsg = "Registered. You should be able to use your account after email confirmation."
		} else {
			status := model.NORMAL_USER
			if ctx.Config.ManualApproval {
				status = model.NORMAL_USER_APPROVAL_NEEDED
			}
			_, err = ctx.DatabaseInterface.RegisterUser(userName, email, string(passwordHash), status)
			if err != nil {
				routes.LogTemplateError(ctx.LoadTemplate("registration").Execute(w, &templates.RegistrationTemplateModel{
					Config: ctx.Config,
					LoginInfo: nil,
					ErrorMsg: fmt.Sprintf("Error while registering: %s. Please try again.", err.Error()),
				}))
				return
			}
			if ctx.Config.UseNamespace {
				_, err = ctx.DatabaseInterface.RegisterNamespace(userName, userName)
				if err != nil {
					ctx.ReportInternalError(
						fmt.Sprintf("Failed at registering namespace %s. Please contact site admin for this issue.", err.Error()),
						w, r,
					)
					return
				}
			}
		}
		if ctx.Config.ManualApproval {
			succeedMsg = "Registered. You should be able to use your account after admin approval."
		}
		loginInfo, _ := routes.GenerateLoginInfoModel(ctx, r)
		routes.LogTemplateError(ctx.LoadTemplate("error").Execute(w, &templates.ErrorTemplateModel{
			Config: ctx.Config,
			ErrorCode: 200,
			ErrorMessage: succeedMsg,
			LoginInfo: loginInfo,
		}))
	}))
}

	

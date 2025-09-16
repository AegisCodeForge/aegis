package admin

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/bctnry/aegis/pkg/aegis/model"
	. "github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
	"golang.org/x/crypto/bcrypt"
)

func bindAdminNewUserController(ctx *RouterContext) {
	http.HandleFunc("GET /admin/new-user", UseMiddleware(
		[]Middleware{Logged, LoginRequired, AdminRequired,
			GlobalVisibility, ErrorGuard,
		}, ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			LogTemplateError(rc.LoadTemplate("admin/new-user").Execute(w, &templates.AdminConfigTemplateModel{
				Config: rc.Config,
				LoginInfo: rc.LoginInfo,
				ErrorMsg: "",
			}))
		},
	))
	
	http.HandleFunc("POST /admin/new-user", UseMiddleware(
		[]Middleware{
			Logged, ValidPOSTRequestRequired,
			LoginRequired, AdminRequired,
			GlobalVisibility, ErrorGuard,
		}, ctx,
		func(rc *RouterContext, w http.ResponseWriter, r *http.Request) {
			userName := r.Form.Get("username")
			if !model.ValidUserName(userName) {
				rc.ReportRedirect("/admin/new-user", 5, "Invalid User Name", "Invalid user name; user name must only contains uppercase & lowercase letters (a-z, A-Z), 0-9, underscore and hyphen.\n", w, r)
				return
			}
			email := r.Form.Get("email")
			password := r.Form.Get("password")
			passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				LogTemplateError(rc.LoadTemplate("admin/new-user").Execute(w, templates.AdminConfigTemplateModel{
					Config: rc.Config,
					LoginInfo: rc.LoginInfo,
					ErrorMsg: fmt.Sprintf("Failed to hash the provided password: %s. Please try again.\n", err.Error()),
				}))
				return
			}
			i, err := strconv.ParseInt(r.Form.Get("status"), 10, 32)
			if err != nil {
				LogTemplateError(rc.LoadTemplate("admin/new-user").Execute(w, templates.AdminConfigTemplateModel{
					Config: rc.Config,
					LoginInfo: rc.LoginInfo,
					ErrorMsg: fmt.Sprintf("Invalid user status: %s.\n", err.Error()),
				}))
				return
			}
			_, err = rc.DatabaseInterface.RegisterUser(userName, email, string(passwordHash), model.AegisUserStatus(i))
			if err != nil {
				LogTemplateError(rc.LoadTemplate("admin/new-user").Execute(w, templates.AdminConfigTemplateModel{
					Config: rc.Config,
					LoginInfo: rc.LoginInfo,
					ErrorMsg: fmt.Sprintf("Failed to create new user: %s.\n", err.Error()),
				}))
				return
			}
			FoundAt(w, "/admin/user-list")
		},
	))
}


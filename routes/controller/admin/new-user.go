package admin

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/templates"
	"golang.org/x/crypto/bcrypt"
)

func bindAdminNewUserController(ctx *routes.RouterContext) {
	http.HandleFunc("GET /admin/new-user", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		routes.LogTemplateError(ctx.LoadTemplate("admin/new-user").Execute(w, &templates.AdminConfigTemplateModel{
			Config: ctx.Config,
			LoginInfo: loginInfo,
			ErrorMsg: "",
		}))
	}))
	http.HandleFunc("POST /admin/new-user", routes.WithLog(func(w http.ResponseWriter, r *http.Request) {
		loginInfo, err := routes.GenerateLoginInfoModel(ctx, r)
		if err != nil { routes.FoundAt(w, "/") }
		if !loginInfo.LoggedIn { routes.FoundAt(w, "/") }
		if !loginInfo.IsAdmin { routes.FoundAt(w, "/") }
		err = r.ParseForm()
		if err != nil {
			routes.LogTemplateError(ctx.LoadTemplate("admin/new-user").Execute(w, templates.AdminConfigTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: fmt.Sprintf("Failed to parse request: %s\n", err.Error()),
			}))
			return
		}
		userName := r.Form.Get("username")
		if !model.ValidUserName(userName) {
			ctx.ReportRedirect("/admin/new-user", 5, "Invalid User Name", "Invalid user name; user name must only contains uppercase & lowercase letters (a-z, A-Z), 0-9, underscore and hyphen.\n", w, r)
			return
		}
		email := r.Form.Get("email")
		password := r.Form.Get("password")
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			routes.LogTemplateError(ctx.LoadTemplate("admin/new-user").Execute(w, templates.AdminConfigTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: fmt.Sprintf("Failed to hash the provided password: %s. Please try again.\n", err.Error()),
			}))
			return
		}
		i, err := strconv.ParseInt(r.Form.Get("status"), 10, 32)
		if err != nil {
			routes.LogTemplateError(ctx.LoadTemplate("admin/new-user").Execute(w, templates.AdminConfigTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: fmt.Sprintf("Invalid user status: %s.\n", err.Error()),
			}))
			return
		}
		u, err := ctx.DatabaseInterface.RegisterUser(userName, email, string(passwordHash), model.AegisUserStatus(i))
		if err != nil {
			routes.LogTemplateError(ctx.LoadTemplate("admin/new-user").Execute(w, templates.AdminConfigTemplateModel{
				Config: ctx.Config,
				LoginInfo: loginInfo,
				ErrorMsg: fmt.Sprintf("Failed to create new user: %s.\n", err.Error()),
			}))
			return
		}
		fmt.Println(u, err)
		routes.FoundAt(w, "/admin/user-list")
	}))
}

package routes

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/templates"
)

func LogIfError(err error) {
	if err != nil {
		log.Print(err.Error())
	}
}

// go don't have ufcs so i'll have to suffer.
func WithLog(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(fmt.Sprintf(" %s %s", r.Method, r.URL.Path))
		f(w, r)
	}
}
func WithLogHandler(f http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(fmt.Sprintf(" %s %s", r.Method, r.URL.Path))
		f.ServeHTTP(w, r)
	}
}

func FoundAt(w http.ResponseWriter, p string) {
	w.Header().Add("Content-Length", "0")
	w.Header().Add("Location", p)
	w.WriteHeader(302)
}

func LoadTemplate(t *template.Template, name string) *template.Template {
	res := t.Lookup(name)
	if res == nil { log.Fatal(fmt.Sprintf("Failed to find template \"%s\"", name)) }
	return res
}

func LogTemplateError(e error) {
	if e != nil { log.Print(e) }
}

func GetUsernameFromCookie(r *http.Request) (string, error) {
	s, err := r.Cookie(COOKIE_KEY_USERNAME)
	if err != nil {
		return "", err
	} else {
		return s.Value, err
	}
}

func CheckUserSession(ctx *RouterContext, r *http.Request) (bool, error) {
	// the fact that go uses product type as disjoint union is
	// actually some ridiculously insane take. i have heard rumours
	// before starting this project, but to actually doing things this
	// way is a different story.
	un, err := GetUsernameFromCookie(r)
	if err == http.ErrNoCookie { return false, nil }
	if err != nil { return false, err }
	s, err := r.Cookie(COOKIE_KEY_SESSION)
	if err == http.ErrNoCookie { return false, nil }
	if err != nil { return false, err }
	res, err := ctx.SessionInterface.VerifySession(un, s.Value)
	if err != nil { return false, err }
	return res, nil
}

func GenerateLoginInfoModel(ctx *RouterContext, r *http.Request) (*templates.LoginInfoModel, error) {
	loggedIn := false
	un, err := GetUsernameFromCookie(r)
	if err != nil {
		if err != http.ErrNoCookie { return nil, err }
		return &templates.LoginInfoModel{
			LoggedIn: loggedIn,
			UserName: "",
		}, nil
	}
	s, err := r.Cookie(COOKIE_KEY_SESSION)
	if err != nil {
		if err != http.ErrNoCookie { return nil, err }
		return &templates.LoginInfoModel{
			LoggedIn: loggedIn,
			UserName: "",
		}, nil
	}
	res, err := ctx.SessionInterface.VerifySession(un, s.Value)
	if err != nil { return nil, err }
	u, err := ctx.DatabaseInterface.GetUserByName(un)
	if err != nil { return nil, err }
	return &templates.LoginInfoModel{
		LoggedIn: res,
		UserName: un,
		IsAdmin: u.Status == model.ADMIN || u.Status == model.SUPER_ADMIN,
	}, nil
}

func GenerateRepoHeader(ctx *RouterContext, repo *model.Repository, typeStr string, nodeName string) *templates.RepoHeaderTemplateModel {
	httpHostName := ctx.Config.ProperHTTPHostName()
	gitSshHostName := ctx.Config.GitSSHHostName()
	rfn := repo.FullName()
	repoHeaderInfo := &templates.RepoHeaderTemplateModel{
		NamespaceName: repo.Namespace,
		RepoName: repo.Name,
		RepoDescription: repo.Description,
		TypeStr: typeStr,
		NodeName: nodeName,
		RepoLabelList: nil,
		RepoURL: fmt.Sprintf("%s/repo/%s", httpHostName, rfn),
		RepoSSH: fmt.Sprintf("%s%s", gitSshHostName, rfn),
	}
	return repoHeaderInfo
}

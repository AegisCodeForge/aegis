package routes

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/bctnry/gitus/pkg/gitus"
	"github.com/bctnry/gitus/pkg/gitus/db"
	"github.com/bctnry/gitus/pkg/gitus/mail"
	"github.com/bctnry/gitus/pkg/gitus/model"
	"github.com/bctnry/gitus/pkg/gitus/receipt"
	"github.com/bctnry/gitus/pkg/gitus/session"
	"github.com/bctnry/gitus/pkg/gitus/ssh"
	"github.com/bctnry/gitus/templates"
)

type RouterContext struct {
	Config *gitus.GitusConfig
	MasterTemplate *template.Template
	GitRepositoryList map[string]*model.Repository
	GitNamespaceList map[string]*model.Namespace
	DatabaseInterface db.GitusDatabaseInterface
	SessionInterface session.GitusSessionStore
	SSHKeyManagingContext *ssh.SSHKeyManagingContext
	ReceiptSystem receipt.GitusReceiptSystemInterface
	Mailer mail.GitusMailerInterface
}

func (ctx RouterContext) LoadTemplate(name string) *template.Template {
	return ctx.MasterTemplate.Lookup(name)
}

func (ctx RouterContext) ReportNotFound(objName string, objType string, namespace string, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	LogTemplateError(ctx.LoadTemplate("error").Execute(w,
		templates.ErrorTemplateModel{
			ErrorCode: 404,
			ErrorMessage: fmt.Sprintf(
				"%s %s not found in %s",
				objType, objName, namespace,
			),
		},
	))
}

func (ctx RouterContext) ReportNormalError(msg string, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(400)
	LogTemplateError(ctx.LoadTemplate("error").Execute(w,
		templates.ErrorTemplateModel{
			ErrorCode: 400,
			ErrorMessage: fmt.Sprintf(
				"Error: %s",
				msg,
			),
		},
	))
}

func (ctx RouterContext) ReportInternalError(msg string, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(500)
	LogTemplateError(ctx.LoadTemplate("error").Execute(w,
		templates.ErrorTemplateModel{
			ErrorCode: 500,
			ErrorMessage: fmt.Sprintf(
				"Internal error: %s",
				msg,
			),
		},
	))
}

func (ctx RouterContext) ReportForbidden(msg string, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(403)
	LogTemplateError(ctx.LoadTemplate("error").Execute(w,
		templates.ErrorTemplateModel{
			ErrorCode: 500,
			ErrorMessage: fmt.Sprintf(
				"Forbidden: %s",
				msg,
			),
		},
	))
}

func (ctx RouterContext) ReportObjectReadFailure(objid string, msg string, w http.ResponseWriter, r *http.Request) {
	ctx.ReportInternalError(
		fmt.Sprintf(
			"Fail to read object %s: %s",
			objid, msg,
		), w, r,
	)
}

func (ctx RouterContext) ReportObjectTypeMismatch(objid string, expectedType string, actualType string, w http.ResponseWriter, r *http.Request) {
	ctx.ReportInternalError(
		fmt.Sprintf(
			"Object type mismatch for %s: %s expected but %s found",
			objid, expectedType, actualType,
		), w, r,
	)
}

func (ctx *RouterContext) SyncAllNamespace() error {
	if ctx.Config.PlainMode {
		if ctx.Config.UseNamespace {
			rp, err := ctx.Config.GetAllNamespacePlain()
			if err != nil { return err }
			ctx.GitNamespaceList = rp
		} else {
			ns, err := model.NewNamespace("", ctx.Config.GitRoot)
			if err != nil { return err }
			if ctx.GitNamespaceList == nil {
				ctx.GitNamespaceList = make(map[string]*model.Namespace, 0)
			}
			ctx.GitNamespaceList[""] = ns
		}
	} else {
		if ctx.Config.UseNamespace {
			ns, err := ctx.DatabaseInterface.GetAllNamespace()
			if err != nil { return err }
			ctx.GitNamespaceList = ns
		} else {
			ns, err := ctx.DatabaseInterface.GetNamespaceByName("")
			if err != nil { return err }
			ctx.GitNamespaceList[""] = ns
		}
	}
	return nil
}

func (ctx *RouterContext) SyncNamespace(ns *model.Namespace) error {
	if ctx.Config.PlainMode {
		a, err := ctx.Config.GetAllRepositoryByNamespacePlain(ns.Name)
		
		if err != nil { return err }
		ns.RepositoryList = a
	} else {
		a, err := ctx.DatabaseInterface.GetAllRepositoryFromNamespace(ns.Name)
		if err != nil { return err }
		ns.RepositoryList = a
	}
	return nil
}

func (ctx *RouterContext) ResolveRepositoryFullName(str string) (string, string, *model.Repository, error) {
	np := strings.Split(strings.TrimSpace(str), ":")
	namespaceName := ""
	repoName := ""
	if len(np) <= 1 {
		namespaceName = ""
		repoName = np[0]
	} else {
		namespaceName = np[0]
		repoName = np[1]
	}
	s, ok := ctx.GitNamespaceList[namespaceName]
	if !ok {
		err := ctx.SyncAllNamespace()
		if err != nil { return "", "", nil, err }
		s, ok = ctx.GitNamespaceList[namespaceName]
		if !ok {
			return "", "", nil, NewRouteError(
				NOT_FOUND, fmt.Sprintf(
					"Namespace %s not found.", namespaceName,
				),
			)
		}
	}
	err := ctx.SyncNamespace(s)
	if err != nil { return "", "", nil, err }
	rp, ok := s.RepositoryList[repoName]
	if !ok {
		return "", "", nil, NewRouteError(
			NOT_FOUND, fmt.Sprintf(
				"Repository %s not found.", repoName,
			),
		)
	}
	return namespaceName, repoName, rp, nil
}


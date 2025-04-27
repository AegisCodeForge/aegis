package routes

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/bctnry/gitus/pkg/gitlib"
	"github.com/bctnry/gitus/pkg/gitus"
	"github.com/bctnry/gitus/pkg/gitus/model"
	"github.com/bctnry/gitus/templates"
)

type RouterContext struct {
	Config *gitus.GitusConfig
	MasterTemplate *template.Template
	GitRepositoryList map[string]*gitlib.LocalGitRepository
	GitNamespaceList map[string]*model.Namespace
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

func (ctx RouterContext) ResolveRepositoryFullName(str string) (string, string, *gitlib.LocalGitRepository, error) {
	np := strings.Split(str, ":")
	namespaceName := ""
	if len(np) <= 1 {
		s, ok := ctx.GitRepositoryList[str]
		if !ok {
			return "", "", nil, NewRouteError(
				NOT_FOUND, fmt.Sprintf(
					"Repository %s not found.", str,
				),
			)
		}
		return "", str, s, nil
	} else {
		namespaceName = np[0]
		ns, ok := ctx.GitNamespaceList[namespaceName]
		if !ok {
			return "", "", nil, NewRouteError(
				NOT_FOUND, fmt.Sprintf(
					"Namespace %s not found.", namespaceName,
				),
			)
		}
		r, ok := ns.RepositoryList[np[1]]
		if !ok {
			return "", "", nil, NewRouteError(
				NOT_FOUND, fmt.Sprintf(
					"Repository %s not found in namespace %s.", np[1], namespaceName,
				),
			)
		}
		return namespaceName, np[1], r, nil
	}
}


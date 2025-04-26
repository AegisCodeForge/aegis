package controller

import (
	"github.com/bctnry/gitus/routes"
)

func InitializeRoute(context routes.RouterContext) {
	bindBlobController(context)
	bindBranchController(context)
	bindCommitController(context)
	bindDiffController(context)
	bindHistoryController(context)
	bindIndexController(context)
	bindRepositoryController(context)
	bindTagController(context)
	bindTreeHandler(context)

	bindHttpCloneController(context)
	if context.Config.UseNamespace {
		bindNamespaceController(context)
	}
}


package controller

import (
	"github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/routes/controller/admin"
)

func InitializeRoute(context *routes.RouterContext) {
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

	if !context.Config.PlainMode {
		bindUserController(context)
		bindLoginController(context)
		bindLogoutController(context)
		bindSettingController(context)
		bindSettingSSHController(context)
		bindSettingGPGController(context)
		bindNewNamespaceController(context)
		bindNewRepositoryController(context)

		bindRegisterController(context)
		bindReceiptController(context)
		bindConfirmRegistrationController(context)

		// bind admin controller
		admin.BindAdminIndexController(context)
	}
}


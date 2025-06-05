package controller

import (
	"github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/routes/controller/admin"
	"github.com/bctnry/aegis/routes/controller/all"
)

func InitializeRoute(context *routes.RouterContext) {
	bindBlobController(context)
	bindBranchController(context)
	bindCommitController(context)
	bindDiffController(context)
	bindHistoryController(context)
	bindIndexController(context)
	bindRepositoryController(context)
	bindRepositorySettingController(context)
	bindTagController(context)
	bindTreeHandler(context)
	all.BindAllController(context)
	bindHttpCloneController(context)
	if context.Config.UseNamespace {
		bindNamespaceController(context)
		bindNamespaceSettingController(context)
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
		admin.BindAllAdminControllers(context)

		bindResetPasswordController(context)

		bindIssueController(context)
	}
}


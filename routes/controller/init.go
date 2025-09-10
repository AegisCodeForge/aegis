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
	bindAllController(context)
	bindHttpCloneController(context)
	bindShutdownNoticeController(context)
	bindMaintenanceNoticeController(context)
	bindPrivateNoticeController(context)
	
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
		bindSettingEmailController(context)
		bindSettingPrivacyController(context)
		bindRepositorySettingController(context)
		bindNewNamespaceController(context)
		bindNewRepositoryController(context)
		bindNewSnippetController(context)

		bindRegisterController(context)
		bindReceiptController(context)
		bindConfirmRegistrationController(context)
		bindVerifyEmailController(context)

		// bind admin controller
		admin.BindAllAdminControllers(context)

		bindResetPasswordController(context)

		bindIssueController(context)
		bindLabelController(context)

		bindSnippetController(context)
	}
}


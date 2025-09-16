package admin

import (
	"github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/routes/controller/admin/edit_user"
)

func BindAllAdminControllers(context *routes.RouterContext) {
	bindAdminIndexController(context)
	bindAdminSiteConfigController(context)
	bindAdminDatabaseSettingController(context)
	bindAdminSessionSettingController(context)
	bindAdminMailerSettingController(context)
	bindAdminReceiptSystemSettingController(context)
	bindAdminUserListController(context)
	edit_user.BindAdminEditUserController(context)
	bindAdminNewUserController(context)
	bindAdminNamespaceListController(context)
	bindAdminEditNamespaceController(context)
	bindAdminNewNamespaceController(context)
	bindAdminIndexConfigController(context)
	bindAdminRepositoryListController(context)
	bindAdminReceiptListController(context)
	bindAdminSiteLockdownController(context)
	bindAdminRegistrationRequestController(context)
}

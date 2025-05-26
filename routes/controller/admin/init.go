package admin

import (
	"github.com/bctnry/aegis/routes"
)

func BindAllAdminControllers(context *routes.RouterContext) {
	bindAdminIndexController(context)
	bindAdminSiteConfigController(context)
	bindAdminDatabaseSettingController(context)
	bindAdminSessionSettingController(context)
	bindAdminMailerSettingController(context)
	bindAdminReceiptSystemSettingController(context)
	bindAdminUserListController(context)
	bindAdminEditUserController(context)
	bindAdminNewUserController(context)
	bindAdminNamespaceListController(context)
	bindAdminEditNamespaceController(context)
	bindAdminNewNamespaceController(context)
	bindAdminIndexConfigController(context)
	bindAdminRepositoryListController(context)
	bindAdminReceiptListController(context)
}

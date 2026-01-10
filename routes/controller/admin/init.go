package admin

import (
	"github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/routes/controller/admin/edit_user"
	"github.com/bctnry/aegis/routes/controller/admin/rrdoc"
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
	rrdoc.BindAdminRRDocController(context)
	bindAdminNewUserController(context)
	bindAdminNamespaceListController(context)
	bindAdminEditNamespaceController(context)
	bindAdminRepositoryListController(context)
	bindAdminReceiptListController(context)
	bindAdminSiteLockdownController(context)
	bindAdminRegistrationRequestController(context)
}

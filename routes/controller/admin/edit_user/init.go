package edit_user

import (
	. "github.com/bctnry/aegis/routes"
)

func BindAdminEditUserController(ctx *RouterContext) {
	bindAdminEditUserInfoController(ctx)
	bindAdminEditUserGPGController(ctx)
	bindAdminEditUserSSHController(ctx)
}


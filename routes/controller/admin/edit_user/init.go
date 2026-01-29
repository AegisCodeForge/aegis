package edit_user

import (
	. "github.com/GitusCodeForge/Gitus/routes"
)

func BindAdminEditUserController(ctx *RouterContext) {
	bindAdminEditUserInfoController(ctx)
	bindAdminEditUserGPGController(ctx)
	bindAdminEditUserSSHController(ctx)
}


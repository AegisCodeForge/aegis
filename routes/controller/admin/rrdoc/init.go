package rrdoc

import (
	. "github.com/GitusCodeForge/Gitus/routes"
)

func BindAdminRRDocController(ctx *RouterContext) {
	bindAdminRRDocListController(ctx)
	bindAdminRRDocNewController(ctx)
	bindAdminRRDocEditController(ctx)
	bindAdminRRDocDeleteController(ctx)
}


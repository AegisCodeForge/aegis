//go:build ignore

package templates

type LoginInfoModel struct {
	LoggedIn bool
	UserName string
	IsSettingMember bool
	IsAdmin bool
	IsSuperAdmin bool
}


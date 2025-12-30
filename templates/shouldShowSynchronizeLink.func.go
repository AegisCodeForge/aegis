//go:build ignore
package templates

import "github.com/bctnry/aegis/pkg/gitlib"

func(s *LoginInfoModel, ci *gitlib.BranchComparisonInfo) bool {
	if !s.LoggedIn { return false }
	if !s.IsOwner { return false }
	if ci == nil { return true }
	if len(ci.ARevList) > 0 && len(ci.BRevList) <= 0 { return true }
	return false
}


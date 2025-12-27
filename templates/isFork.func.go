//go:build ignore
package templates

import "github.com/bctnry/aegis/pkg/model"

func(s *model.Repository) bool {
	return len(s.ForkOriginNamespace) > 0 && len(s.ForkOriginName) >0
}



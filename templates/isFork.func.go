//go:build ignore
package templates

import "github.com/GitusCodeForge/Gitus/pkg/model"

func(s *model.Repository) bool {
	return len(s.ForkOriginNamespace) > 0 && len(s.ForkOriginName) >0
}



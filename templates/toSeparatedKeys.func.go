//go:build ignore
package templates

import "fmt"
import "strings"

func(a any) string {
	aa, ok := a.(map[string]bool)
	if ok {
		v := make([]string, 0)
		for k := range aa {
			v = append(v, k)
		}
		return strings.Join(v, ",")
	}
	return ""
}


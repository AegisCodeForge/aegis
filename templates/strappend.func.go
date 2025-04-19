//go:build ignore
package templates

import "strings"

func(s ...string)string {
	var res strings.Builder
	for _, item := range s {
		res.WriteString(item)
	}
	return res.String()
}


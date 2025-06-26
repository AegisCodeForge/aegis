//go:build ignore
package templates

import "strings"

func(s1 string, prefix string) bool {
	return strings.HasPrefix(s1, prefix)
}


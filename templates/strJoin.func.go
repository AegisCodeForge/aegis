//go:build ignore
package templates

import "strings"

func(s []string, sep string)string {
	return strings.Join(s, sep)
}


//go:build ignore
package templates

import "html"
import "strings"

// this is used to escape urls in html attribute values.

func(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "\\", "\\\\"), "'", "\\'")
}



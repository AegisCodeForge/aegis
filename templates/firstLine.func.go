//go:build ignore
package templates

import "strings"

func(s string) string {
	a := strings.SplitN(s, "\n", 2)
	if (len(a) >= 1) {
		return a[0]
	} else {
		return ""
	}
}


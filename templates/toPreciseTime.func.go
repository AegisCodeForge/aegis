//go:build ignore
package templates

import "time"

func(s int64) string {
	return time.Unix(s, 0).String()
}


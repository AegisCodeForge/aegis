//go:build ignore
package templates

import "time"

func(s int64, timeoutMinute int64) string {
	return time.Unix(s, 0).Add(time.Duration(timeoutMinute * int64(time.Minute))).String()
}


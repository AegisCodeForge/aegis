//go:build ignore
package templates

import "time"
import "fmt"

func(s interface{}) string {
	timeObj, ok := s.(time.Time)
	if !ok {
		timestamp, ok := s.(int64)
		if !ok {
			panic("Cannot determine time type")
		}
		return time.Unix(timestamp, 0).Format(time.RFC3339)
	} else {
		return timeObj.Format(time.RFC3339)
	}
}


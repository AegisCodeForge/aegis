//go:build ignore
package templates

import "time"

func(s interface{}) string {
	var timestamp int64
	timeObj, ok := s.(time.Time)
	if !ok {
		timestamp, ok = s.(int64)
		if !ok {
			panic("Cannot determine time type")
		}
	} else {
		timestamp = timeObj.Unix()
	}
	return time.Unix(timestamp, 0).String()
}


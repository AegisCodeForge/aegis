//go:build ignore
package templates

import "fmt"
import "time"
import "github.com/GitusCodeForge/Gitus/pkg/fuzzytime"

// this is used to escape urls in html attribute values.

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
	return fuzzytime.TimeSpanToFuzzyTimeString(time.Now().Sub(time.Unix(timestamp, 0)))
}



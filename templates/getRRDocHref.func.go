//go:build ignore
package templates

import "fmt"
import "strings"

func(s string) string {
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return s
	} else {
		return fmt.Sprintf("/rrdoc/%s", s)
	}
}




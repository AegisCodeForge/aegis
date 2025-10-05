//go:build ignore
package templates

import "fmt"

func(userName string) string {
	return fmt.Sprintf("/u/%s", userName)
}


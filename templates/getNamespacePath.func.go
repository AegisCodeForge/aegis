//go:build ignore
package templates

import "fmt"

func(namespaceName string) string {
	return fmt.Sprintf("/s/%s", namespaceName)
}


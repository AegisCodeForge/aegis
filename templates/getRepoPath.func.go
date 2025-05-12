//go:build ignore
package templates

import "fmt"

func(namespaceName string, repoName string) string {
	return fmt.Sprintf("/repo/%s:%s", namespaceName, repoName)
}


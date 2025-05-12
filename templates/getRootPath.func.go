//go:build ignore
package templates

import "fmt"

func(namespaceName string, repoName string, typeStr string, nodeName string) string {
	return fmt.Sprintf("/repo/%s:%s/%s/%s", namespaceName, repoName, typeStr, nodeName)
}


//go:build ignore
package templates

import "fmt"

func(namespaceName string, repoName string, typeStr string, nodeName string) string {
	rfn := repoName
	if len(namespaceName) > 0 { rfn = namespaceName + ":" + repoName }
	return fmt.Sprintf("/repo/%s/%s/%s", rfn, typeStr, nodeName)
}


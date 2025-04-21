//go:build ignore
package templates

import "fmt"

func(repoName string, typeStr string, nodeName string) string {
	return fmt.Sprintf("/repo/%s/%s/%s", repoName, typeStr, nodeName)
}


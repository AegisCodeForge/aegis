//go:build ignore
package templates

import "fmt"

func(namespaceName string, repoName string) string {
	rfn := repoName
	if len(namespaceName) > 0 { rfn = namespaceName + ":" + repoName }
	return rfn
}


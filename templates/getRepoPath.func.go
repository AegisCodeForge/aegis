//go:build ignore
package templates

import "fmt"

func(repoName string) string {
	return fmt.Sprintf("/repo/%s", repoName)
}


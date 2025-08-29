//go:build ignore
package templates

import "fmt"

func(m map[string]string, email string) string {
	v, ok := m[email]
	if !ok { return fmt.Sprintf("mailto:%s", email) }
	if len(v) <= 0 { return fmt.Sprintf("mailto:%s", email) }
	return fmt.Sprintf("/u/%s", v)
}


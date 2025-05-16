//go:build ignore
package templates

import "fmt"

func(a... int) int {
	var x int = a[0]
	for i, item := range a {
		if i == 0 { continue }
		x = x - item
	}
	return x
}


//go:build ignore
package templates

import "fmt"

func(a... int64) int64 {
	var x int64 = a[0]
	for i, item := range a {
		if i == 0 { continue }
		x = x - item
	}
	return x
}


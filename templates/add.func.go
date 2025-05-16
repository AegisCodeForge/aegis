//go:build ignore
package templates

import "fmt"

func(a... int) int {
	var x int = 0
	for _, item := range a { x = x + item }
	return x
}


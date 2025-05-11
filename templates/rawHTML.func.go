//go:build ignore
package templates

import ht "html/template"

func(s string) ht.HTML {
	return ht.HTML(s)
}


//go:build ignore
package templates

import "strings"

func(s string) BlobTextTemplateModel {
	return BlobTextTemplateModel{
		FileLineCount: len(strings.Split(s, "\n")),
		FileContent: s,
	}
}


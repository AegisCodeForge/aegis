//go:build ignore
package templates

import "encoding/json"
import "github.com/bctnry/aegis/pkg/gitlib"

func(s string) *gitlib.MergeCheckResult {
	var r *gitlib.MergeCheckResult
	json.Unmarshal([]byte(s), &r)
	return r
}


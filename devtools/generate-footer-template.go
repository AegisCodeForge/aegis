//go:build ignore

package main

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
)

// this is only for including the version info (determined w/ VERSION
// file and latest git commit id) into the _footer template. it has no
// other meaningful use.

func main() {
	versionBytes, _ := os.ReadFile("VERSION")
	versionStr := strings.TrimSpace(string(versionBytes))
	countBytes, _ := os.ReadFile("COUNT")
	countStr := strings.Split(string(countBytes), ",")
	var newCountStr string
	var newVersionStr string
	if versionStr != strings.TrimSpace(countStr[0]) {
		newCountStr = fmt.Sprintf("%s,%d", versionStr, 0)
		newVersionStr = fmt.Sprintf("%s.build_%d", versionStr, 0)
	} else {
		c, _ := strconv.Atoi(countStr[1])
		newCountStr = fmt.Sprintf("%s,%d", versionStr, c+1)
		newVersionStr = fmt.Sprintf("%s.build_%d", versionStr, c+1)
	}
	p := path.Join("templates", "_footer.template.html")
	f, _ := os.OpenFile(p, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if f != nil {
		f.WriteString(fmt.Sprintf(`
{{define "_footer"}}
<div class="footer-message">
    Powered by <a href="https://github.com/GitusCodeForge/Gitus">Gitus</a>, version %s (%s)
</div>
{{end}}`, string(versionBytes), newVersionStr))
		f.Close()
	}
	f2, _ := os.OpenFile("COUNT", os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	f2.WriteString(newCountStr)
	f2.Close()
}


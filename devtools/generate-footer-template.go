//go:build ignore

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
)

// this is only for including the version info (determined w/ VERSION
// file and latest git commit id) into the _footer template. it has no
// other meaningful use.

func main() {
	versionBytes, _ := os.ReadFile("VERSION")
	cmd := exec.Command("git", "rev-list", "HEAD", "-1")
	stdoutBuf := new(bytes.Buffer)
	cmd.Stdout = stdoutBuf
	cmd.Run()
	p := path.Join("templates", "_footer.template.html")
	f, _ := os.OpenFile(p, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if f != nil {
		f.WriteString(fmt.Sprintf(`
{{define "_footer"}}
<div class="footer-message">
    Powered by <a href="https://github.com/bctnry/aegis">Aegis</a>, version %s (%s)
</div>
{{end}}`, string(versionBytes), strings.TrimSpace(stdoutBuf.String())))
	}
}


package gitlib

import (
	"bytes"
	"os/exec"
)

// a wrapper over git-rev-list.
// TODO: find a better way to do this...
func (gr LocalGitRepository) ResolvePathLastCommitId(cobj *CommitObject, p string) (string, error) {
	cmd := exec.Command("git", "rev-list", "-1", cobj.Id, "--", p)
	cmd.Dir = gr.GitDirectoryPath
	buf := new(bytes.Buffer)
	cmd.Stdout = buf
	err := cmd.Run()
	if err != nil { return "", err }
	return buf.String(), nil
}



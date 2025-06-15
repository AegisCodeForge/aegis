package gitlib

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
)

type DiffItem struct {
	
}

// returns the diff by invoking the "git diff" command.
// i should probably implement my own diff and rely as little on the
// git executable as possible at some point.
var ErrDubiousOwnership = errors.New("dubious ownership")
func (gr LocalGitRepository) GetDiff(commitId string) (string, error) {
	cmd := exec.Command("git", "diff-tree", commitId, "-p")
	cmd.Dir = gr.GitDirectoryPath
	stderrBuf := new(bytes.Buffer)
	cmd.Stderr = stderrBuf
	out, err := cmd.Output()
	if err != nil {
		if strings.Contains(stderrBuf.String(), "dubious ownership") {
			return "", ErrDubiousOwnership
		}
		return "", err
	}
	return string(out), nil
}



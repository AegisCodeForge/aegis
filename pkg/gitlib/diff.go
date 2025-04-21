package gitlib

import "os/exec"

// returns the diff by invoking the "git diff" command.
// i should probably implement my own diff and rely as little on the
// git executable as possible at some point.
func (gr LocalGitRepository) GetDiff(commitId string) (string, error) {
	cmd := exec.Command("git", "diff-tree", commitId, "-p")
	cmd.Dir = gr.GitDirectoryPath
	out, err := cmd.Output()
	if err != nil { return "", err }
	return string(out), nil
}



package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/bctnry/aegis/pkg/gitlib"
	"github.com/bctnry/aegis/pkg/shellparse"
	"github.com/bctnry/aegis/routes"
)

// `aegis ssh` handler.

func printGitError(s string) {
	fmt.Print(gitlib.ToPktLine(fmt.Sprintf("ERR %s\n", s)))
}

func HandleSSHLogin(ctx *routes.RouterContext, username string, keyname string) {
	m, err := ctx.DatabaseInterface.GetAuthKeyByName(username, keyname)
	if err != nil {
		printGitError(err.Error())
		os.Exit(1)
	}
	authorizedKey := ctx.SSHKeyManagingContext.GetAuthorizedKey(username, keyname)
	if authorizedKey != m.KeyText {
		printGitError(fmt.Sprintf("Integrity check failed:\n auth: %s\nkt: %s", authorizedKey, m.KeyText))
		os.Exit(1)
	}
	origCmd := os.Getenv("SSH_ORIGINAL_COMMAND")
	// one might be tempted to think that one can just pass SSH_ORIGINAL_COMMAND
	// to exec.Command, but things don't work that way...
	parsedOrigCmd := shellparse.ParseShellCommand(origCmd)
	// see also:
	//     https://git-scm.com/docs/git-receive-pack
	//     https://git-scm.com/docs/git-upload-pack
	//     https://git-scm.com/docs/git-upload-archive
	// all commands have the git dir path at the end of the call, so we resolve it
	// with ctx.Config.
	realGitPath := path.Join(ctx.Config.GitRoot, parsedOrigCmd[len(parsedOrigCmd)-1])
	parsedOrigCmd[len(parsedOrigCmd)-1] = realGitPath
	cmdobj := exec.Command(parsedOrigCmd[0], parsedOrigCmd[1:]...)

	// r, w := io.Pipe()
	cmdobj.Stdout = os.Stdout
	cmdobj.Stdin = os.Stdin
	cmdobj.Stderr = os.Stderr
	err = cmdobj.Run()
	if err != nil {
		printGitError(err.Error())
	}
	os.Exit(0)
}


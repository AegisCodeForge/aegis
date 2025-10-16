package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/model"
	"github.com/bctnry/aegis/pkg/gitlib"
	"github.com/bctnry/aegis/pkg/shellparse"
	"github.com/bctnry/aegis/routes"
)

// `aegis ssh` handler.

func printGitError(s string) {
	fmt.Print(gitlib.ToPktLine(fmt.Sprintf("ERR %s\n", s)))
}

func HandleSSHLogin(ctx *routes.RouterContext, username string, keyname string) {
	if ctx.Config.GlobalVisibility != aegis.GLOBAL_VISIBILITY_PUBLIC &&
		ctx.Config.GlobalVisibility != aegis.GLOBAL_VISIBILITY_PRIVATE {
		printGitError("This instance of Aegis is currently unavailable.")
		os.Exit(1)
	}
	if ctx.Config.IsInPlainMode() {
		printGitError("This instance of Aegis is in Plain Mode which does not allow Git over SSH.")
		os.Exit(1)
	}
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
	isPushingToRemote := parsedOrigCmd[0] == "git-receive-pack"
	relPath := parsedOrigCmd[len(parsedOrigCmd)-1]
	if relPath[0] == '~' || relPath[0] == '/' {
		relPath = relPath[1:]
	}
	relPathSegment := strings.SplitN(relPath, "/", 2)
	namespaceName := ""
	repositoryName := ""
	if ctx.Config.UseNamespace {
		if len(relPathSegment) <= 1 {
			relPathSegment = strings.SplitN(relPath, ":", 2)
			if len(relPathSegment) <= 1 {
				printGitError("Invalid repository path specification.")
				os.Exit(1)
			}
			namespaceName = relPathSegment[0]
			repositoryName = relPathSegment[1]
		}
		namespaceName = relPathSegment[0]
		repositoryName = relPathSegment[1]
	} else {
		if len(relPathSegment) > 1 {
			printGitError("Invalid repository path specification.")
			os.Exit(1)
		}
		namespaceName = ""
		repositoryName = relPathSegment[0]
	}

	// check acl.
	r, err := ctx.DatabaseInterface.GetRepositoryByName(namespaceName, repositoryName)
	if err != nil {
		printGitError(fmt.Sprintf("Failed while reading ACL: %s.", err.Error()))
		os.Exit(1)
	}
	if r.Status == model.REPO_ARCHIVED && isPushingToRemote {
		printGitError(fmt.Sprintf("The repository %s/%s is ARCHIVED; no push to remote is allowed. ", namespaceName, repositoryName))
		os.Exit(1)
	}
	ns, err := ctx.DatabaseInterface.GetNamespaceByName(namespaceName)
	if err != nil {
		printGitError(fmt.Sprintf("Failed while reading namespace: %s.", err.Error()))
		os.Exit(1)
	}
	if r.Owner != username && ns.Owner != username {
		aclt, ok := r.AccessControlList.ACL[username]
		if !ok {
			aclt, ok = ns.ACL.ACL[username]
			if !ok {
				printGitError("Not enough permission.")
				os.Exit(1)
			}
		}
		if !aclt.PushToRepository && isPushingToRemote {
			printGitError("Not enough permission.")
			os.Exit(1)
		}
	}

	// see also:
	//     https://git-scm.com/docs/git-receive-pack
	//     https://git-scm.com/docs/git-upload-pack
	//     https://git-scm.com/docs/git-upload-archive
	// all commands have the git dir path at the end of the call, so we resolve it
	// with ctx.Config.
	realGitPath := path.Join(ctx.Config.GitRoot, r.Namespace, r.Name)
	parsedOrigCmd[len(parsedOrigCmd)-1] = realGitPath
	cmdobj := exec.Command(parsedOrigCmd[0], parsedOrigCmd[1:]...)

	cmdobj.Stdout = os.Stdout
	cmdobj.Stdin = os.Stdin
	cmdobj.Stderr = os.Stderr
	err = cmdobj.Run()
	if err != nil {
		printGitError(err.Error())
	}
	os.Exit(0)
}

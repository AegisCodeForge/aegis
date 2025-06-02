package gitlib

import (
	"os"
	"path"
)

var HookList = []string{
	"applypatch-msg",
	"pre-applypatch",
	"post-applypatch",
	"pre-commit",
	"pre-merge-commit",
	"prepare-commit-msg",
	"commit-msg",
	"post-commit",
	"pre-rebase",
	"post-checkout",
	"post-merge",
	"pre-push",
	"pre-receive",
	"update",
	"proc-receive",
	"post-receive",
	"post-update",
	"reference-transaction",
	"push-to-checkout",
	"pre-auto-gc",
	"post-rewrite",
	"rebase",
	"sendemail-validate",
	"fsmonitor-watchman",
	"p4-changelist",
	"p4-prepare-changelist",
	"p4-post-changelist",
	"p4-pre-submit",
	"post-index-change",
}

func (lgr LocalGitRepository) GetAllSetHooksName() ([]string, error) {
	res := make([]string, 0)
	for _, item := range HookList {
		p := path.Join(lgr.GitDirectoryPath, "hooks", item)
		f, err := os.Open(p)
		if err != nil { continue }
		res = append(res, item)
		f.Close()
	}
	return res, nil
}

func (lgr LocalGitRepository) GetHook(hookName string) (string, error) {
	p := path.Join(lgr.GitDirectoryPath, "hooks", hookName)
	s, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) { return "", nil }
		return "", err
	}
	return string(s), nil
}

func (lgr LocalGitRepository) SaveHook(hookName string, hookContent string) error {
	p := path.Join(lgr.GitDirectoryPath, "hooks", hookName)
	f, err := os.OpenFile(p, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0764)
	if err != nil { return err }
	defer f.Close()
	_, err = f.WriteString(hookContent)
	if err != nil { return err }
	return nil
}

func (lgr LocalGitRepository) DeleteHook(hookName string) error {
	if lgr.Hooks == nil { lgr.Hooks = make(map[string]string, 0) }
	delete(lgr.Hooks, hookName)
	p := path.Join(lgr.GitDirectoryPath, "hooks", hookName)
	return os.Remove(p)
}


package gitlib

import (
	"os"
	"path"

	"github.com/bctnry/aegis/pkg/ini"
)

type SubmoduleConfig struct {
	InRepoPath string
	Url string
	Branch string
	Update string
}

func (lgr *LocalGitRepository) LoadSubmoduleConfig() error {
	p := path.Join(lgr.GitDirectoryPath, ".gitmodules")
	f, err := os.Open(p)
	if err == os.ErrNotExist { lgr.Submodule = nil; return nil }
	if err != nil { return err }
	i, err := ini.ParseINI(f)
	if err != nil { return err }
	sms, ok := i.GetSectionList("submodule")
	if !ok { lgr.Submodule = nil; return nil }
	res := make(map[string]*SubmoduleConfig, 0)
	for _, v := range sms {
		spath := v.Value["path"]
		surl := v.Value["url"]
		sbranch, ok := v.Value["branch"]
		if !ok { sbranch = "" }
		supdate, ok := v.Value["update"]
		if !ok { supdate = "" }
		res[v.SubName] = &SubmoduleConfig{
			InRepoPath: spath,
			Url: surl,
			Branch: sbranch,
			Update: supdate,
		}
	}
	lgr.Submodule = res
	return nil
}


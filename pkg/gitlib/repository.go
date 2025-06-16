package gitlib

import (
	"compress/zlib"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/bctnry/aegis/pkg/ini"
)

type LocalGitRepository struct {
	Namespace string
	Name string
	GitDirectoryPath string
	Description string
	PackIndex map[string]*PackIndex
	BranchIndex map[string]*Branch
	TagIndex map[string]*Tag
	Config ini.INI
	isSHA256 bool
	Hooks map[string]string
	Submodule map[string]*SubmoduleConfig
}

func (gr LocalGitRepository) IsSHA256() bool {
	return gr.isSHA256
}

func (gr LocalGitRepository) FullName() string {
	if len(gr.Namespace) > 0 {
		return fmt.Sprintf("%s:%s", gr.Namespace, gr.Name)
	} else {
		return gr.Name
	}
}

func NewLocalGitRepository(namespace string, name string, p string) *LocalGitRepository {
	res := LocalGitRepository{
		Namespace: namespace,
		Name: name,
		GitDirectoryPath: p,
		PackIndex: nil,
		Hooks: nil,
	}
	pi, err := res.readAllPackIndex()
	if err != nil { log.Fatal(err) }
	res.PackIndex = pi
	description, err := res.readDescription()
	if err != nil { description = "Error due to: " + err.Error() }
	res.Description = description
	config, err := res.readConfig()
	if err != nil {
		res.Config = nil
	} else {
		res.Config = config
	}
	if config == nil {
		res.isSHA256 = false
	} else {
		version, ok1 := config.GetValue("core", "", "repositoryformatversion")
		format, ok2 := config.GetValue("extensions", "", "objectformat")
		if ok1 && ok2 && version == "1" && strings.ToLower(format) == "sha256" {
			res.isSHA256 = true
		} else {
			res.isSHA256 = false
		}
	}
	res.LoadSubmoduleConfig()
	cmd := exec.Command("git", "update-server-info")
	cmd.Dir = p
	// ignore error for now.
	cmd.Run()
	return &res
}

func (gr LocalGitRepository) SyncLocalDescription() error {
	descFilePath := path.Join(gr.GitDirectoryPath, "description")
	os.Remove(descFilePath)
	f, err := os.OpenFile(descFilePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0664)
	if err != nil { return err }
	defer f.Close()
	f.Write([]byte(gr.Description))
	return nil
}

func (gr LocalGitRepository) readConfig() (ini.INI, error) {
	configFilePath := path.Join(gr.GitDirectoryPath, "config")
	f, err := os.Open(configFilePath)
	if err != nil { return nil, err }
	defer f.Close()
	return ini.ParseINI(f)
}

func (gr LocalGitRepository) readDescription() (string, error) {
	descriptionFilePath := path.Join(gr.GitDirectoryPath, "description")
	f, err := os.Open(descriptionFilePath)
	if err != nil { return "", err }
	defer f.Close()
	s, err := io.ReadAll(f)
	if err != nil { return "", err }
	return string(s), nil
}

func (gr LocalGitRepository) OpenDirectlyAccessibleObject(h string) (io.ReadCloser, error) {
	objStorePath := path.Join(gr.GitDirectoryPath, "objects")
	objPath := path.Join(objStorePath, h[:2], h[2:])
	f, err := os.Open(objPath)
	if err != nil { return nil, err }
	return zlib.NewReader(f)
}


// two places to check:
//     refs/heads/*,  packed-refs
func (gr LocalGitRepository) GetAllBranchList() (map[string]*Branch, error) {
	var res map[string]*Branch = make(map[string]*Branch)
	rplocal := path.Join(gr.GitDirectoryPath, "refs", "heads")
	ls, err := os.ReadDir(rplocal)
	if err != nil { return nil, err }
	if len(ls) > 0 {
		for _, item := range ls {
			if item.IsDir() { continue }
			name := item.Name()
			f, err := os.Open(path.Join(rplocal, name))
			if err != nil { return nil, err }
			defer f.Close()
			headIdBytes, err := io.ReadAll(f)
			if err != nil { return nil, err }
			headId := strings.TrimSpace(string(headIdBytes))
			res[name] = &Branch{
				Name: name,
				HeadId: headId,
			}
		}
	}
	pi, err := gr.readPackedRefIndex()
	if err != nil { return nil, err }
	if len(pi) > 0 {
		for _, item := range pi {
			// skip tags when counting branches.
			if !strings.HasPrefix(item.Name, "refs/heads/") { continue }
			name := item.Name[len("refs/heads/"):]
			res[name] = &Branch{
				Name: name,
				HeadId: item.Id,
			}
		}
	}
	return res, nil
}

func (gr *LocalGitRepository) SyncAllBranchList() error {
	br, err := gr.GetAllBranchList()
	if err != nil { return err }
	gr.BranchIndex = br
	return nil
}

// two places to check:
//     refs/tags/*, packed-refs
// afaik fetched remote tags will only appear in packed-refs with
// its id points to the commit object instead of a separate tag
// object like the tags you make locally.
func (gr LocalGitRepository) GetAllTagList() (map[string]*Tag, error) {
	var res map[string]*Tag = make(map[string]*Tag)
	rplocaltags := path.Join(gr.GitDirectoryPath, "refs", "tags")
	ls, err := os.ReadDir(rplocaltags)
	if err != nil { return nil, err }
	if len(ls) > 0 {
		for _, item := range ls {
			if item.IsDir() { continue }
			name := item.Name()
			f, err := os.Open(path.Join(rplocaltags, name))
			if err != nil { return nil, err }
			defer f.Close()
			headIdBytes, err := io.ReadAll(f)
			if err != nil { return nil, err }
			headId := strings.TrimSpace(string(headIdBytes))
			res[name] = &Tag{
				Name: name,
				HeadId: headId,
			}
		}
	}
	pi, err := gr.readPackedRefIndex()
	if err != nil { return nil, err }
	if len(pi) > 0 {
		for _, item := range pi {
			if !strings.HasPrefix(item.Name, "refs/tags/") { continue }
			name := item.Name[len("refs/tags/"):]
			res[name] = &Tag{
				Name: name,
				HeadId: item.Id,
			}
		}
	}
	return res, nil
}

func (gr *LocalGitRepository) SyncAllTagList() error {
	tl, err := gr.GetAllTagList()
	if err != nil { return err }
	gr.TagIndex = tl
	return nil
}

func (gr LocalGitRepository) GetAllObjectId() ([]string, error) {
	var res []string
	op := path.Join(gr.GitDirectoryPath, "objects")
	entryList, err := os.ReadDir(op)
	if err != nil { return nil, err }
	for _, item := range entryList {
		if !item.IsDir() { continue }
		name := item.Name()
		if name == "info" || name == "pack" { continue }
		np := path.Join(op, item.Name())
		subEntryList, err := os.ReadDir(np)
		if err != nil { return nil, err }
		for _, item2 := range subEntryList {
			res = append(res, name + item2.Name())
		}
	}
	return res, nil
}

func (gr LocalGitRepository) GetCommitHistory(cid string) ([]CommitObject, error) {
	return gr.GetCommitHistoryN(cid, 0)
}

func (gr LocalGitRepository) GetCommitHistoryN(cid string, n int) ([]CommitObject, error) {
	var res []CommitObject
	subjId := cid
	limitless := false
	if n <= 0 { limitless = true }
	for (subjId != "" && (limitless || n > 0)) {
		obj, err := gr.ReadObject(subjId)
		if obj.Type() != COMMIT { return nil, err }
		c := obj.(*CommitObject)
		res = append(res, *c)
		subjId = c.ParentId
		if !limitless { n -= 1 }
	}
	return res, nil
}

func (gr LocalGitRepository) GetBranchCommitHistory(b Branch) ([]CommitObject, error) {
	return gr.GetCommitHistory(b.HeadId)
}

func (gr LocalGitRepository) GetBranchCommitHistoryN(b Branch, n int) ([]CommitObject, error) {
	return gr.GetCommitHistoryN(b.HeadId, n)
}


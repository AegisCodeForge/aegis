package gitus

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/bctnry/gitus/pkg/gitlib"
	"github.com/bctnry/gitus/pkg/gitus/model"
)

type GitusConfig struct {
	filePath string
	// the version of the configuration file. currently only 0 is
	// allowed.
	Version int `json:"version"`
	// specify the root directory where all the `.git` directories
	// will reside.
	GitRoot string `json:"root"`
	GitUser string `json:"gitUser"`
	// whether to enable namespace or not.  this actually affects how
	// gitus store / search for existing .git repository: when this
	// field is set to true, the first level directories after GitRoot
	// will never be considered as git repository. e.g. this would be
	// the situation if useNamespace is false:
	//
	//   GitRoot/a.git   ---> valid repo (name "a")
	//   GitRoot/a/.git  ---> valid repo (name "a")
	//   GitRoot/a       ---> not a repo if it itself is not
	//                        a valid git folder or does not
	//                        contain a valid .git folder.
	//   GitRoot/a/b.git ---> "a" is not a valid repo name, same as "b" (!)
	//                        the repo at `b.git` can only be recognized
	//                        if GitRoot is set as `{oldGitRoot}/a`.
	//   GitRoot/a/b/.git ---> same as above.
	//
	// but if useNamespace is true, then this would be the case:
	//
	//   GitRoot/a.git   ---> not a recognized repo (!) and not a namespace.
	//                        gitus namespace name cannot contain period.
	//   GitRoot/a/.git  ---> namespace name "a" but not a recognized repo
	//                        since the name for repos in gitus is the
	//                        part before ".git" and this part cannot be
	//                        empty.
	//   GitRoot/a       ---> a namespace (name "a")
	//   GitRoot/a/b.git ---> namespace "a", repo "b".
	//                        (fullName would be "a:b")
	//   GitRoot/a/b/.git ---> same as above.
	//   GitRoot/xy/cde.git ---> namespace "xy", repo "cde".
	//                           (fullName would be "xy:cde").
	UseNamespace bool `json:"enableNamespace"`
	// setting a gitus instance to be in plain mode will completely
	// remove all the functionalities that isn't built-in to git; this
	// includes things like issue tracking and signature verification.
	// in plain mode, gitus is basically like git instaweb.
	PlainMode bool `json:"plainMode"`
	// when set to true, this field allow user registration.
	AllowRegistration bool `json:"enableUserRegistration"`
	// when set to true, all registration must be screened by the webmaster.
	ManualApproval bool `json:"requireManualApproval"`

	// cosmetic things...
	
	// the name of the depot (i.e. the top level of the site)
	DepotName string `json:"depotName"`
	StaticAssetDirectory string `json:"staticAssetDirectory"`

	// http host name. (NOTE: no slash at the end.)
	HttpHostName string `json:"hostName"`

	// ssh host name. (NOTE: no slash at the end.)
	SshHostName string `json:"sshHostName"`

	BindAddress string `json:"bindAddress"`
	BindPort int `json:"bindPort"`

	// namespaces you need gitus to ignore during initial searching.
	// only valid when plain mode is enabled. (when plain mode is
	// disabled, all namespaces are visible by public by default,
	// even if they don't have any public repository and/or member.
	IgnoreNamespace []string `json:"ignoreNamespace"`
	// repositories you need gitus to ignore during initial searching.
	// only valid when plain mode is enabled. this option is valid
	// whether you use namespace or not. when useNamespace is true,
	// you need to specify the "full name" of the repository ("full
	// name" i.e. `{namespace}:{repoName}`)
	IgnoreRepository []string `json:"ignoreRepository"`

	// the following database-related options are ignored when plain
	// mode is enabled,
	
	// database type. currently only support "sqlite".
	DatabaseType string `json:"dbType"`
	// path to the database file. valid only when dbtype is sqlite;
	// has no effect otherwise.
	DatabasePath string `json:"dbPath"`
	// url to the database. valid only when dbtype is something that
	// is "hosted" as a server (unlike sqlite which is just one file).
	// has no effect when dbtype is sqlite.
	DatabaseURL string `json:"dbUrl"`
	DatabaseUser string `json:"dbUser"`
	// name of the database. valid only when dbtype is something like
	// "postgre" or "mariadb". has no effect when dbtype is sqlite.
	DatabaseName string `json:"dbName"`
	// password of the database. valid only when dbtype is something
	// like "postgre" or "mariadb". has no effect when dbtype is
	// sqlite.
	DatabasePassword string `json:"dbPassword"`
	// table prefix of the database - in case you need to host
	// multiple gitus instance with the same database or you need
	// to make your gitus instance to share a database with other
	// applications.
	DatabaseTablePrefix string `json:"dbTablePrefix"`

	// session type. currently only support "sqlite".
	// planned support includes "redis", "memcached" and  "rocksdb" in the future.
	SessionType string `json:"sessionType"`
	// session path. valid only when sessiontype is sqlite.
	SessionPath string `json:"sessionPath"`
	// session path. valid only when session type is redis.
	SessionURL string `json:"sessionUrl"`
}

func CreateConfigFile(p string) error {
	f, err := os.OpenFile(
		p,
		os.O_CREATE|os.O_EXCL|os.O_WRONLY,
		0644,
	)
	if err != nil { return err }
	defer f.Close()
	marshalRes, err := json.MarshalIndent(GitusConfig{
		Version: 0,
		GitRoot: "",
		GitUser: "git",
		UseNamespace: false,
		PlainMode: true,
		AllowRegistration: true,
		ManualApproval: true,
		DepotName: "Gitus",
		StaticAssetDirectory: "static/",
		BindAddress: "127.0.0.1",
		BindPort: 8000,
		IgnoreNamespace: nil,
		IgnoreRepository: nil,
		DatabaseType: "sqlite",
		DatabasePath: "",
		DatabaseURL: "",
		DatabaseUser: "",
		DatabaseName: "",
		DatabasePassword: "",
		DatabaseTablePrefix: "gitus_",
	}, "", "    ")
	if err != nil { return err }
	f.Write(marshalRes)
	return nil
}

func LoadConfigFile(p string) (*GitusConfig, error) {
	s, err := os.ReadFile(p)
	if err != nil { return nil, err }
	var c GitusConfig
	err = json.Unmarshal(s, &c)
	if err != nil { return nil, err }
	c.filePath = p
	return &c, nil
}

func (cfg *GitusConfig) Sync() error {
	p := cfg.filePath
	s, err := json.Marshal(cfg)
	if err != nil { return err }
	st, err := os.Stat(p)
	if err != nil { return err }
	f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE, st.Mode())
	if err != nil { return err }
	defer f.Close()
	_, err = f.Write(s)
	if err != nil { return err }
	err = f.Sync()
	if err != nil { return err }
	return nil
}

func (cfg *GitusConfig) GetAllRepositoryPlain() (map[string]*model.Repository, error) {
	// NOTE: you should NOT call this when UseNamespace is true.
	// this would return nil when UseNamespace is true, just in case
	// people depend on its behaviour otherwise.
	if cfg.UseNamespace { return nil, nil }
	gitPath := cfg.GitRoot
	res := make(map[string]*model.Repository, 0)
	l, err := os.ReadDir(gitPath)
	if err != nil { return nil, err }
	for _, item := range l {
		repoName := item.Name()
		p := path.Join(gitPath, item.Name())
		if !gitlib.IsValidGitDirectory(p) {
			p = path.Join(gitPath, item.Name(), ".git")
		}
		if !gitlib.IsValidGitDirectory(p) {
			continue
		}
		if strings.HasSuffix(repoName, ".git") {
			repoName = repoName[:len(repoName)-len(".git")]
			if len(repoName) <= 0 { continue }
		}
		k := gitlib.NewLocalGitRepository("", repoName, p)
		res[repoName] = &model.Repository{
			Namespace: k.Namespace,
			Name: k.Name,
			Description: k.Description,
			AccessControlList: "",
			Status: model.REPO_NORMAL_PUBLIC,
			Repository: k,
		}
	}
	return res, nil
}

func (cfg *GitusConfig) GetAllRepositoryByNamespacePlain(ns string) (map[string]*model.Repository, error) {
	gitPath := cfg.GitRoot
	res := make(map[string]*model.Repository, 0)
	nsPath := path.Join(gitPath, ns)
	l, err := os.ReadDir(nsPath)
	if err != nil { return nil, err }
	for _, item := range l {
		repoName := item.Name()
		p := path.Join(nsPath, item.Name())
		if !gitlib.IsValidGitDirectory(p) {
			p = path.Join(nsPath, item.Name(), ".git")
		}
		if !gitlib.IsValidGitDirectory(p) {
			continue
		}
		if strings.HasSuffix(repoName, ".git") {
			repoName = repoName[:len(repoName)-len(".git")]
			if len(repoName) <= 0 { continue }
		}
		k := gitlib.NewLocalGitRepository("", repoName, p)
		res[repoName] = &model.Repository{
			Namespace: k.Namespace,
			Name: k.Name,
			Description: k.Description,
			AccessControlList: "",
			Status: model.REPO_NORMAL_PUBLIC,
			Repository: k,
		}
	}
	return res, nil
}

func (cfg *GitusConfig) GetAllNamespacePlain() (map[string]*model.Namespace, error) {
	res := make(map[string]*model.Namespace, 0)
	if !cfg.UseNamespace {
		ns, err := model.NewNamespace("", cfg.GitRoot)
		if err != nil { return nil, err }
		for _, item := range cfg.IgnoreRepository {
			k := strings.Split(item, ":")
			if len(k) >= 2 {
				if k[0] != "" { continue }
				delete(ns.RepositoryList, k[1])
			} else {
				delete(ns.RepositoryList, k[0])
			}
		}
		res[""] = ns
		return res, nil
	}
	l, err := os.ReadDir(cfg.GitRoot)
	if err != nil { return nil, err }
	for _, item := range l {
		namespaceName := item.Name()
		if !model.ValidNamespaceName(namespaceName) { continue }
		_, shouldIgnore := slices.BinarySearch(cfg.IgnoreNamespace, namespaceName)
		if shouldIgnore { continue }
		p := path.Join(cfg.GitRoot, namespaceName)
		ns, err := model.NewNamespace(namespaceName, p)
		if err != nil { return nil, err }
		// (i'm worried that) this might be slow...
		for _, item := range cfg.IgnoreRepository {
			k := strings.Split(item, ":")
			if len(k) < 2 { continue }
			if k[0] != namespaceName { continue }
			delete(ns.RepositoryList, k[1])
		}
		res[namespaceName] = ns
	}
	return res, nil
}


//go:build ignore
package templates

import "encoding/json"
import "path"
import "github.com/bctnry/aegis/pkg/gitlib"
import "github.com/bctnry/aegis/pkg/aegis"
import "github.com/bctnry/aegis/pkg/aegis/model"

func(cfg *aegis.AegisConfig, s string) *model.PullRequestCommentOnCode {
	var r *model.PullRequestCommentOnCode
	json.Unmarshal([]byte(s), &r)
	p := path.Join(cfg.GitRoot, r.RepoNamespace, r.RepoName)
	localRepo := gitlib.NewLocalGitRepository(r.RepoNamespace, r.RepoName, p)
	gobj, err := localRepo.ReadObject(r.CommitId)
	if err != nil { log.Fatal(err.Error()) }
	if gobj.Type() != gitlib.COMMIT {
		log.Fatalf("Error: %s isn't a commit", r.CommitId)
	}
	cobj := gobj.(*gitlib.CommitObject)
	gobj, err = localRepo.ReadObject(cobj.TreeObjId)
	if err != nil { log.Fatal(err.Error()) }
	if gobj.Type() != gitlib.COMMIT {
		log.Fatalf("Error: %s isn't a commit", r.CommitId)
	}
	tobj := gobj.(*gitlib.TreeObject)
	gobj, err = localRepo.ResolveTreePath(tobj, r.Path)
	bobj := gobj.(*gitlib.BlobObject)
	lines := strings.Split(string(bobj.Data), "\n")
	r.Code = lines[r.LineRangeStart:r.LineRangeEnd]
	return r
}


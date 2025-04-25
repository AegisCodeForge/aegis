package controller

import (
	"fmt"
	"strings"
	"net/http"
	
	. "github.com/bctnry/gitus/routes"
	"github.com/bctnry/gitus/templates"
	"github.com/bctnry/gitus/pkg/gitlib"
)


func bindBlobController(ctx RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/blob/{blobId}/", WithLog(func(w http.ResponseWriter, r *http.Request) {
		repoName := r.PathValue("repoName")
		blobId := r.PathValue("blobId")
		
		repo, ok := ctx.GitRepositoryList[repoName]
		if !ok {
			ctx.ReportNotFound(repoName, "Repository", "depot", w, r)
			return
		}
		repoHeaderInfo := templates.RepoHeaderTemplateModel{
			RepoName: repoName,
			RepoDescription: repo.Description,
			TypeStr: "blob",
			NodeName: blobId,
			RepoLabelList: nil,
		}

		gobj, err := repo.ReadObject(blobId)
		if err != nil {
			ctx.ReportObjectReadFailure(blobId, err.Error(), w, r)
			return
		}
		if gobj.Type() != gitlib.BLOB {
			ctx.ReportObjectTypeMismatch(gobj.ObjectId(), "BLOB", gobj.Type().String(), w, r)
			return
		}

		// NOTE THAT we don't know the path with blob so we can't predict what kind of
		// file it is unless we look at its content and hope that we can make a good
		// assumption without calculating too much. the current behaviour is thus
		// intentional and we shall come back to this in the future...
		templateType := "file-text"
		bobj := gobj.(*gitlib.BlobObject)
		if r.URL.Query().Has("raw") {
			w.Write(bobj.Data)
			return
		}
		str := string(bobj.Data)
		permaLink := fmt.Sprintf("/repo/%s/blob/%s", repoName, blobId)

		LogTemplateError(ctx.LoadTemplate(templateType).Execute(w, templates.FileTemplateModel{
			RepoHeaderInfo: repoHeaderInfo,
			File: templates.BlobTextTemplateModel{
				FileLineCount: strings.Count(str, "\n"),
				FileContent: str,
			},
			PermaLink: permaLink,
			TreePath: nil,
			CommitInfo: nil,
			TagInfo: nil,
		}))

	}))
}



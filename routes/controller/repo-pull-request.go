package controller

import (
	"net/http"

	"github.com/bctnry/aegis/routes"
)

func bindRepositoryPullRequestController(ctx *routes.RouterContext) {
	http.HandleFunc("GET /repo/{repoName}/pull-request", routes.WithLog(func(w http.ResponseWriter, r *http.Request){
		
	}))
}

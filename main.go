package main

//go:generate go run devtools/generate-template.go templates

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/bctnry/gitus/pkg/gitlib"
	"github.com/bctnry/gitus/pkg/gitus"
	"github.com/bctnry/gitus/pkg/gitus/model"
	"github.com/bctnry/gitus/routes"
	"github.com/bctnry/gitus/routes/controller"
	"github.com/bctnry/gitus/templates"
)

func getAllNamespace(gitPath string) (map[string]*model.Namespace, error) {
	res := make(map[string]*model.Namespace, 0)
	l, err := os.ReadDir(gitPath)
	if err != nil { return nil, err }
	for _, item := range l {
		namespaceName := item.Name()
		if !model.ValidNamespaceName(namespaceName) { continue }
		p := path.Join(gitPath, namespaceName)
		ns, err := model.NewNamespace(namespaceName, p)
		if err != nil { return nil, err }
		res[namespaceName] = ns
	}
	return res, nil
}

func getAllGitRepository(gitPath string) (map[string]*gitlib.LocalGitRepository, error) {
	res := make(map[string]*gitlib.LocalGitRepository, 0)
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
		res[repoName] = gitlib.NewLocalGitRepository("", repoName, p)
	}
	return res, nil
}

func main() {
	argparse := flag.NewFlagSet("gitus", flag.ContinueOnError)
	argparse.Usage = func() {
		fmt.Fprintf(argparse.Output(), "Usage: gitus [flags] [config]\n")
		argparse.PrintDefaults()
	}
	initFlag := argparse.Bool("init", false, "Create an initial configuration file at the location specified with [config].")
	argparse.Parse(os.Args[1:])
	if len(argparse.Args()) <= 0 {
		argparse.Usage()
		os.Exit(0)
	}
	configPath := argparse.Args()[0]

	if *initFlag {
		err := gitus.CreateConfigFile(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create configuration file: %s\n", err.Error())
			os.Exit(1)
		} else {
			fmt.Printf("Configuration file created. (Please further edit it to fit your exact requirements.)")
			os.Exit(0)
		}
	}

	config, err := gitus.LoadConfigFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration file: %s\n", err.Error())
		os.Exit(1)
	}
	
	masterTemplate := templates.LoadTemplate()
	
	context := routes.RouterContext{
		Config: config,
		MasterTemplate: masterTemplate,
	}
	if !config.UseNamespace {
		grlist, err := getAllGitRepository(config.GitRoot)
		if err != nil {
			log.Panicf("Failed to load git repository: %s\n", err.Error())
		}
		context.GitRepositoryList = grlist
	} else {
		nslist, err := getAllNamespace(config.GitRoot)
		if err != nil {
			log.Panicf("Failed to load git repository: %s\n", err.Error())
		}
		context.GitNamespaceList = nslist
	}

	controller.InitializeRoute(context)

	staticPrefix := config.StaticAssetDirectory
	var fs = http.FileServer(http.Dir(staticPrefix))
	http.Handle("GET /favicon.ico", routes.WithLogHandler(fs))
	http.Handle("GET /static/", http.StripPrefix("/static/", routes.WithLogHandler(fs)))

	log.Println("Serve at :8000")
	http.ListenAndServe(":8000", nil)
}

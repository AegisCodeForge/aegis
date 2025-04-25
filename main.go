package main

//go:generate go run devtools/generate-template.go templates

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/bctnry/gitus/pkg/gitlib"
	"github.com/bctnry/gitus/pkg/gitus"
	"github.com/bctnry/gitus/templates"
	"github.com/bctnry/gitus/routes"
	"github.com/bctnry/gitus/routes/controller"
)

func getAllGitRepository(gitPath string) (map[string]gitlib.LocalGitRepository, error) {
	res := make(map[string]gitlib.LocalGitRepository, 0)
	l, err := os.ReadDir(gitPath)
	if err != nil { return nil, err }
	for _, item := range l {
		p := path.Join(gitPath, item.Name())
		if gitlib.IsValidGitDirectory(p) {
			res[item.Name()] = gitlib.NewLocalGitRepository(p)
			continue
		}
		p2 := path.Join(p, ".git")
		if gitlib.IsValidGitDirectory(p2) {
			res[item.Name() + ".git"] = gitlib.NewLocalGitRepository(p2)
			continue
		}
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
	fmt.Println(argparse.NArg())
	fmt.Println("initFlag", *initFlag, os.Args)
	if len(argparse.Args()) <= 0 {
		argparse.Usage()
		os.Exit(0)
	}
	fmt.Println((*argparse).NArg())
	configPath := argparse.Args()[0]
	fmt.Println("configPath", configPath)

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

	grlist, err := getAllGitRepository(config.GitRoot)
	if err != nil {
		log.Panicf("Failed to load git repository: %s\n", err.Error())
	}

	context := routes.RouterContext{
		Config: config,
		MasterTemplate: masterTemplate,
		GitRepositoryList: grlist,
	}
	controller.InitializeRoute(context)

	staticPrefix := config.StaticAssetDirectory
	var fs = http.FileServer(http.Dir(staticPrefix))
	http.Handle("GET /favicon.ico", routes.WithLogHandler(fs))
	http.Handle("GET /static/", http.StripPrefix("/static/", routes.WithLogHandler(fs)))

	log.Println("Serve at :8000")
	http.ListenAndServe(":8000", nil)
}

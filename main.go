package main

//go:generate go run devtools/generate-template.go templates

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/bctnry/gitus/pkg/gitus"
	"github.com/bctnry/gitus/routes"
	"github.com/bctnry/gitus/routes/controller"
	"github.com/bctnry/gitus/templates"
)

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
	ns, err := config.GetAllNamespace()
	if err != nil {
		log.Panicf("Failed to load git repository: %s\n", err.Error())
	}
	if !config.UseNamespace {
		context.GitRepositoryList = ns[""].RepositoryList
	} else {
		context.GitNamespaceList = ns
	}

	controller.InitializeRoute(context)

	staticPrefix := config.StaticAssetDirectory
	var fs = http.FileServer(http.Dir(staticPrefix))
	http.Handle("GET /favicon.ico", routes.WithLogHandler(fs))
	http.Handle("GET /static/", http.StripPrefix("/static/", routes.WithLogHandler(fs)))

	log.Println("Serve at :8000")
	http.ListenAndServe(":8000", nil)
}

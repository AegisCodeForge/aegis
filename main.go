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
	"github.com/bctnry/gitus/pkg/gitus/db"
	dbinit "github.com/bctnry/gitus/pkg/gitus/db/init"
	ssinit "github.com/bctnry/gitus/pkg/gitus/session/init"
	"github.com/bctnry/gitus/pkg/gitus/ssh"
	"github.com/bctnry/gitus/pkg/passwd"
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
	configArg := argparse.String("config", "", "Speicfy the path to the config fire.")
	argparse.Parse(os.Args[1:])
	
	configPath := *configArg

	if *initFlag {
		err := gitus.CreateConfigFile(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create configuration file: %s\n", err.Error())
			os.Exit(1)
		} else {
			fmt.Printf("Configuration file created. (Please further edit it to fit your exact requirements.)\n")
			os.Exit(0)
		}
	}

	mainCall := argparse.Args()

	config, err := gitus.LoadConfigFile(configPath)
	noConfig := err != nil
	if noConfig && len(mainCall) > 0 && mainCall[0] == "ssh" {
		// assumes that we have a clone/push through ssh and assumes the program to be
		// in the git user's ~/git-shell-commands. go doc said os.Executable may return
		// symlink path if the program is run through symlink, but in this case we
		// don't care since `gitus ssh` is meant to be only run by git shell which
		// means that whatever symlink it is it can only be in ~/git-shell-commands.
		p, err := os.Executable()
		if err != nil {
			fmt.Print(gitlib.ToPktLine(fmt.Sprintf("ERR Failed while trying to figure out last config: %s\n", err.Error())))
			os.Exit(1)
		}
		lastCfgPath := path.Join(path.Dir(path.Dir(p)), "last-config")
		f, err := os.ReadFile(lastCfgPath)
		if err != nil {
			fmt.Print(gitlib.ToPktLine(fmt.Sprintf("ERR Failed while trying to figure out last config: %s\n", err.Error())))
			os.Exit(1)
		}
		configPath = strings.TrimSpace(string(f))
		config, err = gitus.LoadConfigFile(configPath)
		if err != nil {
			fmt.Print(gitlib.ToPktLine(fmt.Sprintf("ERR Failed while trying to figure out last config: %s\n", err.Error())))
			os.Exit(1)
		}
		noConfig = false
	}
	
	if noConfig {
		fmt.Fprintf(os.Stderr, "Failed to load configuration file: %s\n", err.Error())
		os.Exit(1)
	}

	masterTemplate := templates.LoadTemplate()
	
	context := routes.RouterContext{
		Config: config,
		MasterTemplate: masterTemplate,
	}
	
	var dbif db.GitusDatabaseInterface = nil
	
	if !noConfig && !config.PlainMode {
		// check db if plainmode is false.
		dbif, err = dbinit.InitializeDatabase(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load database: %s\n", err.Error())
			os.Exit(1)
		}
		context.DatabaseInterface = dbif

		ssif, err := ssinit.InitializeDatabase(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load database: %s\n", err.Error())
			os.Exit(1)
		}
		context.SessionInterface = ssif

		ok, err := gitusReadyCheck(context)
		if !ok {
			InstallGitus(context)
			os.Exit(1)
		}

		u, err := passwd.GetUser(config.GitUser)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read /etc/passwd while setting up last-config link: %s\n", err.Error())
			fmt.Fprintf(os.Stderr, "You should try to fix the problem and run Gitus again, or else you might not be able to clone/push through SSH.\n")
			os.Exit(1)
		}
		
		lastConfigFilePath := path.Join(u.HomeDir, "last-config")
		err = os.WriteFile(lastConfigFilePath, []byte(configPath), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write last-config link: %s\n", err.Error())
			fmt.Fprintf(os.Stderr, "You should try to fix the problem and run Gitus again, or else you might not be able to clone/push through SSH.\n")
			os.Exit(1)
		}

		keyctx, err := ssh.ToContext(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create key managing context: %s\n", err.Error())
			fmt.Fprintf(os.Stderr, "You should try to fix the problem and run Gitus again, or else you might not be able to clone/push through SSH.\n")
			os.Exit(1)
		}
		
		context.SSHKeyManagingContext = keyctx
		if len(mainCall) > 0 {
			switch mainCall[0] {
			case "install":
				if noConfig {
					fmt.Fprintf(os.Stderr, "No config file specified. Cannot continue.\n")
				} else {
					InstallGitus(context)
				}
				return
			case "reset-admin":
				if noConfig {
					fmt.Fprintf(os.Stderr, "No config file specified. Cannot continue.\n")
				} else {
					ResetAdmin(&context)
				}
				return
			case "ssh":
				if len(mainCall) < 3 {
					fmt.Fprintf(os.Stderr, "Error format for `gitus ssh`.\n")
					return
				}
				HandleSSHLogin(&context, mainCall[1], mainCall[2])
				return
			}
		}
	}

	controller.InitializeRoute(&context)

	staticPrefix := config.StaticAssetDirectory
	var fs = http.FileServer(http.Dir(staticPrefix))
	http.Handle("GET /favicon.ico", routes.WithLogHandler(fs))
	http.Handle("GET /static/", http.StripPrefix("/static/", routes.WithLogHandler(fs)))

	log.Println(fmt.Sprintf("Serve at %s:%d", config.BindAddress, config.BindPort))
	http.ListenAndServe(fmt.Sprintf("%s:%d", config.BindAddress, config.BindPort), nil)
}


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

	"github.com/bctnry/aegis/pkg/gitlib"
	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/db"
	dbinit "github.com/bctnry/aegis/pkg/aegis/db/init"
	"github.com/bctnry/aegis/pkg/aegis/mail"
	rsinit "github.com/bctnry/aegis/pkg/aegis/receipt/init"
	ssinit "github.com/bctnry/aegis/pkg/aegis/session/init"
	"github.com/bctnry/aegis/pkg/aegis/ssh"
	"github.com/bctnry/aegis/pkg/passwd"
	"github.com/bctnry/aegis/routes"
	"github.com/bctnry/aegis/routes/controller"
	"github.com/bctnry/aegis/templates"
)

func main() {
	argparse := flag.NewFlagSet("aegis", flag.ContinueOnError)
	argparse.Usage = func() {
		fmt.Fprintf(argparse.Output(), "Usage: aegis [flags] [config]\n")
		argparse.PrintDefaults()
	}
	initFlag := argparse.Bool("init", false, "Create an initial configuration file at the location specified with [config].")
	configArg := argparse.String("config", "", "Speicfy the path to the config fire.")
	argparse.Parse(os.Args[1:])
	
	configPath := *configArg
	root, err := os.Executable()
	if err != nil {
		fmt.Printf("Failed to resolve absolute path for config file: %s\n", err.Error())
		os.Exit(1)
	}
	if !path.IsAbs(configPath) {
		configPath = path.Join(path.Dir(root), configPath)
	}

	if *initFlag {
		err := aegis.CreateConfigFile(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create configuration file: %s\n", err.Error())
			os.Exit(1)
		} else {
			fmt.Printf("Configuration file created. (Please further edit it to fit your exact requirements.)\n")
			os.Exit(0)
		}
	}

	mainCall := argparse.Args()

	config, err := aegis.LoadConfigFile(configPath)
	noConfig := err != nil
	if noConfig && len(mainCall) > 0 && mainCall[0] == "ssh" {
		// assumes that we have a clone/push through ssh and assumes the program to be
		// in the git user's ~/git-shell-commands. go doc said os.Executable may return
		// symlink path if the program is run through symlink, but in this case we
		// don't care since `aegis ssh` is meant to be only run by git shell which
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
		config, err = aegis.LoadConfigFile(configPath)
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
	
	var dbif db.AegisDatabaseInterface = nil
	
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
			fmt.Fprintf(os.Stderr, "Failed to initialize session store: %s\n", err.Error())
			os.Exit(1)
		}
		context.SessionInterface = ssif

		keyctx, err := ssh.ToContext(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create key managing context: %s\n", err.Error())
			fmt.Fprintf(os.Stderr, "You should try to fix the problem and run Aegis again, or else you might not be able to clone/push through SSH.\n")
			os.Exit(1)
		}
		context.SSHKeyManagingContext = keyctx

		rs, err := rsinit.InitializeReceiptSystem(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create receipt system interface: %s\n", err.Error())
			fmt.Fprintf(os.Stderr, "You should try to fix the problem and run Aegis again, or things like user registration & password resetting wouldn't work properly.\n")
			os.Exit(1)
		}
		context.ReceiptSystem = rs

		ml, err := mail.InitializeMailer(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create mailer interface: %s\n", err.Error())
			fmt.Fprintf(os.Stderr, "You should try to fix the problem and run Aegis again, or things thar depends on sending emails wouldn't work properly.\n")
			os.Exit(1)
		}
		context.Mailer = ml

		ok, err := aegisReadyCheck(context)
		if !ok {
			fmt.Fprintf(os.Stderr, "Aegis Ready Check failed: %s\n", err.Error())
			InstallAegis(context)
			os.Exit(1)
		}

		u, err := passwd.GetUser(config.GitUser)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read /etc/passwd while setting up last-config link: %s\n", err.Error())
			fmt.Fprintf(os.Stderr, "You should try to fix the problem and run Aegis again, or else you might not be able to clone/push through SSH.\n")
			os.Exit(1)
		}
		
		lastConfigFilePath := path.Join(u.HomeDir, "last-config")
		err = os.WriteFile(lastConfigFilePath, []byte(configPath), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write last-config link: %s\n", err.Error())
			fmt.Fprintf(os.Stderr, "You should try to fix the problem and run Aegis again, or else you might not be able to clone/push through SSH.\n")
			os.Exit(1)
		}

		if len(mainCall) > 0 {
			switch mainCall[0] {
			case "install":
				if noConfig {
					fmt.Fprintf(os.Stderr, "No config file specified. Cannot continue.\n")
				} else {
					fmt.Println(mainCall)
					InstallAegis(context)
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
					fmt.Print(gitlib.ToPktLine("Error format for `aegis ssh`."))
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


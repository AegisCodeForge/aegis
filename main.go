package main

//go:generate go run devtools/generate-template.go templates

import (
	gocontext "context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/confirm_code"
	dbinit "github.com/bctnry/aegis/pkg/aegis/db/init"
	"github.com/bctnry/aegis/pkg/aegis/mail"
	rsinit "github.com/bctnry/aegis/pkg/aegis/receipt/init"
	ssinit "github.com/bctnry/aegis/pkg/aegis/session/init"
	"github.com/bctnry/aegis/pkg/aegis/ssh"
	"github.com/bctnry/aegis/pkg/gitlib"
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

	// attempt to resolve config file path.
	// if the provided path is relative, resolve it against os.Executable.
	configPath := *configArg
	root, err := os.Executable()
	if err != nil {
		fmt.Printf("Failed to resolve absolute path for config file: %s\n", err.Error())
		os.Exit(1)
	}
	if !path.IsAbs(configPath) {
		configPath = path.Join(path.Dir(root), configPath)
	}

	// check if init. if init, we start web installer or generate
	// config. if we *don't* use the web installer, we don't perform
	// installation of databases because we don't know what kind of
	// config the user want; this is different from web installer
	// because we ask the user to provide required info during the
	// process.
	if *initFlag {
		if askYesNo("Start web installer?") {
			WebInstaller()
			os.Exit(0)
		}
		err := aegis.CreateConfigFile(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create configuration file: %s\n", err.Error())
			os.Exit(1)
		}
		fmt.Printf("Configuration file created. (Please further edit it to fit your exact requirements.)\n")
		os.Exit(0)
	}

	mainCall := argparse.Args()

	config, err := aegis.LoadConfigFile(configPath)
	noConfig := err != nil
	// we use the same executable for the web server and the ssh handling command,
	// this is necessary to separate the two situations.
	// when aegis executable is called through ssh it's a completely different
	// situation where we would absolutely not have the config file path. a
	// `last-config` file is used as a hack to provide this info. see
	// `docs/ssh.org` for more info.
	if noConfig && len(mainCall) > 0 && (mainCall[0] == "ssh" || mainCall[0] == "web-hooks") {
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
		// we attempt to resolve config file location. this also means
		// that on the same server one git user can only run one aegis
		// instance on the same server.
		// we first expect the executable to be at $HOME/git-shell-commands,
		// which means we should reach $HOME by calling path.Dir twice.
		lastCfgPath := path.Join(path.Dir(path.Dir(p)), "last-config")
		f, err := os.ReadFile(lastCfgPath)
		// if it's not found at that path, we try our best to find it
		// from other places...
		if err != nil {
			subj := p
			for err != nil {
				d := path.Dir(subj)
				tp := path.Join(d, "last-config")
				f, err = os.ReadFile(tp)
				if err == nil { break }
				if path.Dir(d) == d {
					// we've reached to the root of the fs. if we can't
					// find it there, we have no chance of finding it.
					fmt.Print(gitlib.ToPktLine(fmt.Sprintf("ERR Failed while trying to figure out last config: %s\n", err.Error())))
					os.Exit(1)
				}
				subj = d
			}
		}
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

	// if we still failed to resolve config file we refuse to go further,
	// since whatever comes next would require info only provided by config.
	if noConfig {
		fmt.Fprintf(os.Stderr, "Failed to load configuration file: %s\n", err.Error())
		os.Exit(1)
	}
	
	masterTemplate := templates.LoadTemplate()
	context := routes.RouterContext{
		Config: config,
		MasterTemplate: masterTemplate,
	}

	// if it's not plain mode we need to setup database.
	if !config.PlainMode {
		dbif, err := dbinit.InitializeDatabase(config)
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
		}
		context.Mailer = ml

		ccm, err := confirm_code.InitializeConfirmCodeManager(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create confirm code manager: %s\n", err.Error())
			fmt.Fprintf(os.Stderr, "You should try to fix the problem and run Aegis again, or things thar depends on sending emails wouldn't work properly.\n")
		}
		context.ConfirmCodeManager = ccm

		ok, err := aegisReadyCheck(context)
		if !ok {
			fmt.Fprintf(os.Stderr, "Aegis Ready Check failed: %s\n", err.Error())
			InstallAegis(context)
			os.Exit(1)
		}

		u, err := user.Lookup(config.GitUser)
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
		gitUser, err := user.Lookup(context.Config.GitUser)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to find Git user %s: %s\n", context.Config.GitUser, err.Error())
			fmt.Fprintf(os.Stderr, "You should try to fix the problem and run Aegis again, or else you might not be able to clone/push through SSH.\n")
			os.Exit(1)
		}
		// peak of stdlib design, golang.
		uid, _ := strconv.Atoi(gitUser.Uid)
		gid, _ := strconv.Atoi(gitUser.Gid)
		err = os.Chown(lastConfigFilePath, uid, gid)

		// the features of these commands are meaningless in the use case of
		// plain mode, so the dispatching is done within this if branch.
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
			case "web-hooks":
				if len(mainCall) < 7 {
					fmt.Print(gitlib.ToPktLine("Error format for `aegis web-hooks`."))
					return
				}
				switch mainCall[1] {
				case "send":
					HandleWebHook(&context, mainCall[2], mainCall[3], mainCall[4], mainCall[5], mainCall[6])
				default:
					fmt.Print(gitlib.ToPktLine(fmt.Sprintf("Error command for `aegis web-hooks`: %s.", mainCall[1])))
				}
				return
			}
		}
	}

	staticPrefix := config.StaticAssetDirectory
	templates.UnpackStaticFileTo(staticPrefix)
	var fs = http.FileServer(http.Dir(staticPrefix))
	http.Handle("GET /favicon.ico", routes.WithLogHandler(fs))
	http.Handle("GET /static/", http.StripPrefix("/static/", routes.WithLogHandler(fs)))
	server := &http.Server{
		Addr: fmt.Sprintf("%s:%d", config.BindAddress, config.BindPort),
	}

	context.RateLimiter = routes.NewRateLimiter(config)
	
	controller.InitializeRoute(&context)

	go func() {
		log.Printf("Trying to serve at %s:%d\n", config.BindAddress, config.BindPort)
		err := server.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
		log.Println("Stopped serving new connections.")
	}()

	// apparently go kills absolutely everything when main returns -
	// all the goroutines and things would be just gone and not even
	// deferred calls are executed, which is insane if you think about
	// it. the `http.Server.Shutdown` method closes the http server
	// gracefully but `http.ListenAndServe` just serves and does not
	// return the server obj, a separate Server obj is needed to
	// close. putting the teardown part after `http.ListenAndServe`
	// doesn't seem to cut it because of SIGINT and the like. we wait
	// on a channel (which we set up beforehand to put up a notifying
	// message when SIGINT/others occur) so that in cases like those
	// we would still have a chance to wrap things up.
	// this is also used for the webinstaller since it's also a http
	// server as well.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	shutdownCtx, shutdownRelease := gocontext.WithTimeout(gocontext.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP shutdown err: %v", err.Error())
	}

	if context.DatabaseInterface != nil {
		if err = context.DatabaseInterface.Dispose(); err != nil {
			log.Printf("Failed to dispose database interface: %s\n", err.Error())
		}
	}
	if context.SessionInterface != nil {
		if err = context.SessionInterface.Dispose(); err != nil {
			log.Printf("Failed to dispose session store: %s\n", err.Error())
		}
	}
	if context.ReceiptSystem != nil {
		if err = context.ReceiptSystem.Dispose(); err != nil {
			log.Printf("Failed to dispose receipt system: %s\n", err.Error())
		}
	}
	
	log.Println("Graceful shutdown complete.")
}


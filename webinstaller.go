package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bctnry/aegis/pkg/aegis"
	"github.com/bctnry/aegis/pkg/aegis/db"
	dbinit "github.com/bctnry/aegis/pkg/aegis/db/init"
	"github.com/bctnry/aegis/pkg/aegis/model"
	rsinit "github.com/bctnry/aegis/pkg/aegis/receipt/init"
	ssinit "github.com/bctnry/aegis/pkg/aegis/session/init"
	"github.com/bctnry/aegis/templates"
	"golang.org/x/crypto/bcrypt"
)

type WebInstallerRoutingContext struct {
	Template *template.Template
	// yes, we do share the same object between multiple goroutie,
	// but i don't think this would be a problem for a simple web
	// installer.
	// step 1 - plain mode or non-plain mode?
	//          use namespace or not?
	//          plain mode - goto step [6]
	// step 2 - database config
	// step 3 - session config
	// step 4 - mailer config
	// step 5 - receipt system config
	// step 6 - git root & git user
	// step 7 - ignored namespaces & repositories
	// step 8 - web front setup:
	//          depot name
	//          front page config
	//          (static assets dir default to be $HOME/aegis-static/)
	//          bind address & port
	//          http host name
	//          ssh host name (disabled if plain mode)
	//          allow registration
	//          email confirmation required
	//          manual approval
	// plain mode on: 1-6-7-8
	// plain mode off: 1-2-3-4-5-6-8
	Step int
	Config *aegis.AegisConfig
	ConfirmStageReached bool
	ResultingFilePath string
}

func logTemplateError(e error) {
	if e != nil { log.Print(e) }
}

func (ctx *WebInstallerRoutingContext) loadTemplate(name string) *template.Template {
	return ctx.Template.Lookup(name)
}

func withLog(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf(" %s %s\n", r.Method, r.URL.Path)
		f(w, r)
	}
}

func foundAt(w http.ResponseWriter, p string) {
	w.Header().Add("Content-Length", "0")
	w.Header().Add("Location", p)
	w.WriteHeader(302)
}

func (ctx *WebInstallerRoutingContext) reportRedirect(target string, timeout int, title string, message string, w http.ResponseWriter) {
	logTemplateError(ctx.loadTemplate("webinstaller/_redirect").Execute(w, templates.WebInstRedirectWithMessageModel{
		Timeout: timeout,
		RedirectUrl: target,
		MessageTitle: title,
		MessageText: message,
	}))
}

func bindAllWebInstallerRoutes(ctx *WebInstallerRoutingContext) {
	http.HandleFunc("GET /", withLog(func(w http.ResponseWriter, r *http.Request) {
		logTemplateError(ctx.loadTemplate("webinstaller/start").Execute(w, &templates.WebInstallerTemplateModel{
			Config: ctx.Config,
			ConfirmStageReached: ctx.ConfirmStageReached,
		}))
	}))
	
	http.HandleFunc("GET /step1", withLog(func(w http.ResponseWriter, r *http.Request) {
		logTemplateError(ctx.loadTemplate("webinstaller/step1").Execute(w, &templates.WebInstallerTemplateModel{
			Config: ctx.Config,
			ConfirmStageReached: ctx.ConfirmStageReached,
		}))
	}))
	http.HandleFunc("POST /step1", withLog(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			ctx.reportRedirect("/step1", 0, "Invalid Request", "The request is of an invalid form. Please try again.", w)
			return
		}
		ctx.Config.PlainMode = len(r.Form.Get("plain-mode")) > 0
		ctx.Config.UseNamespace = len(r.Form.Get("enable-namespace")) > 0
		if ctx.Config.PlainMode {
			foundAt(w, "/step6")
		} else {
			foundAt(w, "/step2")
		}
	}))
	
	http.HandleFunc("GET /step2", withLog(func(w http.ResponseWriter, r *http.Request) {
		logTemplateError(ctx.loadTemplate("webinstaller/step2").Execute(w, &templates.WebInstallerTemplateModel{
			Config: ctx.Config,
			ConfirmStageReached: ctx.ConfirmStageReached,
		}))
	}))
	http.HandleFunc("POST /step2", withLog(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			ctx.reportRedirect("/step2", 0, "Invalid Request", "The request is of an invalid form. Please try again.", w)
			return
		}
		ctx.Config.Database = aegis.AegisDatabaseConfig{
			Type: strings.TrimSpace(r.Form.Get("database-type")),
			Path: strings.TrimSpace(r.Form.Get("database-path")),
			URL: strings.TrimSpace(r.Form.Get("database-url")),
			UserName: strings.TrimSpace(r.Form.Get("database-username")),
			DatabaseName: strings.TrimSpace(r.Form.Get("database-database-name")),
			TablePrefix: strings.TrimSpace(r.Form.Get("database-table-prefix")),
			Password: strings.TrimSpace(r.Form.Get("database-password")),
		}

 		foundAt(w, "/step3")
	}))
	
	http.HandleFunc("GET /step3", withLog(func(w http.ResponseWriter, r *http.Request) {
		logTemplateError(ctx.loadTemplate("webinstaller/step3").Execute(w, &templates.WebInstallerTemplateModel{
			Config: ctx.Config,
			ConfirmStageReached: ctx.ConfirmStageReached,
		}))
	}))
	http.HandleFunc("POST /step3", withLog(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			ctx.reportRedirect("/step3", 0, "Invalid Request", "The request is of an invalid form. Please try again.", w)
			return
		}
		i, err := strconv.ParseInt(strings.TrimSpace(r.Form.Get("session-database-number")), 10, 32)
		if err != nil {
			ctx.reportRedirect("/step3", 0, "Invalid Request", "The request is of an invalid form. Please try again.", w)
			return
		}
		ctx.Config.Session = aegis.AegisSessionConfig{
			Type: strings.TrimSpace(r.Form.Get("session-type")),
			Path: strings.TrimSpace(r.Form.Get("session-path")),
			TablePrefix: strings.TrimSpace(r.Form.Get("session-table-prefix")),
			Host: strings.TrimSpace(r.Form.Get("session-host")),
			UserName: strings.TrimSpace(r.Form.Get("session-username")),
			Password: strings.TrimSpace(r.Form.Get("session-password")),
			DatabaseNumber: int(i),
		}
		foundAt(w, "/step4")
	}))

	
	http.HandleFunc("GET /step4", withLog(func(w http.ResponseWriter, r *http.Request) {
		logTemplateError(ctx.loadTemplate("webinstaller/step4").Execute(w, &templates.WebInstallerTemplateModel{
			Config: ctx.Config,
			ConfirmStageReached: ctx.ConfirmStageReached,
		}))
	}))
	http.HandleFunc("POST /step4", withLog(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			ctx.reportRedirect("/step4", 0, "Invalid Request", "The request is of an invalid form. Please try again.", w)
			return
		}
		i, err := strconv.ParseInt(strings.TrimSpace(r.Form.Get("mailer-smtp-port")), 10, 32)
		if err != nil {
			ctx.reportRedirect("/step4", 0, "Invalid Request", "The request is of an invalid form. Please try again.", w)
			return
		}
		ctx.Config.Mailer = aegis.AegisMailerConfig{
			Type: strings.TrimSpace(r.Form.Get("mailer-type")),
			SMTPServer: strings.TrimSpace(r.Form.Get("mailer-smtp-server")),
			SMTPPort: int(i),
			SMTPAuth: strings.TrimSpace(r.Form.Get("mailer-smtp-auth")),
			User: strings.TrimSpace(r.Form.Get("mailer-user")),
			Password: strings.TrimSpace(r.Form.Get("mailer-password")),
		}
		foundAt(w, "/step5")
	}))
	
	http.HandleFunc("GET /step5", withLog(func(w http.ResponseWriter, r *http.Request) {
		logTemplateError(ctx.loadTemplate("webinstaller/step5").Execute(w, &templates.WebInstallerTemplateModel{
			Config: ctx.Config,
			ConfirmStageReached: ctx.ConfirmStageReached,
		}))
	}))
	http.HandleFunc("POST /step5", withLog(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			ctx.reportRedirect("/step5", 0, "Invalid Request", "The request is of an invalid form. Please try again.", w)
			return
		}
		ctx.Config.ReceiptSystem = aegis.AegisReceiptSystemConfig{
			Type: strings.TrimSpace(r.Form.Get("receipt-system-type")),
			Path: strings.TrimSpace(r.Form.Get("receipt-system-path")),
			URL: strings.TrimSpace(r.Form.Get("receipt-system-url")),
			UserName: strings.TrimSpace(r.Form.Get("receipt-system-username")),
			DatabaseName: strings.TrimSpace(r.Form.Get("receipt-system-database-name")),
			Password: strings.TrimSpace(r.Form.Get("receipt-system-password")),
			TablePrefix: strings.TrimSpace(r.Form.Get("receipt-system-table-prefix")),
		}
		foundAt(w, "/step6")
	}))
	
	http.HandleFunc("GET /step6", withLog(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(ctx.Config.GitUser, ctx.Config.GitRoot)
		logTemplateError(ctx.loadTemplate("webinstaller/step6").Execute(w, &templates.WebInstallerTemplateModel{
			Config: ctx.Config,
			ConfirmStageReached: ctx.ConfirmStageReached,
		}))
	}))
	http.HandleFunc("POST /step6", withLog(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			ctx.reportRedirect("/step6", 0, "Invalid Request", "The request is of an invalid form. Please try again.", w)
			return
		}
		ctx.Config.GitRoot = strings.TrimSpace(r.Form.Get("git-root"))
		ctx.Config.GitUser = strings.TrimSpace(r.Form.Get("git-user"))
		next := ""
		if ctx.Config.PlainMode {
			next = "/step7"
		} else {
			next = "/step8"
		}
		err = templates.UnpackStaticFileTo(ctx.Config.StaticAssetDirectory)
		if err != nil {
			ctx.reportRedirect(next, 0, "Failed", fmt.Sprintf("Static file unpack is unsuccessful due to reason: %s. You can still move forward but would have to unpack static file yourself.", err.Error()), w)
			return
		}
		foundAt(w, next)
	}))
	
	http.HandleFunc("GET /step7", withLog(func(w http.ResponseWriter, r *http.Request) {
		logTemplateError(ctx.loadTemplate("webinstaller/step7").Execute(w, &templates.WebInstallerTemplateModel{
			Config: ctx.Config,
			ConfirmStageReached: ctx.ConfirmStageReached,
		}))
	}))
	http.HandleFunc("POST /step7", withLog(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			ctx.reportRedirect("/step1", 0, "Invalid Request", "The request is of an invalid form. Please try again.", w)
			return
		}
		ctx.Config.IgnoreNamespace = make([]string, 0)
		for k := range strings.SplitSeq(r.Form.Get("ignore-namespace"), ",") {
			ctx.Config.IgnoreNamespace = append(ctx.Config.IgnoreNamespace, k)
		}
		ctx.Config.IgnoreRepository = make([]string, 0)
		for k := range strings.SplitSeq(r.Form.Get("ignore-repository"), ",") {
			ctx.Config.IgnoreRepository = append(ctx.Config.IgnoreRepository, k)
		}
		foundAt(w, "/step8")
	}))
	
	http.HandleFunc("GET /step8", withLog(func(w http.ResponseWriter, r *http.Request) {
		logTemplateError(ctx.loadTemplate("webinstaller/step8").Execute(w, &templates.WebInstallerTemplateModel{
			Config: ctx.Config,
			ConfirmStageReached: ctx.ConfirmStageReached,
		}))
	}))
	http.HandleFunc("POST /step8", withLog(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			ctx.reportRedirect("/step1", 0, "Invalid Request", "The request is of an invalid form. Please try again.", w)
			return
		}
		ctx.Config.DepotName = strings.TrimSpace(r.Form.Get("depot-name"))
		ctx.Config.BindAddress = strings.TrimSpace(r.Form.Get("bind-address"))
		i, err := strconv.ParseInt(strings.TrimSpace(r.Form.Get("bind-port")), 10, 32)
		if err != nil {
			ctx.reportRedirect("/step8", 0, "Invalid Request", "The request is of an invalid form. Please try again.", w)
			return
		}
		ctx.Config.BindPort = int(i)
		ctx.Config.HttpHostName = strings.TrimSpace(r.Form.Get("http-host-name"))
		ctx.Config.SshHostName = strings.TrimSpace(r.Form.Get("ssh-host-name"))
		frontPageType := strings.TrimSpace(r.Form.Get("front-page-type"))
		switch frontPageType {
		case "all/namespace": fallthrough
		case "all/repository":
			ctx.Config.FrontPageType = frontPageType
		case "static/html": fallthrough
		case "static/text": fallthrough
		case "static/markdown": fallthrough
		case "static/org":
			ctx.Config.FrontPageType = frontPageType
			ctx.Config.FrontPageContent = r.Form.Get("front-page-text")
		case "repository":
			v := r.Form.Get("front-page-value")
			ctx.Config.FrontPageType = "repository/" + v
		case "namespace":
			v := r.Form.Get("front-page-value")
			ctx.Config.FrontPageType = "namespace/" + v
		}
		ctx.Config.AllowRegistration = len(strings.TrimSpace(r.Form.Get("allow-registration"))) > 0
		ctx.Config.EmailConfirmationRequired = len(strings.TrimSpace(r.Form.Get("email-confirmation-required"))) > 0
		ctx.Config.ManualApproval = len(strings.TrimSpace(r.Form.Get("manual-approval"))) > 0
		foundAt(w, "/confirm")
	}))
	
	http.HandleFunc("GET /confirm", withLog(func(w http.ResponseWriter, r *http.Request) {
		ctx.ConfirmStageReached = true
		logTemplateError(ctx.loadTemplate("webinstaller/confirm").Execute(w, &templates.WebInstallerTemplateModel{
			Config: ctx.Config,
			ConfirmStageReached: ctx.ConfirmStageReached,
		}))
	}))

	http.HandleFunc("GET /install", withLog(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8" />
    <title>Aegis Web Installer</title>`)
		ctx.loadTemplate("webinstaller/_style").Execute(w, nil)
		fmt.Fprint(w, `
  </head>
  <body>
    <header>
	  <h1><a href="/">Aegis Web Installer</a></h1>
	  <ul>
        <li><a href="/step1">Step 1: Plain Mode; Use Namespace</a></li>
        <li><a href="/step2">Step 2: Database Config</a></li>
        <li><a href="/step3">Step 3: Session Config</a></li>
        <li><a href="/step4">Step 4: Mailer Config</a></li>
        <li><a href="/step5">Step 5: Receipt System Config</a></li>
        <li><a href="/step6">Step 6: Git Root &amp; Git User</a></li>
        <li><a href="/step7">Step 7: Ignored Namespace/Repositories</a></li>
        <li><a href="/step8">Step 8: Misc. Setup</a></li>
        <li><a href="/confirm">Confirm</a></li>
      </ul>
	</header>

	<hr />
`)
		if len(strings.TrimSpace(ctx.Config.GitUser)) <= 0 {
			fmt.Fprint(w, "<p>Git user empty. Please fix this...</p>")
			goto leave
		}
		if !func()bool{
			_, err := user.Lookup(ctx.Config.GitUser)
			if err == nil { return true }
			fmt.Fprint(w, "<p>Creating Git user...</p>")
			gitShellPath, err := whereIs("git-shell")
			if err != nil {
				fmt.Fprintf(w, "<p>Failed to search for git-shell: %s</p>", err.Error())
				return false
			}
			if len(gitShellPath) <= 0 {
				fmt.Fprint(w, "<p>Failed to search for git-shell: git-shell path empty.</p>")
				return false
			}
			homePath := fmt.Sprintf("/home/%s", ctx.Config.GitUser)
			ctx.Config.StaticAssetDirectory = path.Join(homePath, "aegis-static-assets")
			err = os.MkdirAll(homePath, os.ModeDir|0755)
			if err != nil {
				fmt.Fprintf(w, "<p>Failed to create home directory for user %s: %s</p>", ctx.Config.GitUser, homePath)
				return false
			}
			useraddPath, err := whereIs("useradd")
			if err != nil {
				fmt.Fprintf(w, "<p>Failed to find command \"useradd\": %s</p>", err.Error())
				return false
			}
			if len(useraddPath) <= 0 {
				fmt.Fprint(w, "<p>Failed to find command \"useradd\": useradd path empty")
				return false
			}
			cmd := exec.Command(useraddPath, "-d", homePath, "-m", "-s", gitShellPath, ctx.Config.GitUser)
			err = cmd.Run()
			if err != nil {
				fmt.Fprintf(w, "<p>Failed to run useradd: %s</p>", err.Error())
				return false
			}
			return true
		}() { goto leave }

		if !func()bool{
			gitUser, err := user.Lookup(ctx.Config.GitUser)
			if err != nil {
				fmt.Fprintf(w, "<p>Somehow failed to retrieve user after registering: %s\n", err.Error())
				return false
			}
			homePath := gitUser.HomeDir
			uid, _ := strconv.Atoi(gitUser.Uid)
			gid, _ := strconv.Atoi(gitUser.Gid)
			fmt.Fprint(w,"<p>Chown-ing git user home directory...</p>")
			err = os.Chown(homePath, int(uid), int(gid))
			if err != nil {
				fmt.Fprintf(w, "<p>Failed to chown the git user home directory: %s</p>", err.Error())
				return false
			}
			fmt.Fprint(w, "<p>Creating git-shell-commands directory...</p>")
			gitShellCommandPath := path.Join(homePath, "git-shell-commands")
			err = createOtherOwnedDirectory(gitShellCommandPath, gitUser.Uid, gitUser.Gid)
			if err != nil {
				fmt.Fprintf(w, "<p>Failed to chown the git shell command directory: %s</p>", err.Error())
				return false
			}
			fmt.Fprint(w, "<p>Creating .ssh directory...</p>")
			sshPath := path.Join(homePath, ".ssh")
			err = createOtherOwnedDirectory(sshPath, gitUser.Uid, gitUser.Gid)
			if err != nil {
				fmt.Fprintf(w, "<p>Failed to create the .ssh folder: %s</p>", err.Error())
				return false
			}
			fmt.Fprint(w, "<p>Creating authorized_keys file...</p>")
			authorizedKeysPath := path.Join(homePath, ".ssh", "authorized_keys")
			err = createOtherOwnedFile(authorizedKeysPath, gitUser.Uid, gitUser.Gid)
			if err != nil {
				fmt.Fprintf(w, "<p>Failed to create the authorized_keys file: %s</p>", err.Error())
				return false
			}
			fmt.Fprint(w, "<p>Copying aegis executable...</p>")
			s, err := os.Executable()
			if err != nil {
				fmt.Fprintf(w, "<p>Failed to copy Aegis executable: %s</p>", err.Error())
				return false
			}
			f, err := os.Open(s)
			if err != nil {
				fmt.Fprintf(w, "<p>Failed to copy Aegis executable: %s</p>", err.Error())
				return false
			}
			defer f.Close()
			aegisPath := path.Join(homePath, "git-shell-commands", "aegis")
			fout, err := os.OpenFile(aegisPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0754)
			if err != nil {
				fmt.Fprintf(w, "<p>Failed to copy Aegis executable: %s\n</p>", err.Error())
				return false
			}
			defer fout.Close()
			_, err = io.Copy(fout, f)
			if err != nil {
				fmt.Fprintf(w, "<p>Failed to copy Aegis executable: %s\n</p>", err.Error())
				return false
			}
			err = os.Chown(aegisPath, uid, gid)
			if err != nil {
				fmt.Fprintf(w, "<p>Failed to copy Aegis executable: %s\n</p>", err.Error())
				return false
			}
			err = os.MkdirAll(ctx.Config.GitRoot, os.ModeDir|0755)
			if errors.Is(err, os.ErrExist) {
				err = os.Chown(ctx.Config.GitRoot, uid, gid)
				if err != nil {
					fmt.Fprintf(w, "<p>Failed to chown git root: %s\n</p>", err.Error())
					return false
				}
			}
			if err != nil {
				fmt.Fprintf(w, "<p>Failed to chown git root: %s\n</p>", err.Error())
				return false
			}
			ctx.Config.FilePath = path.Join(homePath, fmt.Sprintf("aegis-config-%d.json", time.Now().Unix()))
			fmt.Fprint(w, "<p>Git user setup done.</p>")
			ctx.Config.RecalculateProperPath()
			err = ctx.Config.Sync()
			if err != nil {
				fmt.Fprintf(w, "<p>Failed to save config file: %s\n. You might need to do this again or even manually.</p>", err.Error())
				return false
			}
			return true
		}() { goto leave }

		if !func()bool{
			fmt.Fprint(w, "<p>Initializing database...</p>")
			dbif, err := dbinit.InitializeDatabase(ctx.Config)
			if err != nil {
				fmt.Fprintf(w, "<p>Failed to initialize database: %s</p>", err.Error())
				return false
			}
			defer dbif.Dispose()
			chkres, err := dbif.IsDatabaseUsable()
			if err != nil {
				fmt.Fprintf(w, "<p>Failed to initialize database: %s</p>", err.Error())
				return false
			}
			if !chkres {
				err = dbif.InstallTables()
				if err != nil {
					fmt.Fprintf(w, "<p>Failed to initialize database: %s</p>", err.Error())
					return false
				}
			}
			
			fmt.Fprint(w, "<p>Initialization done.</p>")
			return true
		}() { goto leave }
		
		if !func()bool{
			fmt.Fprint(w, "<p>Initializing session store...</p>")
			ssif, err := ssinit.InitializeDatabase(ctx.Config)
			if err != nil {
				fmt.Fprintf(w, "<p>Failed to initialize session store: %s</p>", err.Error())
				return false
			}
			defer ssif.Dispose()
			chkres, err := ssif.IsSessionStoreUsable()
			if err != nil {
				fmt.Fprintf(w, "<p>Failed to initialize session store: %s</p>", err.Error())
				return false
			}
			if !chkres {
				err = ssif.Install()
				if err != nil {
					fmt.Fprintf(w, "<p>Failed to initialize session store: %s</p>", err.Error())
					return false
				}
			}
			fmt.Fprint(w, "<p>Initialization done.</p>")
			return true
		}() { goto leave }
		
		if !func()bool{
			w.Write([]byte("<p>Initializing receipt system...</p>"))
			rsif, err := rsinit.InitializeReceiptSystem(ctx.Config)
			if err != nil {
				fmt.Fprintf(w, "<p>Failed to initialize receipt system: %s</p>", err.Error())
				return false
			}
			defer rsif.Dispose()
			chkres, err := rsif.IsReceiptSystemUsable()
			if err != nil {
				fmt.Fprintf(w, "<p>Failed to initialize receipt system: %s</p>", err.Error())
				return false
			}
			if !chkres {
				err = rsif.Install()
				if err != nil {
					fmt.Fprintf(w, "<p>Failed to initialize receipt system: %s</p>", err.Error())
					return false
				}
			}
			fmt.Fprint(w, "<p>Initialization done.</p>")
			return true
		}() { goto leave }
		
		if !func()bool{
			fmt.Fprint(w, "<p>Setting up admin user.</p>")
			dbif, err := dbinit.InitializeDatabase(ctx.Config)
			if err != nil {
				fmt.Fprintf(w, "<p>Failed to open database while setting up admin user: %s</p>", err.Error())
				return false
			}
			defer dbif.Dispose()
			adminExists := false
			_, err = dbif.GetUserByName("admin")
			if err == db.ErrEntityNotFound {
				adminExists = false
			} else if err != nil {
				fmt.Fprintf(w, "<p>Failed to check database while setting up admin user: %s</p>", err.Error())
				return false
			} else {
				adminExists = true
			}
			if adminExists {
				err = dbif.HardDeleteUserByName("admin")
				if err != nil {
					fmt.Fprintf(w, "<p>Failed to remove original admin user while setting up new admin user: %s</p>", err.Error())
					return false
				}
			}
			userPassword := mkpass()
			r, err := bcrypt.GenerateFromPassword([]byte(userPassword), bcrypt.DefaultCost)
			if err != nil {
				fmt.Fprintf(w, "<p>Failed to generate password: %s</p>", err.Error())
				return false
			}
			_, err = dbif.RegisterUser("admin", "", string(r), model.SUPER_ADMIN)
			if err != nil {
				fmt.Fprintf(w, "<p>Failed to register user: %s</p>", err.Error())
				return false
			}
			fmt.Fprintf(w, "<p>Admin user set up properly.</p><pre>Username: admin\nPassword: %s</pre><p>Please copy the password above because we don't store the plaintext; but, in the case you forgot, you can always run the following command to reset the admin user's password:</p><pre>aegis -config %s reset-admin</pre>", userPassword, ctx.Config.FilePath)
			return true
		}() { goto leave }

		if !func()bool{
			gitUser, _ := user.Lookup(ctx.Config.GitUser)
			var uid int
			var gid int
			if gitUser != nil {
				uid, _ = strconv.Atoi(gitUser.Uid)
				gid, _ = strconv.Atoi(gitUser.Gid)
			}
			fmt.Println("fp", ctx.Config.FilePath)
			if ctx.Config.Database.Type == "sqlite" {
				if gitUser == nil {
					fmt.Fprint(w, "<p class=\"warning\">Failed to fild Git user's uid & gid when chowning sqlite database. You need to perform this action on your own after this installation process...")
				} else {
					err := os.Chown(ctx.Config.ProperDatabasePath(), uid, gid)
					fmt.Println("prop", ctx.Config.ProperDatabasePath())
					if err != nil {
						fmt.Fprintf(w, "<p class=\"warning\">Failed to chown sqlite database: %s. You need to perform this action on your own after this installation process...", err.Error())
					}
				}
			}
			if ctx.Config.Session.Type == "sqlite" {
				if gitUser == nil {
					fmt.Fprintf(w, "<p class=\"warning\">Failed to fild Git user's uid & gid when chowning sqlite database. You need to perform this action on your own after this installation process...")
				} else {
					err := os.Chown(ctx.Config.ProperSessionPath(), uid, gid)
					if err != nil {
						fmt.Fprintf(w, "<p class=\"warning\">Failed to chown sqlite database: %s. You need to perform this action on your own after this installation process...", err.Error())
					}
				}
			}
			if ctx.Config.ReceiptSystem.Type == "sqlite" {
				if gitUser == nil {
					fmt.Fprintf(w, "<p class=\"warning\">Failed to fild Git user's uid & gid when chowning sqlite database. You need to perform this action on your own after this installation process...")
				} else {
					err := os.Chown(ctx.Config.ProperReceiptSystemPath(), uid, gid)
					if err != nil {
						fmt.Fprintf(w, "<p class=\"warning\">Failed to chown sqlite database: %s. You need to perform this action on your own after this installation process...", err.Error())
					}
				}
			}
			return true
		}() { goto leave }

		
		fmt.Fprint(w, "<p>Done! <a href=\"./finish\">Go to the next step.</a></p>")
		goto footer

	leave:
		fmt.Fprintf(w, "<p>The installation process failed but the config file might've been saved successfully at <code>%s</code>. In this case, you need to run the following command:</p><pre>aegis -config %s</pre></p>", ctx.Config.FilePath, ctx.Config.FilePath)

	footer:
		fmt.Fprint(w, `
    <hr />
    <footer>
      <div class="footer-message">
        Powered by <a href="https://github.com/bctnry/aegis">Aegis</a>.
      </div>
    </footer>
  </body>
</html>`)
	}))
	
	http.HandleFunc("GET /finish", withLog(func(w http.ResponseWriter, r *http.Request) {
		
		logTemplateError(ctx.loadTemplate("webinstaller/finish").Execute(w, &templates.WebInstallerTemplateModel{
			Config: ctx.Config,
		}))
	}))
}

func WebInstaller() {
	fmt.Println("This is the Aegis web installer. We will start a web server, which allows us to provide you a more user-friendly interface for configuring your Aegis instance. This web server will be shut down when the installation is finished. You can always start the web installer by using the `-init` flag or the `install` command.")
	var portNum int = 0
	for {
		r, err := askString("Please enter the port number this web server would bind to.", "8001")
		if err != nil {
			fmt.Printf("Failed to get a response: %s\n", err.Error())
			os.Exit(1)
		}
		portNum, err = strconv.Atoi(strings.TrimSpace(r))
		if err == nil { break }
		fmt.Println("Please enter a valid number...")
	}
	masterTemplate := templates.LoadTemplate()
	
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	server := &http.Server{
		Addr: fmt.Sprintf("0.0.0.0:%d", portNum),
	}
	bindAllWebInstallerRoutes(&WebInstallerRoutingContext{
		Template: masterTemplate,
		Config: &aegis.AegisConfig{},
	})
	go func() {
		log.Printf("Trying to serve at %s:%d\n", "0.0.0.0", portNum)
		err := server.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
		log.Println("Stopped serving new connections.")
	}()

	<-sigChan
	
	if err := server.Shutdown(context.TODO()); err != nil {
		log.Fatalf("HTTP shutdown fail: %v", err)
	}
	
	log.Println("Graceful shutdown complete.")
}


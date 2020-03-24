package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ardanlabs/conf"
	"github.com/golangcollege/sessions"
	"github.com/pkg/errors"
	"github.com/tullo/snptx/internal/platform/database"
	"github.com/tullo/snptx/internal/snippet"
	"github.com/tullo/snptx/internal/user"

	"github.com/tullo/snptx/pkg/models/postgres"
)

// build is the git version of this application. It is set using build flags in the makefile.
var build = "develop"

// the key must be unexported type to avoid collisions
type contextKey string

const contextKeyIsAuthenticated = contextKey("isAuthenticated")

// define the interfaces inline to keep the code simple
type application struct {
	debug    bool
	errorLog *log.Logger
	infoLog  *log.Logger
	session  *sessions.Session
	shutdown chan os.Signal
	snippets interface {
		Insert(string, string, string) (string, error)
		Get(string) (*snippet.Snippet, error)
		Latest() ([]snippet.Snippet, error)
	}
	templateCache map[string]*template.Template
	users         interface {
		Insert(string, string, string) error
		Authenticate(string, string) (string, error)
		Get(string) (*user.User, error)
		ChangePassword(id string, currentPassword, newPassword string) error
	}
}

// SignalShutdown is used to gracefully shutdown the app when an integrity
// issue is identified.
func (a *application) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

func main() {
	if err := run(); err != nil {
		log.Printf("error: %s", err)
		os.Exit(1)
	}
}

func run() error {

	// =========================================================================
	// Configuration

	// session secret (should be 32 bytes long) is used to encrypt and authenticate session cookies
	// e.g. 'openssl rand -base64 32'

	var cfg struct {
		Web struct {
			APIHost         string        `conf:"default::4200"`
			DebugMode       bool          `conf:"default:false"`
			SessionSecret   string        `conf:"default:un/MjLYrdgFiQxAHDge/lI/kydfyZRo4T0UF+Mn4xag="`
			IdleTimeout     time.Duration `conf:"default:1m"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:5s"`
			ShutdownTimeout time.Duration `conf:"default:5s"`
		}
		DB struct {
			User       string `conf:"default:postgres"`
			Password   string `conf:"default:postgres,noprint"`
			Host       string `conf:"default:0.0.0.0"`
			Name       string `conf:"default:postgres"`
			DisableTLS bool   `conf:"default:false"`
		}
		Args conf.Args
	}

	if err := conf.Parse(os.Args[1:], "SNPTX", &cfg); err != nil {
		if err == conf.ErrHelpWanted {
			usage, err := conf.Usage("SNPTX", &cfg)
			if err != nil {
				return errors.Wrap(err, "generating usage")
			}
			fmt.Println(usage)
			return nil
		}
		return errors.Wrap(err, "error: parsing config")
	}

	// =========================================================================
	// Start Database

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	infoLog.Println("main : Started : Initializing database support")

	db, err := database.Open(database.Config{
		User:       cfg.DB.User,
		Password:   cfg.DB.Password,
		Host:       cfg.DB.Host,
		Name:       cfg.DB.Name,
		DisableTLS: cfg.DB.DisableTLS,
	})
	if err != nil {
		return errors.Wrap(err, "connecting to db")
	}
	defer func() {
		infoLog.Printf("main : Database Stopping : %s", cfg.DB.Host)
		db.Close()
	}()

	// =========================================================================
	// Start Web Application

	infoLog.Println("main : Started : Initializing web application")

	// initialize template cache
	templateCache, err := newTemplateCache("./ui/html/")
	if err != nil {
		errorLog.Fatal(err)
	}

	// sessions expire after 12 hours
	session := sessions.New([]byte(cfg.Web.SessionSecret))
	session.Lifetime = 12 * time.Hour
	// set the secure flag on session cookies and
	// serve all requests over https in production environment
	session.Secure = true
	session.SameSite = http.SameSiteStrictMode

	// make a channel to listen for an interrupt or terminate signal from the OS.
	// use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	app := &application{
		debug:         cfg.Web.DebugMode,
		errorLog:      errorLog,
		infoLog:       infoLog,
		session:       session,
		shutdown:      shutdown,
		snippets:      &postgres.SnippetModel{DB: db},
		templateCache: templateCache,
		users:         &postgres.UserModel{DB: db},
	}

	// use Go’s favored cipher suites (support for forward secrecy)
	// and elliptic curves that are performant under heavy loads
	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	srv := &http.Server{
		Addr:         cfg.Web.APIHost,
		ErrorLog:     errorLog,
		Handler:      app.routes(),
		TLSConfig:    tlsConfig,
		IdleTimeout:  cfg.Web.IdleTimeout,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
	}

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the application listening for requests.
	go func() {
		infoLog.Printf("Starting server on %s", cfg.Web.APIHost)
		serverErrors <- srv.ListenAndServeTLS("./tls/localhost/cert.pem", "./tls/localhost/key.pem")
	}()

	// =========================================================================
	// Shutdown

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return errors.Wrap(err, "server error")

	case sig := <-shutdown:
		infoLog.Printf("main : %v : Start shutdown", sig)

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		// Asking listener to shutdown and load shed.
		err := srv.Shutdown(ctx)
		if err != nil {
			infoLog.Printf("main : Graceful shutdown did not complete in %v : %v", cfg.Web.ShutdownTimeout, err)
			err = srv.Close()
		}

		// Log the status of this shutdown.
		switch {
		case sig == syscall.SIGSTOP:
			return errors.New("integrity issue caused shutdown")
		case err != nil:
			return errors.Wrap(err, "could not stop server gracefully")
		}
	}

	return nil
}

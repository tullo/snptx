package main

import (
	"context"
	"crypto/tls"
	"encoding/base64"
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
	"github.com/tullo/snptx/internal/platform/auth"
	"github.com/tullo/snptx/internal/platform/database"
	"github.com/tullo/snptx/internal/platform/sec"
	"github.com/tullo/snptx/internal/snippet"
	"github.com/tullo/snptx/internal/user"
)

// build is the git version of this application. It is set using build flags in the makefile.
var build = "develop"

// the key must be unexported type to avoid collisions
type contextKey string

const contextKeyIsAuthenticated = contextKey("isAuthenticated")

// define the interfaces inline to keep the code simple
type app struct {
	debug    bool
	log      *log.Logger
	session  *sessions.Session
	shutdown chan os.Signal
	snippets interface {
		Create(context.Context, snippet.NewSnippet, time.Time) (*snippet.Info, error)
		Delete(context.Context, string) error
		Latest(context.Context) ([]snippet.Info, error)
		Update(context.Context, string, snippet.UpdateSnippet, time.Time) error
		Retrieve(context.Context, string) (*snippet.Info, error)
	}
	templateCache map[string]*template.Template
	users         interface {
		Authenticate(context.Context, time.Time, string, string) (auth.Claims, error)
		Create(context.Context, user.NewUser, time.Time) (*user.Info, error)
		ChangePassword(context.Context, string, string, string) error
		QueryByID(context.Context, string) (*user.Info, error)
	}
	version string
}

// SignalShutdown is used to gracefully shutdown the app when an integrity
// issue is identified.
func (a *app) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

func main() {
	log := log.New(os.Stdout, "SNPTX : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	if err := run(log); err != nil {
		log.Println("main: error:", err)
		os.Exit(1)
	}
}

func run(log *log.Logger) error {

	// =========================================================================
	// Configuration

	// session secret (must be 32 bytes long) is used to encrypt and authenticate session cookies
	// e.g. 'openssl rand -base64 32'

	var cfg struct {
		Web struct {
			APIHost         string        `conf:"default::4200"`
			DebugMode       bool          `conf:"default:false"`
			SessionSecret   string        `conf:"noprint"`
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
		Aragon struct {
			// Note: Changing the value of Parallelism - changes the hash output!
			Memory      uint `conf:"default:131072"` // 128 * 1024 (KB) - memory used by the Argon2 algorithm
			Iterations  uint `conf:"default:4"`      // number of passes over the memory
			Parallelism uint `conf:"default:4"`      // number of threads to use on a machine with multiple cores
			SaltLength  uint `conf:"default:16"`     // 16 bytes is recommended for password hashing
			KeyLength   uint `conf:"default:32"`     // length of the generated password hash
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

	log.Println("Initializing Database support")

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
		log.Printf("Database Stopping : %s", cfg.DB.Host)
		db.Close()
	}()

	// =========================================================================
	// Start Web Application

	log.Printf("Initializing Application: version %q\n", build)

	out, err := conf.String(&cfg)
	if err != nil {
		return errors.Wrap(err, "generating config for output")
	}
	log.Printf("Config:\n%v\n", out)

	// parameters used for password hashing
	hp := sec.HashParams{
		Memory:      uint32(cfg.Aragon.Memory),
		Iterations:  uint32(cfg.Aragon.Iterations),
		Parallelism: uint8(cfg.Aragon.Parallelism),
		SaltLength:  uint32(cfg.Aragon.SaltLength),
		KeyLength:   uint32(cfg.Aragon.KeyLength),
	}

	// initialize template cache
	templateCache, err := newTemplateCache("./ui/html/")
	if err != nil {
		return err
	}

	decoded, err := base64.StdEncoding.DecodeString(cfg.Web.SessionSecret)
	if err != nil {
		return errors.Wrap(err, "decoding session secret")
	}
	if len(decoded) != 32 {
		return errors.New("session secret must be exactly 32 bytes long")
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

	snippets := snippet.New(db)
	users := user.New(db, sec.Params(hp))

	app := &app{
		debug:         cfg.Web.DebugMode,
		log:           log,
		session:       session,
		shutdown:      shutdown,
		snippets:      snippets,
		templateCache: templateCache,
		users:         users,
		version:       build,
	}

	// use Go’s favored cipher suites (support for forward secrecy)
	// and elliptic curves that are performant under heavy loads
	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
	}

	srv := &http.Server{
		Addr:         cfg.Web.APIHost,
		ErrorLog:     log,
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
		log.Printf("Starting server on %s (%s)", cfg.Web.APIHost, build[:7])
		serverErrors <- srv.ListenAndServeTLS("./tls/localhost/cert.pem", "./tls/localhost/key.pem")
	}()

	// =========================================================================
	// Shutdown

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return errors.Wrap(err, "server error")

	case sig := <-shutdown:
		log.Printf("Received signal [%v] Start shutdown", sig)

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		// Asking listener to shutdown and load shed.
		err := srv.Shutdown(ctx)
		if err != nil {
			log.Printf("Graceful shutdown did not complete in %v : %v", cfg.Web.ShutdownTimeout, err)
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

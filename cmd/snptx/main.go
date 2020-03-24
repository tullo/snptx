package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
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

func main() {
	if err := run(); err != nil {
		log.Printf("error: %s", err)
		os.Exit(1)
	}
}

func run() error {

	// =========================================================================
	// Configuration

	var cfg struct {
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

	// This is used for multiple commands below.
	dbConfig := database.Config{
		User:       cfg.DB.User,
		Password:   cfg.DB.Password,
		Host:       cfg.DB.Host,
		Name:       cfg.DB.Name,
		DisableTLS: cfg.DB.DisableTLS,
	}

	addr := flag.String("addr", ":4200", "HTTP network address")
	debug := flag.Bool("debug", false, "Enable debug mode")
	// force the db driver to convert TIME and DATE fields to time.Time (parseTime=true)
	//dsn := flag.String("dsn", "web:snptx@tcp(0.0.0.0:3306)/snptx?parseTime=true", "MySQL data source name")
	// session secret (should be 32 bytes long) is used to encrypt and authenticate session cookies
	// e.g. 'openssl rand -base64 32'
	secret := flag.String("secret", "un/MjLYrdgFiQxAHDge/lI/kydfyZRo4T0UF+Mn4xag=", "Secret key")

	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	//db, err := openDB(*dsn)
	db, err := database.Open(dbConfig)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	// initialize template cache
	templateCache, err := newTemplateCache("./ui/html/")
	if err != nil {
		errorLog.Fatal(err)
	}

	// sessions expire after 12 hours
	session := sessions.New([]byte(*secret))
	session.Lifetime = 12 * time.Hour
	// set the secure flag on session cookies and
	// serve all requests over https in production environment
	session.Secure = true
	session.SameSite = http.SameSiteStrictMode

	app := &application{
		debug:         *debug,
		errorLog:      errorLog,
		infoLog:       infoLog,
		session:       session,
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
		Addr:         *addr,
		ErrorLog:     errorLog,
		Handler:      app.routes(),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServeTLS("./tls/localhost/cert.pem", "./tls/localhost/key.pem")
	errorLog.Fatal(err)

	return err
}

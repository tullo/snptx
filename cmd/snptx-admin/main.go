package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/tullo/conf"
	"github.com/tullo/snptx/internal/platform/database"
	"github.com/tullo/snptx/internal/schema"
)

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
			User       string `conf:"default:admin"`
			Password   string `conf:"default:postgres,noprint"`
			Host       string `conf:"default:0.0.0.0:26257"`
			Name       string `conf:"default:postgres"`
			DisableTLS bool   `conf:"default:false"`
		}
		Args conf.Args
	}

	if err := conf.Parse(os.Args[1:], "ADMIN", &cfg); err != nil {
		if err == conf.ErrHelpWanted {
			usage, err := conf.Usage("ADMIN", &cfg)
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

	var err error
	switch cfg.Args.Num(0) {
	case "migrate":
		err = migrate(dbConfig)
	case "seed":
		err = seed(dbConfig)
	default:
		err = errors.New("Must specify a command")
	}

	if err != nil {
		return err
	}

	return nil
}

func migrate(cfg database.Config) error {
	if err := schema.Migrate(database.ConnString(cfg)); err != nil {
		return err
	}

	fmt.Println("Migrations complete")

	return nil
}

func seed(cfg database.Config) error {
	deadline := time.Now().Add(time.Second * 15)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	db, err := database.Connect(ctx, cfg)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := schema.Seed(ctx, db); err != nil {
		return err
	}

	fmt.Println("Seed data complete")
	return nil
}

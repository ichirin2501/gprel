package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/ichirin2501/gprel"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

// Version is the version of gprel
const Version = "0.1.0"

func showVersionString() string {
	return fmt.Sprintf(
		"gprel version %s built with %s %s %s",
		Version, runtime.Version(), runtime.GOOS, runtime.GOARCH,
	)
}

func run(c *gprel.Configuration) error {

	if c.DryRun {
		log.Info("Dry-run mode")
	}
	log.Info("purge relay-log started")

	var dsn string
	if c.Socket != "" {
		dsn = fmt.Sprintf("%s:%s@unix(%s)/%s?interpolateParams=true", c.User, c.Password, c.Socket, c.DatabaseName)
	} else {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?interpolateParams=true", c.User, c.Password, c.Host, c.Port, c.DatabaseName)
	}
	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	purger := gprel.NewPurger(db, c.PurgeDelaySeconds, c.DryRun)

	priv, err := purger.HasPurgePrivilege()
	if err != nil {
		return err
	}
	if !priv {
		return errors.New("the user doesn't have purge privilege")
	}
	log.Info("purge privilege OK")

	if err := purger.Purge(); err != nil {
		return err
	}
	log.Info("relay-log purging operations succeeded")

	if c.DryRun {
		log.Warn("Dry-run have finished. Please specify -go if you want to purge relay-log")
	}

	return nil
}

func main() {
	c, err := gprel.ParseOptions(os.Args)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	if c.ShowVersion {
		fmt.Println(showVersionString())
		os.Exit(0)
	}

	if err := run(c); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

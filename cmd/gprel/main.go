package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ichirin2501/gprel"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func run(c *gprel.Configuration) error {
	log.Println("purge relay-log started")

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

	purger := gprel.NewPurger(db, c.PurgeDelaySeconds)

	if err := purger.Purge(); err != nil {
		return err
	}
	log.Println("OK")
	log.Println("relay-log purging operations succeeded")

	return nil
}

func main() {
	log.SetOutput(os.Stdout)

	c, err := gprel.ParseOptions(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := run(c); err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", time.Now().Format("2006/01/02 15:04:05"), err)
		os.Exit(1)
	}
}

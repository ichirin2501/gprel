package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ichirin2501/gprel/config"

	_ "github.com/go-sql-driver/mysql"
)

const (
	ExitCodeOK = iota
	ExitCodeError
)

func isSlave(db *sql.DB) (bool, error) {
	rows, err := db.Query("SHOW SLAVE STATUS")
	if err != nil {
		return false, err
	}
	defer rows.Close()
	if !rows.Next() { // Not a slave
		return false, nil
	}
	return true, nil
}

func isRelayLogPurge(db *sql.DB) (bool, error) {
	var v int
	err := db.QueryRow("SELECT @@global.relay_log_purge AS Value").Scan(&v)
	if err != nil {
		return false, err
	}
	return v == 1, nil
}

func purgeRelayLogs(db *sql.DB) error {
	if _, err := db.Exec("SET GLOBAL relay_log_purge = 1"); err != nil {
		return err
	}
	if _, err := db.Exec("FLUSH NO_WRITE_TO_BINLOG RELAY LOGS"); err != nil {
		return err
	}
	time.Sleep(time.Duration(config.Config.PurgeDelaySeconds) * time.Second)
	if _, err := db.Exec("SET GLOBAL relay_log_purge = 0"); err != nil {
		return err
	}
	return nil
}

func run() (int, error) {
	if err := config.ParseOptions(); err != nil {
		return ExitCodeError, err
	}

	log.Println("purge relay-log started")

	var dsn string
	if config.Config.Socket != "" {
		dsn = fmt.Sprintf("%s:%s@unix(%s)/%s?interpolateParams=true", config.Config.User, config.Config.Password, config.Config.Socket, config.Config.DatabaseName)
	} else {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?interpolateParams=true", config.Config.User, config.Config.Password, config.Config.Host, config.Config.Port, config.Config.DatabaseName)
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return ExitCodeError, err
	}
	defer db.Close()

	if ok, err := isSlave(db); !ok {
		if err == nil {
			return 1, fmt.Errorf("target mysql server is not defined as replication slave")
		}
		return ExitCodeError, err
	}

	if ok, err := isRelayLogPurge(db); ok {
		if err == nil {
			return 1, fmt.Errorf("relay_log_purge is enabled")
		}
		return ExitCodeError, err
	}

	log.Println("Executing SET GLOBAL relay_log_purge=1; FLUSH NO_WRITE_TO_BINLOG RELAY LOGS; sleeping..... SET GLOBAL relay_log_purge=0; ...")
	if err := purgeRelayLogs(db); err != nil {
		return ExitCodeError, err
	}
	log.Println("OK")
	log.Println("relay-log purging operations succeeded")

	return ExitCodeOK, nil
}

func main() {
	log.SetOutput(os.Stdout)

	exitCode, err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s %v\n", time.Now().Format("2006/1/2 15:04:05"), err)
	}
	os.Exit(exitCode)
}

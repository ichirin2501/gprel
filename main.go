package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ichirin2501/gprel/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

const (
	ExitCodeOK = iota
	ExitCodeError
)

func isIOSQLRunning(db *sqlx.DB) (bool, error) {
	rows, err := db.Queryx("SHOW SLAVE STATUS")
	if err != nil {
		return false, err
	}
	defer rows.Close()

	var isSlave bool
	for rows.Next() {
		isSlave = true
		d := make(map[string]interface{})
		if err := rows.MapScan(d); err != nil {
			return false, err
		}

		io := d["Slave_IO_Running"].([]byte)
		sql := d["Slave_SQL_Running"].([]byte)

		if string(io) != "Yes" {
			return false, nil
		}
		if string(sql) != "Yes" {
			return false, nil
		}
	}

	if !isSlave {
		return false, fmt.Errorf("not a slave")
	}

	return true, nil
}

func isRelayLogPurge(db *sqlx.DB) (bool, error) {
	var v int
	err := db.QueryRow("SELECT @@global.relay_log_purge AS Value").Scan(&v)
	if err != nil {
		return false, err
	}
	return v == 1, nil
}

func purgeRelayLogs(db *sqlx.DB) error {
	log.Println("Executing FLUSH NO_WRITE_TO_BINLOG RELAY LOGS")
	if _, err := db.Exec("FLUSH NO_WRITE_TO_BINLOG RELAY LOGS"); err != nil {
		return err
	}

	log.Println("Executing sleep delay...")
	time.Sleep(time.Duration(config.Config.PurgeDelaySeconds) * time.Second)

	// last check
	log.Println("check SQL/IO Thread state")
	if ok, err := isIOSQLRunning(db); !ok {
		if err == nil {
			return fmt.Errorf("stop slave?")
		}
		return err
	}

	log.Println("Executing SET GLOBAL relay_log_purge = 1")
	if _, err := db.Exec("SET GLOBAL relay_log_purge = 1"); err != nil {
		return err
	}
	log.Println("Executing FLUSH NO_WRITE_TO_BINLOG RELAY LOGS (again)")
	if _, err := db.Exec("FLUSH NO_WRITE_TO_BINLOG RELAY LOGS"); err != nil {
		return err
	}

	time.Sleep(3 * time.Second)

	log.Println("Executing SET GLOBAL relay_log_purge = 0")
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
	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return ExitCodeError, err
	}
	defer db.Close()

	if ok, err := isIOSQLRunning(db); !ok {
		if err == nil {
			return ExitCodeError, fmt.Errorf("SQL or IO Thread is not running")
		}
		return ExitCodeError, err
	}

	if ok, err := isRelayLogPurge(db); ok {
		if err == nil {
			return ExitCodeError, fmt.Errorf("relay_log_purge is enabled")
		}
		return ExitCodeError, err
	}

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
		fmt.Fprintf(os.Stderr, "%s %v\n", time.Now().Format("2006/01/02 15:04:05"), err)
	}
	os.Exit(exitCode)
}

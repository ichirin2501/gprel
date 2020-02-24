package gprel

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	log "github.com/sirupsen/logrus"
)

type Purger struct {
	db           *sqlx.DB
	DelaySeconds int
	DryRun       bool
}

func NewPurger(db *sqlx.DB, delay int, dryRun bool) *Purger {
	return &Purger{
		db:           db,
		DelaySeconds: delay,
		DryRun:       dryRun,
	}
}

func (p *Purger) isIOSQLThreadRunning() (bool, error) {
	rows, err := p.db.Queryx("SHOW SLAVE STATUS")
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

func (p *Purger) isRelayLogPurge() (bool, error) {
	var v int
	err := p.db.QueryRow("SELECT @@global.relay_log_purge AS Value").Scan(&v)
	if err != nil {
		return false, err
	}
	return v == 1, nil
}

func (p *Purger) Purge() error {
	if ok, err := p.isIOSQLThreadRunning(); !ok {
		if err == nil {
			return fmt.Errorf("SQL or IO Thread is not running")
		}
		return err
	}
	if ok, err := p.isRelayLogPurge(); ok {
		if err == nil {
			return fmt.Errorf("relay_log_purge is enabled")
		}
		return err
	}

	log.Debug("Executing FLUSH NO_WRITE_TO_BINLOG RELAY LOGS")
	if !p.DryRun {
		if _, err := p.db.Exec("FLUSH NO_WRITE_TO_BINLOG RELAY LOGS"); err != nil {
			return err
		}
	}

	log.Debug("Executing sleep delay...")
	time.Sleep(time.Duration(p.DelaySeconds) * time.Second)

	// last check
	log.Debug("check SQL/IO Thread state")
	if ok, err := p.isIOSQLThreadRunning(); !ok {
		if err == nil {
			return fmt.Errorf("stop slave?")
		}
		return err
	}

	log.Debug("Executing SET GLOBAL relay_log_purge = 1")
	if !p.DryRun {
		if _, err := p.db.Exec("SET GLOBAL relay_log_purge = 1"); err != nil {
			return err
		}
	}
	log.Debug("Executing FLUSH NO_WRITE_TO_BINLOG RELAY LOGS (again)")
	if !p.DryRun {
		if _, err := p.db.Exec("FLUSH NO_WRITE_TO_BINLOG RELAY LOGS"); err != nil {
			return err
		}
	}

	time.Sleep(3 * time.Second)

	log.Debug("Executing SET GLOBAL relay_log_purge = 0")
	if !p.DryRun {
		if _, err := p.db.Exec("SET GLOBAL relay_log_purge = 0"); err != nil {
			return err
		}
	}
	return nil
}

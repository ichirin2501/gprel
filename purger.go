package gprel

import (
	"context"
	"errors"
	"fmt"
	"strings"
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

var errNotReplica = errors.New("not replica")

func (p *Purger) isIOSQLThreadRunning(ctx context.Context) (bool, error) {
	rows, err := p.db.QueryxContext(ctx, `SHOW SLAVE STATUS`)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	var isReplica bool
	for rows.Next() {
		isReplica = true
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
	if err := rows.Err(); err != nil {
		return false, err
	}

	if !isReplica {
		return false, errNotReplica
	}

	return true, nil
}

func (p *Purger) isRelayLogPurge(ctx context.Context) (bool, error) {
	var v int
	if err := p.db.QueryRowxContext(ctx, `SELECT @@global.relay_log_purge AS Value`).Scan(&v); err != nil {
		return false, err
	}
	return v == 1, nil
}

func (p *Purger) HasPurgePrivilege(ctx context.Context) (bool, error) {
	rows, err := p.db.QueryxContext(ctx, `show grants for current_user()`)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	var (
		hasSUPER             bool
		hasRELOAD            bool
		hasReplicationClient bool
	)
	for rows.Next() {
		var (
			grantData string
		)
		if err := rows.Scan(&grantData); err != nil {
			return false, err
		}
		if strings.Contains(grantData, `GRANT ALL PRIVILEGES ON *.*`) {
			return true, nil
		}
		if strings.Contains(grantData, `SUPER`) && strings.Contains(grantData, ` ON *.*`) {
			hasSUPER = true
		}
		if strings.Contains(grantData, `RELOAD`) && strings.Contains(grantData, ` ON *.*`) {
			hasRELOAD = true
		}
		if strings.Contains(grantData, `REPLICATION CLIENT`) && strings.Contains(grantData, ` ON *.*`) {
			hasReplicationClient = true
		}
	}
	if err := rows.Err(); err != nil {
		return false, err
	}
	return hasSUPER && hasRELOAD && hasReplicationClient, nil
}

func (p *Purger) Purge(ctx context.Context) error {
	if ok, err := p.isIOSQLThreadRunning(ctx); !ok {
		if err == nil {
			log.Info("SQL or IO Thread is not running")
			return nil
		}
		if err == errNotReplica {
			log.Info(err)
			return nil
		}
		return err
	}
	if ok, err := p.isRelayLogPurge(ctx); ok {
		if err == nil {
			return fmt.Errorf("relay_log_purge is enabled")
		}
		return err
	}

	log.Debug("Executing FLUSH NO_WRITE_TO_BINLOG RELAY LOGS")
	if !p.DryRun {
		if _, err := p.db.ExecContext(ctx, "FLUSH NO_WRITE_TO_BINLOG RELAY LOGS"); err != nil {
			return err
		}
	}

	log.Debug("Executing sleep delay...")
	delayTimer := time.NewTimer(time.Duration(p.DelaySeconds) * time.Second)
	defer delayTimer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-delayTimer.C:
	}

	// last check
	log.Debug("check SQL/IO Thread state")
	if ok, err := p.isIOSQLThreadRunning(ctx); !ok {
		if err == nil {
			return fmt.Errorf("stop replication?")
		}
		return err
	}

	log.Debug("Executing SET GLOBAL relay_log_purge = 1")
	if !p.DryRun {
		if _, err := p.db.ExecContext(ctx, "SET GLOBAL relay_log_purge = 1"); err != nil {
			return err
		}
	}

	err := func() error {
		log.Debug("Executing FLUSH NO_WRITE_TO_BINLOG RELAY LOGS (again)")
		if !p.DryRun {
			if _, err := p.db.ExecContext(ctx, "FLUSH NO_WRITE_TO_BINLOG RELAY LOGS"); err != nil {
				return err
			}
		}
		timer := time.NewTimer(3 * time.Second)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
		}
		return nil
	}()
	if err != nil {
		// clean up
		if !p.DryRun {
			log.Info("Try a cleanup process...")
			c, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			_, e := p.db.ExecContext(c, "SET GLOBAL relay_log_purge = 0")
			if e == nil {
				log.Info("The cleanup process has been completed")
			} else {
				log.Warn("The cleanup process failed")
			}
		}
		return err
	}

	log.Debug("Executing SET GLOBAL relay_log_purge = 0")
	if !p.DryRun {
		if _, err := p.db.ExecContext(ctx, "SET GLOBAL relay_log_purge = 0"); err != nil {
			return err
		}
	}
	return nil
}

package stats

import (
	"database/sql"
	"fmt"
	"log"
)

type migrations struct {
	db         *sql.DB
	migrations []migration
}

func newMigrations(db *sql.DB) *migrations {
	return &migrations{
		db: db,
		migrations: []migration{
			&migration001_initial{},
		},
	}
}

func (m *migrations) run() error {
	version, err := m.getMigrationVersion()
	if err != nil {
		return fmt.Errorf("error running migrations: %v", err)
	}
	for number, migration := range m.migrations[version:] {
		log.Printf("running migration: %s", migration.name())
		if err := m.runMigration(migration); err != nil {
			return fmt.Errorf("error running migration number %d: %v", number, err)
		}
		if err := m.saveVersionNumber(number + 1); err != nil {
			return fmt.Errorf("error saving updated number %d: %v", number+1, err)
		}
	}
	return nil
}

func (m *migrations) getMigrationVersion() (int, error) {
	tx, err := m.db.Begin()
	if err != nil {
		return 0, err
	}

	if _, err := tx.Exec(`CREATE TABLE IF NOT EXISTS migrations (version INT NOT NULL)`); err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("error creating migrations table: %v", err)
	}

	var version int
	if err := tx.QueryRow(`SELECT version FROM migrations LIMIT 1`).Scan(&version); err != nil {
		if err == sql.ErrNoRows {
			if _, err := tx.Exec(`INSERT INTO migrations (version) VALUES (0)`); err != nil {
				tx.Rollback()
				return 0, fmt.Errorf("error inserting initial migration value: %v", err)
			} else {
				tx.Commit()
				return 0, nil
			}
		} else {
			tx.Rollback()
			return 0, err
		}
	}

	return version, nil
}

func (m *migrations) saveVersionNumber(version int) error {
	if _, err := m.db.Exec(`UPDATE migrations SET version = $1`, version); err != nil {
		return fmt.Errorf("error saving version number %d: %v", version, err)
	}
	return nil
}

func (m *migrations) runMigration(migration migration) error {
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("unable to start transaction for migation: %v", err)
	}
	err = migration.migrate(tx)
	if err != nil {
		tx.Rollback()
		err = fmt.Errorf("unable to run migration %s: %v", migration.name(), err)
		return err
	}
	err = tx.Commit()
	if err != nil {
		err = fmt.Errorf("unable to commit migration %s: %v", migration.name(), err)
		return err
	}
	return nil
}

type migration interface {
	name() string
	migrate(tx *sql.Tx) error
}

type migration001_initial struct{}

func (m *migration001_initial) name() string { return "migration001_initial" }
func (m *migration001_initial) migrate(tx *sql.Tx) error {
	if _, err := tx.Exec(`
CREATE TABLE stats (
  id        BIGSERIAL PRIMARY KEY    NOT NULL,
  stat      VARCHAR(24)              NOT NULL,
  value     FLOAT                    NOT NULL,
  timestamp TIMESTAMP WITH TIME ZONE NOT NULL
);
CREATE INDEX idx_stats_stat
  ON stats (stat);

CREATE TABLE logs (
  id        BIGSERIAL PRIMARY KEY    NOT NULL,
  level     INT                      NOT NULL,
  message   VARCHAR(1024)            NOT NULL,
  timestamp TIMESTAMP WITH TIME ZONE NOT NULL
);
CREATE INDEX idx_logs_level
  ON logs (level);

  `); err != nil {
		return err
	}
	return nil
}

package stats

import (
	"database/sql"

	"github.com/explodes/migrations-go"
)

func migrateSqliteDatabase(db *sql.DB) error {
	migrator := migrations.NewMigrator(db, storageSqliteMigrations{})
	return migrator.MigrateToVersion(versionSqliteLatest)
}

type storageSqliteMigrations struct{}

func (storageSqliteMigrations) GetMigration(version int) migrations.Migration {
	switch version {
	case versionSqliteInitial:
		// empty string, downgrade not supported
		return migrations.NewSimpleMigration("initial", upgradeSqliteInitial, downgradeSqliteInitial)
	}
	return nil
}

const (
	versionSqliteInitial = 1
	versionSqliteLatest  = versionSqliteInitial
)

const (
	upgradeSqliteInitial = `
CREATE TABLE stats (
  id        INTEGER PRIMARY KEY AUTOINCREMENT,
  stat      INTEGER NOT NULL,
  value     FLOAT   NOT NULL,
  nanostamp INTEGER NOT NULL
);
CREATE INDEX idx_stats_stat
  ON stats (stat);

CREATE TABLE logs (
  id        INTEGER PRIMARY KEY AUTOINCREMENT,
  level     INTEGER NOT NULL,
  message   TEXT    NOT NULL,
  nanostamp INTEGER NOT NULL
);
CREATE INDEX idx_logs_level
  ON logs (level);
  `
	downgradeSqliteInitial = `
DROP TABLE logs;
DROP TABLE stats;
`
)

package stats

import (
	"database/sql"

	"github.com/explodes/migrations-go"
)

func migrateDatabase(db *sql.DB) error {
	migrator := migrations.NewMigrator(db, storageMigrations{})
	return migrator.MigrateToVersion(versionLatest)
}

type storageMigrations struct{}

func (storageMigrations) GetMigration(version int) migrations.Migration {
	switch version {
	case versionInitial:
		// empty string, downgrade not supported
		return migrations.NewSimpleMigration("initial", upgradeInitial, downgradeInitial)
	}
	return nil
}

const (
	versionInitial = 1
	versionLatest  = versionInitial
)

const (
	upgradeInitial = `
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

  `
	downgradeInitial = `
DROP TABLE logs;
DROP TABLE stats;
`
)

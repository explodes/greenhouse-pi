package stats

import (
	"database/sql"

	"github.com/explodes/migrations-go"
)

func migratePgDatabase(db *sql.DB) error {
	migrator := migrations.NewMigrator(db, storagePgMigrations{})
	return migrator.MigrateToVersion(versionPgLatest)
}

type storagePgMigrations struct{}

func (storagePgMigrations) GetMigration(version int) migrations.Migration {
	switch version {
	case versionPgInitial:
		// empty string, downgrade not supported
		return migrations.NewSimpleMigration("initial", upgradePgInitial, downgradePgInitial)
	}
	return nil
}

const (
	versionPgInitial = 1
	versionPgLatest  = versionPgInitial
)

const (
	upgradePgInitial = `
CREATE TABLE stats (
  id        BIGSERIAL PRIMARY KEY    NOT NULL,
  stat      INTEGER                  NOT NULL,
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
	downgradePgInitial = `
DROP TABLE logs;
DROP TABLE stats;
`
)

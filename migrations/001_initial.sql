--
-- Migrate Up: Build our database
--
-- +migrate Up
--
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
--
-- Migrate Down: Reverse the process
--
-- +migrate Down
--
DROP TABLE logs;
DROP TABLE stats;
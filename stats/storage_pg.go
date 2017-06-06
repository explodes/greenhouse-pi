package stats

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/explodes/greenhouse-pi/logging"
	"github.com/rubenv/sql-migrate"
)

const (
	pgDriver = "postgres"
)

type pgStorage struct {
	db *sql.DB
}

func NewPgStorage(conn, migrationsDir string) (Storage, error) {
	db, err := sql.Open(pgDriver, conn)
	if err != nil {
		return nil, fmt.Errorf("error opening database storage: %v", err)
	}
	storage := &pgStorage{
		db: db,
	}
	if err := storage.migrate(migrationsDir); err != nil {
		return nil, fmt.Errorf("error preparing database storage: %v", err)
	}
	return storage, nil
}

func (pg *pgStorage) migrate(migrationsDir string) error {
	migrationFiles := &migrate.FileMigrationSource{
		Dir: migrationsDir,
	}
	n, err := migrate.Exec(pg.db, pgDriver, migrationFiles, migrate.Up)
	if n > 0 {
		pg.Log(logging.LevelInfo, "ran %d migrations", n)
	}
	return err
}

func (pg *pgStorage) Record(stat Stat) error {
	_, err := pg.db.Exec(`INSERT INTO stats (stat, value, timestamp) VALUES($1, $2, $3)`, stat.StatType, stat.Value, stat.When)
	return err
}

func (pg *pgStorage) Fetch(statType StatType, start, end time.Time) ([]Stat, error) {
	scan := struct {
		value float64
		when  time.Time
	}{}
	rows, err := pg.db.Query(`SELECT value, timestamp FROM stats WHERE stat = $1 AND timestamp BETWEEN $2 AND $3 ORDER BY timestamp DESC LIMIT 1000`, statType, start, end)
	if err != nil {
		return nil, fmt.Errorf("error fetching stats: %v", err)
	}
	defer rows.Close()

	results := make([]Stat, 0, 100)
	for rows.Next() {
		if err := rows.Scan(&scan.value, &scan.when); err != nil {
			return nil, fmt.Errorf("error scanning stats: %v", err)
		}
		entry := Stat{
			StatType: statType,
			Value:    scan.value,
			When:     scan.when,
		}
		results = append(results, entry)
	}
	return results, nil
}

func (pg *pgStorage) Latest(statType StatType) (Stat, error) {
	scan := struct {
		value     float64
		timestamp time.Time
	}{}
	err := pg.db.QueryRow(`SELECT value, timestamp FROM stats WHERE stat = $1 ORDER BY timestamp DESC LIMIT 1`, statType).Scan(&scan.value, &scan.timestamp)
	if err == sql.ErrNoRows {
		return Stat{}, ErrNoStats
	}
	stat := Stat{
		StatType: statType,
		When:     scan.timestamp,
		Value:    scan.value,
	}
	return stat, nil
}

func (pg *pgStorage) Log(level logging.Level, format string, args ...interface{}) (logging.LogEntry, error) {
	entry := logging.LogEntry{
		Message: fmt.Sprintf(format, args...),
		Level:   level,
		When:    time.Now(),
	}
	log.Println(entry.Message)
	if _, err := pg.db.Exec(`INSERT INTO logs (message, timestamp, level) VALUES($1, $2, $3)`, entry.Message, entry.When, entry.Level); err != nil {
		return entry, err
	}
	return entry, nil
}

func (pg *pgStorage) Logs(level logging.Level, start, end time.Time) ([]logging.LogEntry, error) {
	scan := struct {
		level   int
		message string
		when    time.Time
	}{}
	rows, err := pg.db.Query(`SELECT message, timestamp, level FROM logs WHERE level >= $1 AND timestamp BETWEEN $2 AND $3 ORDER BY timestamp DESC LIMIT 1000`, level, start, end)
	if err != nil {
		return nil, fmt.Errorf("error fetching logs: %v", err)
	}
	defer rows.Close()

	results := make([]logging.LogEntry, 0, 100)
	for rows.Next() {
		if err := rows.Scan(&scan.message, &scan.when, &scan.level); err != nil {
			return nil, fmt.Errorf("error scanning logs: %v", err)
		}
		entry := logging.LogEntry{
			Level:   logging.Level(scan.level),
			Message: scan.message,
			When:    scan.when,
		}
		results = append(results, entry)
	}
	return results, nil
}

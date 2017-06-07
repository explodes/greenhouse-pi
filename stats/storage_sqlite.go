package stats

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/explodes/greenhouse-pi/logging"
)

const (
	sqliteDriver = "sqlite3"
)

type sqliteStorage struct {
	db *sql.DB
}

func NewSqliteStorage(conn string) (Storage, error) {
	db, err := sql.Open(sqliteDriver, conn)
	if err != nil {
		return nil, fmt.Errorf("error opening database storage: %v", err)
	}
	storage := &sqliteStorage{
		db: db,
	}
	if err := storage.migrate(); err != nil {
		return nil, fmt.Errorf("error preparing database storage: %v", err)
	}
	return storage, nil
}

func (ss *sqliteStorage) migrate() error {
	return migrateSqliteDatabase(ss.db)
}

func (ss *sqliteStorage) Record(stat Stat) error {
	_, err := ss.db.Exec(`INSERT INTO stats (stat, value, nanostamp) VALUES($1, $2, $3)`, stat.StatType, stat.Value, stat.When.UnixNano())
	return err
}

func (ss *sqliteStorage) Fetch(statType StatType, start, end time.Time) ([]Stat, error) {
	scan := struct {
		value     float64
		nanostamp int64
	}{}
	rows, err := ss.db.Query(`SELECT value, nanostamp FROM stats WHERE stat = $1 AND nanostamp > $2 AND nanostamp < $3 ORDER BY nanostamp DESC LIMIT 1000`, statType, start.UnixNano(), end.UnixNano())
	if err != nil {
		return nil, fmt.Errorf("error fetching stats: %v", err)
	}
	defer rows.Close()

	results := make([]Stat, 0, 100)
	for rows.Next() {
		if err := rows.Scan(&scan.value, &scan.nanostamp); err != nil {
			return nil, fmt.Errorf("error scanning stats: %v", err)
		}
		entry := Stat{
			StatType: statType,
			Value:    scan.value,
			When:     time.Unix(0, scan.nanostamp),
		}
		results = append(results, entry)
	}
	return results, nil
}

func (ss *sqliteStorage) Latest(statType StatType) (Stat, error) {
	scan := struct {
		value     float64
		nanostamp int64
	}{}
	err := ss.db.QueryRow(`SELECT value, nanostamp FROM stats WHERE stat = $1 ORDER BY nanostamp DESC LIMIT 1`, statType).Scan(&scan.value, &scan.nanostamp)
	if err == sql.ErrNoRows {
		return Stat{}, ErrNoStats
	}
	stat := Stat{
		StatType: statType,
		When:     time.Unix(0, scan.nanostamp),
		Value:    scan.value,
	}
	return stat, nil
}

func (ss *sqliteStorage) Log(level logging.Level, format string, args ...interface{}) (logging.LogEntry, error) {
	entry := logging.LogEntry{
		Message: fmt.Sprintf(format, args...),
		Level:   level,
		When:    time.Now(),
	}
	log.Println(entry.Message)
	if _, err := ss.db.Exec(`INSERT INTO logs (message, nanostamp, level) VALUES($1, $2, $3)`, entry.Message, entry.When.UnixNano(), entry.Level); err != nil {
		return entry, err
	}
	return entry, nil
}

func (ss *sqliteStorage) Logs(level logging.Level, start, end time.Time) ([]logging.LogEntry, error) {
	scan := struct {
		level     int
		message   string
		nanostamp int64
	}{}
	rows, err := ss.db.Query(`SELECT message, nanostamp, level FROM logs WHERE level >= $1 AND nanostamp > $2 AND nanostamp < $3 ORDER BY nanostamp DESC LIMIT 1000`, level, start.UnixNano(), end.UnixNano())
	if err != nil {
		return nil, fmt.Errorf("error fetching logs: %v", err)
	}
	defer rows.Close()

	results := make([]logging.LogEntry, 0, 100)
	for rows.Next() {
		if err := rows.Scan(&scan.message, &scan.nanostamp, &scan.level); err != nil {
			return nil, fmt.Errorf("error scanning logs: %v", err)
		}
		entry := logging.LogEntry{
			Level:   logging.Level(scan.level),
			Message: scan.message,
			When:    time.Unix(0, scan.nanostamp),
		}
		results = append(results, entry)
	}
	return results, nil
}

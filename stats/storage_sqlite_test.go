package stats

import (
	"reflect"
	"testing"
	"time"

	"github.com/explodes/greenhouse-pi/logging"
	_ "github.com/mattn/go-sqlite3"
)

func sqliteTest(f func(t *testing.T, s *sqliteStorage)) (string, func(*testing.T)) {
	name := testFunctionName(f)
	testFunc := func(t *testing.T) {
		t.Parallel()

		storage, err := NewSqliteStorage(":memory:")
		if err != nil {
			t.Fatal(err)
		}
		defer storage.Close()

		f(t, storage.(*sqliteStorage))
	}

	return name, testFunc
}

func TestSqliteStorage(t *testing.T) {
	t.Parallel()
	t.Run("Migrations", func(t *testing.T) {
		t.Parallel()
		t.Run(sqliteTest(sqlite_FullMigration))
	})
	t.Run("Storage", func(t *testing.T) {
		t.Parallel()
		t.Run(sqliteTest(sqlite_RecordLatestFetch))
		t.Run(sqliteTest(sqlite_Logging))
	})
}

func sqlite_FullMigration(t *testing.T, s *sqliteStorage) {
	var version int
	if err := s.db.QueryRow(`SELECT MAX(version) FROM migrations`).Scan(&version); err != nil {
		t.Fatal(err)
	}

	if version != versionSqliteLatest {
		t.Error("migrations not correctly or fully run")
	}
}

func sqlite_RecordLatestFetch(t *testing.T, s *sqliteStorage) {
	fuzz := Stat{StatType: StatTypeFan, Value: 1, When: time.Now()}
	if err := s.Record(fuzz); err != nil {
		t.Fatal(err)
	}
	stat := Stat{StatType: StatTypeWater, Value: 1, When: time.Now()}
	if err := s.Record(stat); err != nil {
		t.Fatal(err)
	}
	latest, err := s.Latest(StatTypeWater)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(stat, latest) {
		t.Fatalf("unexpected stat\nneed: %#v\nhave: %#v", stat, latest)
	}
	history, err := s.Fetch(StatTypeWater, time.Now().Add(-2*time.Hour), time.Now().Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 0 {
		t.Fatalf("unexpected history: %#v", history)
	}
	history, err = s.Fetch(StatTypeWater, time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 1 {
		t.Fatalf("unexpected history: %#v", history)
	}
	if !reflect.DeepEqual(stat, history[0]) {
		t.Fatalf("unexpected history\nneed: %#v\nhave: %#v", stat, history[0])
	}
}

func sqlite_Logging(t *testing.T, s *sqliteStorage) {
	if _, err := s.Log(logging.LevelDebug, "fuzz"); err != nil {
		t.Fatal(err)
	}
	var log logging.LogEntry
	var err error
	if log, err = s.Log(logging.LevelInfo, "test"); err != nil {
		t.Fatal(err)
	}

	logs, err := s.Logs(logging.LevelWarn, time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 0 {
		t.Fatalf("unexpected log entries: %#v", logs)
	}
	logs, err = s.Logs(logging.LevelInfo, time.Now().Add(-2*time.Hour), time.Now().Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 0 {
		t.Fatalf("unexpected log entries: %#v", logs)
	}
	logs, err = s.Logs(logging.LevelInfo, time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 1 {
		t.Fatalf("unexpected log entries: %#v", logs)
	}
	if !reflect.DeepEqual(log, logs[0]) {
		t.Fatalf("unexpected log entries\nneed: %#v\nhave: %#v", log, logs[0])
	}
}

//+build integration

package stats

import (
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/explodes/greenhouse-pi/logging"
	_ "github.com/lib/pq"
)

const (
	envPgConnection = "TEST_PG_CONNECTION"
)

func pgTest(f func(t *testing.T, s *pgStorage)) (string, func(*testing.T)) {
	name := testFunctionName(f)
	testFunc := func(t *testing.T) {
		storage, err := NewPostgresStorage(os.Getenv(envPgConnection))
		if err != nil {
			t.Fatal(err)
		}

		pg := storage.(*pgStorage)
		defer func() {
			errs := []error{}
			if _, err := pg.db.Exec(`drop table if exists stats`); err != nil {
				errs = append(errs, err)
			}
			if _, err := pg.db.Exec(`drop table if exists logs`); err != nil {
				errs = append(errs, err)
			}
			if _, err := pg.db.Exec(`drop table if exists migrations`); err != nil {
				errs = append(errs, err)
			}
			if err := storage.Close(); err != nil {
				errs = append(errs, err)
			}
			if len(errs) != 0 {
				fmt.Printf("error cleaning up test: %v", errs)
				t.Fatal(errs)
			}
		}()

		f(t, pg)
	}

	return name, testFunc
}

func TestPostgresStorage(t *testing.T) {

	if os.Getenv(envPgConnection) == "" {
		t.Fatalf("%s not set. Need database connection string.", envPgConnection)
	}

	t.Run("Migrations", func(t *testing.T) {
		t.Run(pgTest(pg_FullMigration))
	})
	t.Run("Storage", func(t *testing.T) {
		t.Run(pgTest(pg_RecordLatestFetch))
		t.Run(pgTest(pg_Logging))
	})
}

func pg_FullMigration(t *testing.T, s *pgStorage) {
	var version int
	if err := s.db.QueryRow(`SELECT MAX(version) FROM migrations`).Scan(&version); err != nil {
		t.Fatal(err)
	}

	if version != versionSqliteLatest {
		t.Error("migrations not correctly or fully run")
	}
}

func pg_RecordLatestFetch(t *testing.T, s *pgStorage) {
	fuzz := Stat{StatType: StatTypeFan, Value: 1, When: time.Now()}
	if err := s.Record(fuzz); err != nil {
		t.Fatal(err)
	}
	stat := Stat{StatType: StatTypeWater, Value: 1, When: time.Now()}
	if err := s.Record(stat); err != nil {
		t.Fatal(err)
	}

	stat.When = time.Time{} // hack to compare times

	latest, err := s.Latest(StatTypeWater)
	if err != nil {
		t.Fatal(err)
	}
	latest.When = time.Time{} // hack to compare times

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

	history[0].When = time.Time{} // hack to compare times

	if !reflect.DeepEqual(stat, history[0]) {
		t.Fatalf("unexpected history\nneed: %#v\nhave: %#v", stat, history[0])
	}
}

func pg_Logging(t *testing.T, s *pgStorage) {
	if _, err := s.Log(logging.LevelDebug, "fuzz"); err != nil {
		t.Fatal(err)
	}
	var log logging.LogEntry
	var err error
	if log, err = s.Log(logging.LevelInfo, "test"); err != nil {
		t.Fatal(err)
	}

	log.When = time.Time{} // hack to compare times

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

	logs[0].When = time.Time{} // hack to compare times
	if !reflect.DeepEqual(log, logs[0]) {
		t.Fatalf("unexpected log entries\nneed: %#v\nhave: %#v", log, logs[0])
	}
}

package api_test

import (
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/explodes/greenhouse-pi/api"
	"github.com/explodes/greenhouse-pi/controllers"
	"github.com/explodes/greenhouse-pi/logging"
	"github.com/explodes/greenhouse-pi/stats"
)

const (
	iso8601 = "2006-01-02T15:04:05-07:00"
)

func functionName(i interface{}) string {
	qname := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	parts := strings.Split(qname, "/")
	qname = parts[len(parts)-1]
	parts = strings.Split(qname, ".")
	return strings.Join(parts[1:], ".")
}

func testFunctionName(i interface{}) string {
	name := functionName(i)
	parts := strings.Split(name, "_")
	if len(parts) < 2 {
		panic("Test name must be in <function>_<Condition> format")
	}
	return strings.Join(parts[1:], "_")
}

func apiViewTest(f func(t *testing.T, a *api.Api, w *responseWriterRecorder)) (string, func(*testing.T)) {
	name := testFunctionName(f)
	testFunc := func(t *testing.T) {
		t.Parallel()

		scheduler := controllers.NewScheduler()
		storage := stats.NewFakeStatsStorage(10)
		water, err := controllers.NewController(controllers.NewFakeUnit(stats.StatTypeWater, storage), storage, scheduler)
		if err != nil {
			t.Fatal(err)
		}
		fan, err := controllers.NewController(controllers.NewFakeUnit(stats.StatTypeFan, storage), storage, scheduler)
		if err != nil {
			t.Fatal(err)
		}

		a := api.New(storage, water, fan)
		w := NewResponseWriterRecorder()

		f(t, a, w)
	}

	return name, testFunc
}

func TestApiView(t *testing.T) {
	t.Parallel()
	t.Run("History", func(t *testing.T) {
		t.Parallel()
		t.Run(apiViewTest(history_OK))
		t.Run(apiViewTest(history_OKwithValues))
		t.Run(apiViewTest(history_MissingStat))
		t.Run(apiViewTest(history_MissingStart))
		t.Run(apiViewTest(history_MissingEnd))
	})
	t.Run("Latest", func(t *testing.T) {
		t.Parallel()
		t.Run(apiViewTest(latest_OK))
		t.Run(apiViewTest(latest_OKwithValues))
		t.Run(apiViewTest(latest_MissingStat))
	})
	t.Run("Status", func(t *testing.T) {
		t.Parallel()
		t.Run(apiViewTest(status_OK))
		t.Run(apiViewTest(status_OKwithValues))
	})
	t.Run("Schedule", func(t *testing.T) {
		t.Parallel()
		t.Run(apiViewTest(schedule_OK))
		t.Run(apiViewTest(schedule_invalidStat))
		t.Run(apiViewTest(schedule_MissingStat))
		t.Run(apiViewTest(schedule_MissingStart))
		t.Run(apiViewTest(schedule_MissingEnd))
	})
	t.Run("Logs", func(t *testing.T) {
		t.Parallel()
		t.Run(apiViewTest(logs_OK))
		t.Run(apiViewTest(logs_OKwithValues))
		t.Run(apiViewTest(logs_MissingLevel))
		t.Run(apiViewTest(logs_InvalidLevel))
		t.Run(apiViewTest(logs_MissingStart))
		t.Run(apiViewTest(logs_MissingEnd))
	})
}

func history_OK(t *testing.T, a *api.Api, w *responseWriterRecorder) {
	start := time.Now().Add(-time.Hour).Format(iso8601)
	end := time.Now().Format(iso8601)

	a.History(w, nil, map[string]string{
		"stat":  "temperature",
		"start": start,
		"end":   end,
	})

	w.Assert(t).
		StatusEquals(http.StatusOK).
		JsonBodyEquals(map[string]interface{}{
			"start": start,
			"end":   end,
			"stat":  "temperature",
			"items": []api.KnownStat{},
		})
}

func history_OKwithValues(t *testing.T, a *api.Api, w *responseWriterRecorder) {
	when1 := time.Now().Add(-time.Minute)
	a.Storage.Record(stats.Stat{StatType: stats.StatTypeTemperature, When: when1, Value: 1})
	when2 := time.Now().Add(-time.Second)
	a.Storage.Record(stats.Stat{StatType: stats.StatTypeTemperature, When: when2, Value: 2})
	start := time.Now().Add(-time.Hour).Format(iso8601)
	end := time.Now().Format(iso8601)

	a.History(w, nil, map[string]string{
		"stat":  "temperature",
		"start": start,
		"end":   end,
	})

	w.Assert(t).
		StatusEquals(http.StatusOK).
		JsonBodyEquals(map[string]interface{}{
			"start": start,
			"end":   end,
			"stat":  "temperature",
			"items": []api.KnownStat{
				{When: when1, Value: 1},
				{When: when2, Value: 2},
			}})
}

func history_MissingStat(t *testing.T, a *api.Api, w *responseWriterRecorder) {
	start := time.Now().Add(-time.Hour).Format(iso8601)
	end := time.Now().Format(iso8601)

	a.History(w, nil, map[string]string{
		"start": start,
		"end":   end,
	})

	w.Assert(t).
		StatusEquals(http.StatusBadRequest).
		StringBodyEquals(`{"error":"missing stat"}`)
}

func history_MissingStart(t *testing.T, a *api.Api, w *responseWriterRecorder) {
	end := time.Now().Format(iso8601)

	a.History(w, nil, map[string]string{
		"stat": "temperature",
		"end":  end,
	})

	w.Assert(t).
		StatusEquals(http.StatusBadRequest).
		StringBodyEquals(`{"error":"missing start time"}`)
}

func history_MissingEnd(t *testing.T, a *api.Api, w *responseWriterRecorder) {
	start := time.Now().Add(-time.Hour).Format(iso8601)

	a.History(w, nil, map[string]string{
		"stat":  "temperature",
		"start": start,
	})

	w.Assert(t).
		StatusEquals(http.StatusBadRequest).
		StringBodyEquals(`{"error":"missing end time"}`)
}

func latest_OK(t *testing.T, a *api.Api, w *responseWriterRecorder) {
	a.Latest(w, nil, map[string]string{
		"stat": "temperature",
	})

	w.Assert(t).
		StatusEquals(http.StatusOK).
		JsonBodyEquals(map[string]interface{}{
			"stat":  "temperature",
			"value": float64(0),
		})
}

func latest_OKwithValues(t *testing.T, a *api.Api, w *responseWriterRecorder) {
	when1 := time.Now().Add(-time.Minute)
	a.Storage.Record(stats.Stat{StatType: stats.StatTypeTemperature, When: when1, Value: 1})
	when2 := time.Now().Add(-time.Second)
	a.Storage.Record(stats.Stat{StatType: stats.StatTypeTemperature, When: when2, Value: 2})

	a.Latest(w, nil, map[string]string{
		"stat": "temperature",
	})

	w.Assert(t).
		StatusEquals(http.StatusOK).
		JsonBodyEquals(map[string]interface{}{
			"stat":  "temperature",
			"value": float64(2),
		})
}

func latest_MissingStat(t *testing.T, a *api.Api, w *responseWriterRecorder) {
	start := time.Now().Add(-time.Hour).Format(iso8601)
	end := time.Now().Format(iso8601)

	a.Latest(w, nil, map[string]string{
		"start": start,
		"end":   end,
	})

	w.Assert(t).
		StatusEquals(http.StatusBadRequest).
		StringBodyEquals(`{"error":"missing stat"}`)
}

func status_OK(t *testing.T, a *api.Api, w *responseWriterRecorder) {
	a.Status(w, nil, nil)

	w.Assert(t).
		StatusEquals(http.StatusOK).
		JsonBodyEquals(map[string]interface{}{
			"water":       map[string]interface{}{"status": "off"},
			"fan":         map[string]interface{}{"status": "off"},
			"temperature": map[string]interface{}{"value": float64(0)},
			"humidity":    map[string]interface{}{"value": float64(0)},
		})
}

func status_OKwithValues(t *testing.T, a *api.Api, w *responseWriterRecorder) {
	when1 := time.Now().Add(-time.Minute)
	a.Storage.Record(stats.Stat{StatType: stats.StatTypeTemperature, When: when1, Value: 1})
	when2 := time.Now().Add(-time.Second)
	a.Storage.Record(stats.Stat{StatType: stats.StatTypeHumidity, When: when2, Value: 2})

	a.Status(w, nil, nil)

	w.Assert(t).
		StatusEquals(http.StatusOK).
		JsonBodyEquals(map[string]interface{}{
			"water":       map[string]interface{}{"status": "off"},
			"fan":         map[string]interface{}{"status": "off"},
			"temperature": map[string]interface{}{"value": float64(1)},
			"humidity":    map[string]interface{}{"value": float64(2)},
		})
}

func schedule_OK(t *testing.T, a *api.Api, w *responseWriterRecorder) {
	start := time.Now().Add(time.Hour).Format(iso8601)
	end := time.Now().Add(2 * time.Hour).Format(iso8601)

	a.Schedule(w, nil, map[string]string{
		"stat":  "water",
		"start": start,
		"end":   end,
	})

	w.Assert(t).StatusEquals(http.StatusNoContent)
}

func schedule_MissingStat(t *testing.T, a *api.Api, w *responseWriterRecorder) {
	start := time.Now().Add(time.Hour).Format(iso8601)
	end := time.Now().Add(2 * time.Hour).Format(iso8601)

	a.Schedule(w, nil, map[string]string{
		"start": start,
		"end":   end,
	})

	w.Assert(t).
		StatusEquals(http.StatusBadRequest).
		StringBodyEquals(`{"error":"missing stat"}`)
}

func schedule_invalidStat(t *testing.T, a *api.Api, w *responseWriterRecorder) {
	start := time.Now().Add(time.Hour).Format(iso8601)
	end := time.Now().Add(2 * time.Hour).Format(iso8601)

	a.Schedule(w, nil, map[string]string{
		"stat":  "temperature",
		"start": start,
		"end":   end,
	})

	w.Assert(t).
		StatusEquals(http.StatusBadRequest).
		StringBodyEquals(`{"error":"invalid stat type"}`)
}

func schedule_MissingStart(t *testing.T, a *api.Api, w *responseWriterRecorder) {
	end := time.Now().Add(2 * time.Hour).Format(iso8601)

	a.Schedule(w, nil, map[string]string{
		"stat": "water",
		"end":  end,
	})

	w.Assert(t).
		StatusEquals(http.StatusBadRequest).
		StringBodyEquals(`{"error":"missing start time"}`)
}

func schedule_MissingEnd(t *testing.T, a *api.Api, w *responseWriterRecorder) {
	start := time.Now().Add(time.Hour).Format(iso8601)

	a.Schedule(w, nil, map[string]string{
		"stat":  "water",
		"start": start,
	})

	w.Assert(t).
		StatusEquals(http.StatusBadRequest).
		StringBodyEquals(`{"error":"missing end time"}`)
}

func logs_OK(t *testing.T, a *api.Api, w *responseWriterRecorder) {
	start := time.Now().Add(-time.Hour).Format(iso8601)
	end := time.Now().Add(time.Hour).Format(iso8601)

	a.Logs(w, nil, map[string]string{
		"level": "debug",
		"start": start,
		"end":   end,
	})

	w.Assert(t).
		StatusEquals(http.StatusOK).
		JsonBodyEquals(map[string]interface{}{
			"items": make([]map[string]interface{}, 0),
		})
}

func logs_OKwithValues(t *testing.T, a *api.Api, w *responseWriterRecorder) {
	entry, err := a.Storage.Log(logging.LevelDebug, "hello world")
	if err != nil {
		t.Fatal(err)
	}
	start := time.Now().Add(-time.Hour).Format(iso8601)
	end := time.Now().Add(time.Hour).Format(iso8601)

	a.Logs(w, nil, map[string]string{
		"level": "debug",
		"start": start,
		"end":   end,
	})

	w.Assert(t).
		StatusEquals(http.StatusOK).
		JsonBodyEquals(map[string]interface{}{
			"items": []map[string]interface{}{
				{
					"level":   "debug",
					"when":    entry.When,
					"message": "hello world",
				},
			}})
}

func logs_MissingLevel(t *testing.T, a *api.Api, w *responseWriterRecorder) {
	start := time.Now().Add(-time.Hour).Format(iso8601)
	end := time.Now().Add(time.Hour).Format(iso8601)

	a.Logs(w, nil, map[string]string{
		"start": start,
		"end":   end,
	})

	w.Assert(t).
		StatusEquals(http.StatusBadRequest).
		StringBodyEquals(`{"error":"missing log level"}`)
}

func logs_InvalidLevel(t *testing.T, a *api.Api, w *responseWriterRecorder) {
	start := time.Now().Add(-time.Hour).Format(iso8601)
	end := time.Now().Add(time.Hour).Format(iso8601)

	a.Logs(w, nil, map[string]string{
		"level": "invalid",
		"start": start,
		"end":   end,
	})

	w.Assert(t).
		StatusEquals(http.StatusBadRequest).
		StringBodyEquals(`{"error":"invalid log level"}`)
}

func logs_MissingStart(t *testing.T, a *api.Api, w *responseWriterRecorder) {
	end := time.Now().Add(2 * time.Hour).Format(iso8601)

	a.Logs(w, nil, map[string]string{
		"level": "debug",
		"end":   end,
	})

	w.Assert(t).
		StatusEquals(http.StatusBadRequest).
		StringBodyEquals(`{"error":"missing start time"}`)
}

func logs_MissingEnd(t *testing.T, a *api.Api, w *responseWriterRecorder) {
	start := time.Now().Add(time.Hour).Format(iso8601)

	a.Logs(w, nil, map[string]string{
		"level": "debug",
		"start": start,
	})

	w.Assert(t).
		StatusEquals(http.StatusBadRequest).
		StringBodyEquals(`{"error":"missing end time"}`)
}

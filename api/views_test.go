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
		water, err := controllers.NewController(controllers.NewFakeUnit("fake-water"), storage, scheduler)
		if err != nil {
			t.Fatal(err)
		}
		fan, err := controllers.NewController(controllers.NewFakeUnit("fake-fan"), storage, scheduler)
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
		StatusEquals(t, http.StatusOK).
		JsonBodyEquals(t, map[string]interface{}{
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
		StatusEquals(t, http.StatusOK).
		JsonBodyEquals(t, map[string]interface{}{
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
		StatusEquals(t, http.StatusBadRequest).
		StringBodyEquals(t, "missing stat")
}

func history_MissingStart(t *testing.T, a *api.Api, w *responseWriterRecorder) {
	end := time.Now().Format(iso8601)

	a.History(w, nil, map[string]string{
		"stat": "temperature",
		"end":  end,
	})

	w.Assert(t).
		StatusEquals(t, http.StatusBadRequest).
		StringBodyEquals(t, "missing start time")
}

func history_MissingEnd(t *testing.T, a *api.Api, w *responseWriterRecorder) {
	start := time.Now().Add(-time.Hour).Format(iso8601)

	a.History(w, nil, map[string]string{
		"stat":  "temperature",
		"start": start,
	})

	w.Assert(t).
		StatusEquals(t, http.StatusBadRequest).
		StringBodyEquals(t, "missing end time")
}

func latest_OK(t *testing.T, a *api.Api, w *responseWriterRecorder) {
	a.Latest(w, nil, map[string]string{
		"stat": "temperature",
	})

	w.Assert(t).
		StatusEquals(t, http.StatusOK).
		JsonBodyEquals(t, map[string]interface{}{
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
		StatusEquals(t, http.StatusOK).
		JsonBodyEquals(t, map[string]interface{}{
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
		StatusEquals(t, http.StatusBadRequest).
		StringBodyEquals(t, "missing stat")
}

func status_OK(t *testing.T, a *api.Api, w *responseWriterRecorder) {
	a.Status(w, nil, nil)

	w.Assert(t).
		StatusEquals(t, http.StatusOK).
		JsonBodyEquals(t, map[string]interface{}{
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
		StatusEquals(t, http.StatusOK).
		JsonBodyEquals(t, map[string]interface{}{
		"water":       map[string]interface{}{"status": "off"},
		"fan":         map[string]interface{}{"status": "off"},
		"temperature": map[string]interface{}{"value": float64(1)},
		"humidity":    map[string]interface{}{"value": float64(2)},
	})
}

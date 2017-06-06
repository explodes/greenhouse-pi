package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/explodes/greenhouse-pi/stats"
)

var (
	errInvalidStat = errors.New("invalid stat type")
)

func validateStat(name string) (stats.StatType, error) {
	switch name {
	case string(stats.StatTypeTemperature):
		return stats.StatTypeTemperature, nil
	case string(stats.StatTypeHumidity):
		return stats.StatTypeHumidity, nil
	case string(stats.StatTypeWater):
		return stats.StatTypeWater, nil
	default:
		return stats.StatType(""), errInvalidStat
	}
}

func parseTime(s string) (time.Time, error) {
	var err error
	var result time.Time
	for _, format := range dateInputFormats {
		result, err = time.Parse(format, s)
		if err == nil {
			return result, nil
		}
	}
	return time.Time{}, err
}

func convertStatsToResponse(stats []stats.Stat) []KnownStat {
	results := make([]KnownStat, 0, len(stats))
	for _, stat := range stats {
		results = append(results, KnownStat{
			When:  stat.When,
			Value: stat.Value,
		})
	}
	return results
}

// History returns the history for a given stat and date range
func (api *Api) History(w http.ResponseWriter, r *http.Request, vars map[string]string) {
	// extract stat type
	// input
	statTypeRaw, ok := vars["stat"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing stat"))
		return
	}
	// parse
	statType, err := validateStat(statTypeRaw)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid stat type"))
		return
	}

	// extract start date
	// input
	startRaw, ok := vars["start"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing start time"))
		return
	}
	// parse
	start, err := parseTime(startRaw)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid start time"))
		return
	}

	// extract end date
	// input
	endRaw, ok := vars["end"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing end time"))
		return
	}
	// parse
	end, err := parseTime(endRaw)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid end time"))
		return
	}

	results, err := api.Storage.Fetch(statType, start, end)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("error fetching results: %v", err)))
		return
	}

	body, err := json.Marshal(map[string]interface{}{
		"start": start,
		"end":   end,
		"stat":  statType,
		"items": convertStatsToResponse(results),
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("unable to marshal json: %v", err)))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

// Latest returns the current value of a statistic
func (api *Api) Latest(w http.ResponseWriter, r *http.Request, vars map[string]string) {
	// extract stat type
	// input
	statTypeRaw, ok := vars["stat"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing stat"))
		return
	}
	// parse
	statType, err := validateStat(statTypeRaw)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid stat type"))
		return
	}

	var value float64
	result, err := api.Storage.Latest(statType)
	if err == stats.ErrNoStats {
		value = 0
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("error fetching results: %v", err)))
		return
	} else {
		value = result.Value
	}

	body, err := json.Marshal(map[string]interface{}{
		"stat":  statType,
		"value": value,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("unable to marshal json: %v", err)))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

// Status returns the current status of the system
func (api *Api) Status(w http.ResponseWriter, r *http.Request, vars map[string]string) {

	results := make(map[string]interface{}, 4)

	putResult := func(key, valueKey string, err error, value interface{}) {
		if err != nil {
			results[key] = map[string]interface{}{
				"error": err,
			}
		} else {
			results[key] = map[string]interface{}{
				valueKey: value,
			}
		}
	}

	waterStatus, waterErr := api.Water.Unit.Status()
	putResult("water", "status", waterErr, waterStatus)

	fanStatus, fanErr := api.Fan.Unit.Status()
	putResult("fan", "status", fanErr, fanStatus)

	temp, tempErr := api.Storage.Latest(stats.StatTypeTemperature)
	if tempErr == stats.ErrNoStats {
		tempErr = nil
	}
	putResult("temperature", "value", tempErr, temp.Value)

	humidity, humidityErr := api.Storage.Latest(stats.StatTypeHumidity)
	if humidityErr == stats.ErrNoStats {
		humidityErr = nil
	}
	putResult("humidity", "value", humidityErr, humidity.Value)

	body, err := json.Marshal(results)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("unable to marshal json: %v", err)))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

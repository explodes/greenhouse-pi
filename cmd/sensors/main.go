package main

import (
	"fmt"
	"time"

	"github.com/explodes/greenhouse-pi/api"
	"github.com/explodes/greenhouse-pi/monitor"
	"github.com/explodes/greenhouse-pi/sensors"
	"github.com/explodes/greenhouse-pi/stats"
)

const (
	sensorFrq = 100 * time.Millisecond
	logFrq    = 1000 * time.Millisecond
)

func main() {
	thermometer := sensors.NewFakeThermometer(sensorFrq)
	hygrometer := sensors.NewFakeHygrometer(sensorFrq)
	storage := stats.NewFakeStatsStorage(40)

	sensorMonitor := monitor.Monitor{
		Thermometer: thermometer,
		Hygrometer:  hygrometer,
		Storage:     storage,
	}

	go sensorMonitor.Begin()
	go logStatsLoop(storage)

	api.New(storage).Serve(":8096")
}

func logStatsLoop(storage stats.Storage) {
	logTimer := time.NewTicker(logFrq).C
	for {
		select {
		case <-logTimer:
			values, err := storage.Fetch(stats.StatTypeTemperature, time.Time{}, time.Now())
			logStats(stats.StatTypeTemperature, values, err)

			values, err = storage.Fetch(stats.StatTypeHumidity, time.Time{}, time.Now())
			logStats(stats.StatTypeHumidity, values, err)
		}
	}
}

func logStats(statType stats.StatType, stats []stats.Stat, err error) {
	if err != nil {
		fmt.Printf("error fetching %s: %v\n", statType, err)
		return
	}
	fmt.Printf("%s %s: ", time.Now().Format("2006-01-02 15:04:05"), statType)
	for _, stat := range stats {
		fmt.Printf("%.2g ", stat.Value)
	}
	fmt.Println()
}

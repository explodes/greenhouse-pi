package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/explodes/greenhouse-pi/api"
	"github.com/explodes/greenhouse-pi/monitor"
	"github.com/explodes/greenhouse-pi/sensors"
	"github.com/explodes/greenhouse-pi/stats"
)

var (
	flagBind      = flag.String("bind", ":8096", "Bind address for the API server")
	flagLogFrq    = flag.Int("logfrq", defaultLogFrq, "How frequently to log current sensor values to the console. 0 disables logging.")
	flagSensorFrq = flag.Int("sensorfrq", defaultSensorFrq, fmt.Sprintf("How frequently to read sensor values. Minimum %d", minSensorFreq))
)

const (
	defaultSensorFrq = 2000
	defaultLogFrq    = 2000
	minSensorFreq    = 100
)

func init() {
	flag.Parse()
}

func main() {
	validateFlags()

	sensorFrq := time.Duration(*flagSensorFrq) * time.Millisecond

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

	api.New(storage).Serve(*flagBind)
}

func validateFlags() {
	valid := true
	if *flagBind == "" {
		log.Printf("invalid bind address: %s", *flagBind)
		valid = false
	}
	if *flagLogFrq < 0 {
		log.Printf("logfrq must be a positive number")
		valid = false
	}
	if *flagSensorFrq < minSensorFreq {
		log.Printf("sensors value is invalid, or is too small. must be at least %dms", minSensorFreq)
		valid = false
	}

	if !valid {
		os.Exit(1)
	}
}

func logStatsLoop(storage stats.Storage) {
	if *flagLogFrq == 0 {
		return
	}
	frq := time.Duration(*flagLogFrq) * time.Millisecond
	logTimer := time.NewTicker(frq).C
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

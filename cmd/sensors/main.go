package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/explodes/greenhouse-pi/api"
	"github.com/explodes/greenhouse-pi/controllers"
	"github.com/explodes/greenhouse-pi/logging"
	"github.com/explodes/greenhouse-pi/monitor"
	"github.com/explodes/greenhouse-pi/sensors"
	"github.com/explodes/greenhouse-pi/stats"
)

var (
	flagBind      = flag.String("bind", ":8096", "Bind address for the API server")
	flagSensorFrq = flag.Int("sensorfrq", defaultSensorFrq, fmt.Sprintf("How frequently to read sensor values. Minimum %d", minSensorFreq))
)

const (
	defaultSensorFrq = 2000
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

	scheduler := controllers.NewScheduler()

	waterUnit := controllers.NewFakeUnit(stats.StatTypeWater, storage)
	waterController, err := controllers.NewController(waterUnit, storage, scheduler)
	if err != nil {
		log.Fatalf("unable to start water controller: %v", err)
	}

	fanUnit := controllers.NewFakeUnit(stats.StatTypeFan, storage)
	fanController, err := controllers.NewController(fanUnit, storage, scheduler)
	if err != nil {
		log.Fatalf("unable to start fan controller: %v", err)
	}
	if _, err := storage.Log(logging.LevelInfo, "sensors startup"); err != nil {
		log.Fatalf("error logging sensor startup: %v", err)
	}

	sensorMonitor := monitor.Monitor{
		Thermometer: thermometer,
		Hygrometer:  hygrometer,
		Storage:     storage,
	}

	go sensorMonitor.Begin()

	log.Fatal(api.New(storage, waterController, fanController).Serve(*flagBind))
}

func validateFlags() {
	valid := true
	if *flagBind == "" {
		log.Printf("invalid bind address: %s", *flagBind)
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

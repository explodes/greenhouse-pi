package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/explodes/greenhouse-pi/api"
	"github.com/explodes/greenhouse-pi/controllers"
	"github.com/explodes/greenhouse-pi/logging"
	"github.com/explodes/greenhouse-pi/monitor"
	"github.com/explodes/greenhouse-pi/sensors"
	"github.com/explodes/greenhouse-pi/stats"
)

var (
	flagBind      = flag.String("bind", "0.0.0.0:8096", "Bind address for the API server")
	flagSensorFrq = flag.Int("sensorfrq", defaultSensorFrq, fmt.Sprintf("How frequently to read sensor values. Minimum %d", minSensorFreq))
	flagDbConn    = flag.String("db", "mock://fake/40", "Database connection string")
	flagThermConn = flag.String("therm", "mock://fake", "Temperature sensor connection string")
	flagHygroConn = flag.String("hygro", "mock://fake", "Humidity sensor connection string")
	flagWaterConn = flag.String("water", "mock://fake", "Water unit connection string")
	flagFanConn   = flag.String("fan", "mock://fake", "Fan unit connection string")
)

const (
	defaultSensorFrq = 30000
	minSensorFreq    = 2000
)

func init() {
	flag.Parse()
}

func main() {
	validateFlags()

	storage, err := createStorage()
	if err != nil {
		log.Fatalf("error creating storage: %v", err)
	}

	sensorFrq := time.Duration(*flagSensorFrq) * time.Millisecond
	thermometer, err := createThermometer(sensorFrq)
	if err != nil {
		log.Fatalf("error creating thermometer: %v", err)
	}
	hygrometer, err := createHygrometer(sensorFrq)
	if err != nil {
		log.Fatalf("error creating hygrometer: %v", err)
	}

	if _, err := storage.Log(logging.LevelInfo, "sensors startup"); err != nil {
		log.Fatalf("error logging sensor startup: %v", err)
	}

	scheduler := controllers.NewScheduler()

	waterUnit, err := createWaterUnit(storage)
	if err != nil {
		log.Fatalf("error creating water unit: %v", err)
	}
	waterController, err := controllers.NewController(waterUnit, storage, scheduler)
	if err != nil {
		log.Fatalf("unable to start water controller: %v", err)
	}

	fanUnit, err := createFanUnit(storage)
	if err != nil {
		log.Fatalf("error creating fan unit: %v", err)
	}
	fanController, err := controllers.NewController(fanUnit, storage, scheduler)
	if err != nil {
		log.Fatalf("unable to start fan controller: %v", err)
	}

	if _, err := storage.Log(logging.LevelInfo, "unit controller startup"); err != nil {
		log.Fatalf("error logging sensor startup: %v", err)
	}

	sensorMonitor := monitor.Monitor{
		Thermometer: thermometer,
		Hygrometer:  hygrometer,
		Storage:     storage,
	}

	go sensorMonitor.Begin()
	if _, err := storage.Log(logging.LevelInfo, "sensor monitor startup"); err != nil {
		log.Fatalf("error logging monitor startup: %v", err)
	}

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

func createStorage() (stats.Storage, error) {
	conn := *flagDbConn
	if strings.Index(conn, "mock://") == 0 {
		parts := strings.Split(conn, "/")
		if len(parts) != 4 || parts[2] != "fake" {
			return nil, fmt.Errorf("bad fake-storage connection, expected mock://fake/40: %s", conn)
		}
		limit, err := strconv.Atoi(parts[3])
		if err != nil {
			return nil, fmt.Errorf("bad fake-storage storage limit: %s", parts[4])
		}
		return stats.NewFakeStatsStorage(limit), nil

	}
	return nil, fmt.Errorf("unknown database system: %s", conn)
}

func createThermometer(frq time.Duration) (sensors.Thermometer, error) {
	conn := *flagThermConn
	if conn == "mock://fake" {
		return sensors.NewFakeThermometer(frq), nil
	}
	return nil, fmt.Errorf("unknown thermometer: %s", conn)
}

func createHygrometer(frq time.Duration) (sensors.Hygrometer, error) {
	conn := *flagHygroConn
	if conn == "mock://fake" {
		return sensors.NewFakeHygrometer(frq), nil
	}
	return nil, fmt.Errorf("unknown hygrometer: %s", conn)
}

func createWaterUnit(storage stats.Storage) (controllers.Unit, error) {
	conn := *flagWaterConn
	if conn == "mock://fake" {
		return controllers.NewFakeUnit(stats.StatTypeWater, storage), nil
	}
	return nil, fmt.Errorf("unknown water unit: %s", conn)
}

func createFanUnit(storage stats.Storage) (controllers.Unit, error) {
	conn := *flagFanConn
	if conn == "mock://fake" {
		return controllers.NewFakeUnit(stats.StatTypeFan, storage), nil
	}
	return nil, fmt.Errorf("unknown fan unit: %s", conn)
}

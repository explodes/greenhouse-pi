package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/explodes/greenhouse-pi/api"
	"github.com/explodes/greenhouse-pi/cmd/builder"
	"github.com/explodes/greenhouse-pi/controllers"
	"github.com/explodes/greenhouse-pi/logging"
	"github.com/explodes/greenhouse-pi/monitor"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var (
	flagBind      = flag.String("bind", "0.0.0.0:8096", fmt.Sprintf("Bind address for the API server [%s]", envBind))
	flagSensorFrq = flag.Int("sensorfrq", defaultSensorFrq, fmt.Sprintf("How frequently to read sensor values in milliseconds. Minimum %dms [%s]", minSensorFreq, envSensorFrq))
	flagDbConn    = flag.String("db", "mock://fake/40", fmt.Sprintf("Database connection string (postgres: postgresql://user:pass@host/db, fake://mock/40, sqlite3:///usr/local/greenhouse/greenhouse.db) [%s]", envDbConn))
	flagThermConn = flag.String("therm", "mock://fake", fmt.Sprintf("Temperature sensor connection string [%s]", envThermConn))
	flagHygroConn = flag.String("hygro", "mock://fake", fmt.Sprintf("Humidity sensor connection string [%s]", envHygroConn))
	flagWaterConn = flag.String("water", "mock://fake", fmt.Sprintf("Water unit connection string [%s]", envWaterConn))
	flagFanConn   = flag.String("fan", "mock://fake", fmt.Sprintf("Fan unit connection string [%s]", envFanConn))
)

const (
	defaultSensorFrq = 30000
	minSensorFreq    = 2000

	envBind      = "GH_BIND"
	envSensorFrq = "GH_SENSOR_FRQ"
	envDbConn    = "GH_DATABASE"
	envThermConn = "GH_THERMOMETER"
	envHygroConn = "GH_HYGROMETER"
	envWaterConn = "GH_WATER"
	envFanConn   = "GH_FAN"
)

func init() {
	flag.Parse()
	mapEnvironmentVariableString(envBind, flagBind)
	mapEnvironmentVariableInt(envSensorFrq, flagSensorFrq)
	mapEnvironmentVariableString(envDbConn, flagDbConn)
	mapEnvironmentVariableString(envThermConn, flagThermConn)
	mapEnvironmentVariableString(envHygroConn, flagHygroConn)
	mapEnvironmentVariableString(envWaterConn, flagWaterConn)
	mapEnvironmentVariableString(envFanConn, flagFanConn)
	validateConfiguration()
}

func mapEnvironmentVariableString(env string, flag *string) {
	value := os.Getenv(env)
	if value != "" {
		*flag = value
	}
}

func mapEnvironmentVariableInt(env string, flag *int) {
	value := os.Getenv(env)
	if value != "" {
		valueInt, err := strconv.Atoi(value)
		if err != nil {
			log.Fatalf("Unable to parse %s as int, got %s", env, value)
		}
		*flag = valueInt
	}
}

func main() {

	storage, err := builder.CreateStorage(*flagDbConn)
	if err != nil {
		log.Fatalf("error creating storage: %v", err)
	}
	defer storage.Close()

	sensorFrq := time.Duration(*flagSensorFrq) * time.Millisecond
	thermometer, err := builder.CreateThermometer(*flagThermConn, sensorFrq)
	if err != nil {
		log.Fatalf("error creating thermometer: %v", err)
	}
	defer thermometer.Close()

	hygrometer, err := builder.CreateHygrometer(*flagHygroConn, sensorFrq)
	if err != nil {
		log.Fatalf("error creating hygrometer: %v", err)
	}
	defer hygrometer.Close()

	if _, err := storage.Log(logging.LevelInfo, "sensors startup"); err != nil {
		log.Fatalf("error logging sensor startup: %v", err)
	}

	scheduler := controllers.NewScheduler()

	waterUnit, err := builder.CreateWaterUnit(*flagWaterConn, storage)
	if err != nil {
		log.Fatalf("error creating water unit: %v", err)
	}
	defer waterUnit.Close()

	waterController, err := controllers.NewController(waterUnit, storage, scheduler)
	if err != nil {
		log.Fatalf("unable to start water controller: %v", err)
	}

	fanUnit, err := builder.CreateFanUnit(*flagFanConn, storage)
	if err != nil {
		log.Fatalf("error creating fan unit: %v", err)
	}
	defer fanUnit.Close()

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

	log.Fatal(api.New(storage, waterController, fanController, thermometer, hygrometer).Serve(*flagBind))
}

func validateConfiguration() {
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

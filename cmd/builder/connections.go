package builder

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/explodes/greenhouse-pi/controllers"
	"github.com/explodes/greenhouse-pi/sensors"
	"github.com/explodes/greenhouse-pi/stats"
)

func CreateStorage(conn string) (stats.Storage, error) {
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
	if strings.Index(conn, "postgresql://") == 0 {
		storage, err := stats.NewPostgresStorage(conn)
		if err != nil {
			return nil, fmt.Errorf("error connecting to database: %v", err)
		}
		return storage, nil
	}
	if strings.Index(conn, "sqlite3://") == 0 {
		conn = conn[len("sqlite3://"):]
		storage, err := stats.NewSqliteStorage(conn)
		if err != nil {
			return nil, fmt.Errorf("error connecting to database: %v", err)
		}
		return storage, nil
	}
	return nil, fmt.Errorf("unknown database system: %s", conn)
}

func CreateThermometer(conn string, frq time.Duration) (sensors.Thermometer, error) {
	if conn == "mock://fake" {
		return sensors.NewFakeThermometer(frq), nil
	}
	return nil, fmt.Errorf("unknown thermometer: %s", conn)
}

func CreateHygrometer(conn string, frq time.Duration) (sensors.Hygrometer, error) {
	if conn == "mock://fake" {
		return sensors.NewFakeHygrometer(frq), nil
	}
	return nil, fmt.Errorf("unknown hygrometer: %s", conn)
}

func CreateWaterUnit(conn string, storage stats.Storage) (controllers.Unit, error) {
	if conn == "mock://fake" {
		return controllers.NewFakeUnit(stats.StatTypeWater, storage), nil
	}
	return nil, fmt.Errorf("unknown water unit: %s", conn)
}

func CreateFanUnit(conn string, storage stats.Storage) (controllers.Unit, error) {
	if conn == "mock://fake" {
		return controllers.NewFakeUnit(stats.StatTypeFan, storage), nil
	}
	return nil, fmt.Errorf("unknown fan unit: %s", conn)
}

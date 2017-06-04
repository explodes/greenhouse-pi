package monitor

import (
	"time"

	"github.com/explodes/greenhouse-pi/sensors"
	"github.com/explodes/greenhouse-pi/stats"
)

type Monitor struct {
	Thermometer sensors.Thermometer
	Hygrometer  sensors.Hygrometer

	Storage stats.Storage
}

func (m *Monitor) Begin() {
	tempStream := m.Thermometer.Read()

	humidityStream := m.Hygrometer.Read()

	for {
		select {
		case temp := <-tempStream:
			m.Storage.Record(stats.Stat{
				StatType: stats.StatTypeTemperature,
				When:     time.Now(),
				Value:    float64(temp),
			})
		case humidity := <-humidityStream:
			m.Storage.Record(stats.Stat{
				StatType: stats.StatTypeHumidity,
				When:     time.Now(),
				Value:    float64(humidity),
			})

		}
	}
}

package stats

const (
	StatTypeTemperature = StatType("temperature")
	StatTypeHumidity    = StatType("humidity")
	StatTypeWater       = StatType("water")
)

type StatType string

package stats

const (
	StatTypeTemperature = StatType("temperature")
	StatTypeHumidity    = StatType("humidity")
	StatTypeWater       = StatType("water")
	StatTypeFan         = StatType("fan")
)

type StatType string

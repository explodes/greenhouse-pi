package stats

const (
	StatTypeTemperature StatType = 1 + iota
	StatTypeHumidity    StatType = 1 + iota
	StatTypeWater       StatType = 1 + iota
	StatTypeFan         StatType = 1 + iota
)

type StatType uint8

func (st StatType) String() string {
	switch st {
	case StatTypeTemperature:
		return "temperature"
	case StatTypeHumidity:
		return "humidity"
	case StatTypeWater:
		return "water"
	case StatTypeFan:
		return "fan"
	default:
		return "unknown"
	}
}

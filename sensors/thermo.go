package sensors

type Thermometer interface {
	Read() <-chan Temperature
}

type Temperature float64

func (t Temperature) Celsius() float64 {
	return float64(t)
}

func (t Temperature) Fahrenheit() float64 {
	return float64(t)*9./5. + 32.
}

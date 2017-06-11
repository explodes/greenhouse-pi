package sensors

import "time"

// Thermometer is a sensor made for reading temperature data
type Thermometer interface {
	// Read returns a channel on which
	// sensor data can be read from
	Read() <-chan Temperature

	// Frequency returns the frequency at
	// which  this sensor is reading values
	Frequency() time.Duration

	// Close the underlying connection
	// to the sensor. Read will no longer
	// be a valid channel
	Close() error
}

// Temperature is the value of Hygrometer data
// represented in celsius
type Temperature float64

// Celsius returns the celsius value of this temperature
func (t Temperature) Celsius() float64 {
	return float64(t)
}

// Fahrenheit returns the celsius value of this temperature
func (t Temperature) Fahrenheit() float64 {
	return float64(t)*9./5. + 32.
}

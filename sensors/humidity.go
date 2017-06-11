package sensors

import "time"

// Hygrometer is a sensor made for reading humidity data
type Hygrometer interface {
	// Read returns a channel on which
	// sensor data can be read from
	Read() <-chan Humidity

	// Frequency returns the frequency at
	// which  this sensor is reading values
	Frequency() time.Duration

	// Close the underlying connection
	// to the sensor. Read will no longer
	// be a valid channel
	Close() error
}

// Humidity is the value of Hygrometer data
// represented as a percent-humidity
type Humidity float64

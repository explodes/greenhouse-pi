package sensors

type Hygrometer interface {
	Read() <-chan Humidity
	Close() error
}

type Humidity float64

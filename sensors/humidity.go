package sensors

type Hygrometer interface {
	Read() <-chan Humidity
}

type Humidity float64

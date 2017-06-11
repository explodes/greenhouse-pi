package sensors

import (
	"log"
	"time"
)

const (
	fakeHygrometerFrq = 2 * time.Second
	fakeHygrometerMin = 40
	fakeHygrometerMax = 90
)

type fakeHygrometer struct {
	frq    time.Duration
	closed chan struct{}
}

func NewFakeHygrometer(frq time.Duration) Hygrometer {
	fake := &fakeHygrometer{
		frq:    frq,
		closed: make(chan struct{}),
	}
	return fake
}

func (f *fakeHygrometer) nextValue() Humidity {
	hum := Humidity(theRand.Float64()*(fakeHygrometerMax-fakeHygrometerMin) + fakeHygrometerMin)
	log.Printf("humidity: %g", hum)
	return hum
}

func (f *fakeHygrometer) Read() <-chan Humidity {
	results := make(chan Humidity)
	go func() {
		defer close(results)
		for {
			select {
			case <-f.closed:
				return
			case <-time.After(f.frq):
				results <- f.nextValue()
			}
		}
	}()
	return results
}

func (f *fakeHygrometer) Frequency() time.Duration {
	return f.frq
}

func (f *fakeHygrometer) Close() error {
	close(f.closed)
	return nil
}

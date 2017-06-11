package sensors

import (
	"log"
	"time"
)

const (
	fakeThermometerMin = 20
	fakeThermometerMax = 30
)

type fakeThermometer struct {
	frq    time.Duration
	closed chan struct{}
}

func NewFakeThermometer(frq time.Duration) Thermometer {
	fake := &fakeThermometer{
		frq:    frq,
		closed: make(chan struct{}),
	}
	return fake
}

func (f *fakeThermometer) nextValue() Temperature {
	temp := Temperature(theRand.Float64()*(fakeThermometerMax-fakeThermometerMin) + fakeThermometerMin)
	log.Printf("temp: %g", temp)
	return temp
}

func (f *fakeThermometer) Read() <-chan Temperature {
	results := make(chan Temperature)
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

func (f *fakeThermometer) Frequency() time.Duration {
	return f.frq
}

func (f *fakeThermometer) Close() error {
	close(f.closed)
	return nil
}

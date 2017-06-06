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
	results chan Temperature
}

func NewFakeThermometer(frq time.Duration) Thermometer {
	fake := &fakeThermometer{
		results: make(chan Temperature),
	}

	go fake.start(frq)

	return fake
}

func (f *fakeThermometer) start(frq time.Duration) {
	for {
		f.results <- f.nextTemp()
		<-time.After(frq)
	}
}

func (f *fakeThermometer) nextTemp() Temperature {
	temp := Temperature(theRand.Float64()*(fakeThermometerMax-fakeThermometerMin) + fakeThermometerMin)
	log.Printf("temp: %g", temp)
	return temp
}

func (f *fakeThermometer) Read() <-chan Temperature {
	return f.results
}

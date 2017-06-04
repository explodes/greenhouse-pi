package sensors

import "time"

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
	return Temperature(theRand.Float64()*(fakeThermometerMax-fakeThermometerMin) + fakeThermometerMin)
}

func (f *fakeThermometer) Read() <-chan Temperature {
	return f.results
}

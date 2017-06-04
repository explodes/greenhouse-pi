package sensors

import "testing"

func TestTemperature_Celsius(t *testing.T) {
	temp := Temperature(0)
	if 0 != temp.Celsius() {
		t.Fatal("unexpected temperature")
	}
}

func TestTemperature_Fahrenheit(t *testing.T) {
	cases := []struct {
		celsius    float64
		fahrenheit float64
	}{
		{celsius: 0, fahrenheit: 32},
	}

	for _, c := range cases {
		temp := Temperature(c.celsius)
		if c.fahrenheit != temp.Fahrenheit() {
			t.Errorf("Unexpected fahrenheit %g C -> %g F: %g", c.celsius, c.fahrenheit, temp.Fahrenheit())
		}
	}
}

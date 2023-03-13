package scd30

import "fmt"

// Measurement is a result of the measurement read from the device.
type Measurement struct {
	CO2         float32 // PPM [0 – 10000]
	Temperature float32 // °C  [-40 – 125]
	Humidity    float32 // %RH [0 – 100]
}

func (m *Measurement) String() string {
	return fmt.Sprintf(
		"CO2: %f ppm, temperature: %f °C, humidity: %f %%RH",
		m.CO2,
		m.Temperature,
		m.Humidity,
	)
}

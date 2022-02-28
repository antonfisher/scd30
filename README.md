# SCD30 (CO2 sensor)

[![GoDoc](https://godoc.org/github.com/antonfisher/scd30?status.svg)](https://godoc.org/github.com/antonfisher/scd30)
[![Go Report Card](https://goreportcard.com/badge/github.com/antonfisher/scd30)](https://goreportcard.com/report/github.com/antonfisher/scd30)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-green.svg)](https://conventionalcommits.org)

Go/TinyGo driver for Sensirion SCD30: ambient CO2, humidity, and temperature
sensor module.

<p align="center">
    <img alt="photo of SCD30 sensor and led matrix showing CO2 reading" src="https://raw.githubusercontent.com/antonfisher/scd30/docs/images/scd30-demo.jpg">
</p>

## Driver

Driver implements all commands described in the official
[interface description](https://sensirion.com/media/documents/D7CEEF4A/6165372F/Sensirion_CO2_Sensors_SCD30_Interface_Description.pdf)
document.

## Hardware details

- driver communicates over I2C interface (address is `0x61`)
- product page: [https://sensirion.com/products/catalog/SCD30/](https://sensirion.com/products/catalog/SCD30/)
- datasheet: [https://sensirion.com/media/documents/4EAF6AF8/61652C3C/Sensirion_CO2_Sensors_SCD30_Datasheet.pdf](https://sensirion.com/media/documents/4EAF6AF8/61652C3C/Sensirion_CO2_Sensors_SCD30_Datasheet.pdf)
- interface description: [https://sensirion.com/media/documents/D7CEEF4A/6165372F/Sensirion_CO2_Sensors_SCD30_Interface_Description.pdf](https://sensirion.com/media/documents/D7CEEF4A/6165372F/Sensirion_CO2_Sensors_SCD30_Interface_Description.pdf)
- buy sensor from SparkFun: [https://www.sparkfun.com/products/15112](https://www.sparkfun.com/products/15112)

## Example

Driver can work with any Go program that provides I2C interface like this:
```go
type I2C interface {
	Tx(addr uint16, w, r []byte) error
}
```

This is [TinyGo](https://github.com/tinygo-org/tinygo) example that uses
`machine` package's I2C to control SCD30:

```go
package main

import (
  "time"

  "machine"

  "github.com/antonfisher/scd30"
)

func main() {
  bus := machine.I2C0
  err := bus.Configure(machine.I2CConfig{})
  if err != nil {
    println("could not configure I2C:", err)
    return
  }

  // Create driver for SCD30
  co2sensor, err := scd30.New(bus)
  if err != nil {
    println("could not create driver:", err)
    return
  }

  // Read sensor's firmware version
  version, err := co2sensor.GetSoftwareVersion()
  if err != nil {
    println("failed to get software version:", err)
    return
  }
  println("software version:", version)

  // Start continuous measurement without provided ambient pressure
  // (should be ON by default on a new chip with 2 seconds interval)
  err = co2sensor.StartContinuousMeasurement(uint16(0))
  if err != nil {
    println("ERROR: co2 sensor: failed to trigger continuous measurement:", err)
    return
  }

  // Check is the sensor has data and read it every 2 seconds
  for {
    hasDataReady, err := co2sensor.HasDataReady()
    if err != nil {
      println("failed to check for data:", err)
    } else if hasDataReady {
      measurement, err := co2sensor.ReadMeasurement()
      if err != nil {
        println("failed to read measurement:", err)
      } else {
        //TODO can use println?
        fmt.Printf("measurement: %s\n\r", measurement.String())
      }
    } else {
      println("skipping, no data...")
    }

    time.Sleep(time.Second * 2)
  }
}
```

This example was tested on [nRF52840](https://www.adafruit.com/product/4062)
controller. Command to flash it with the:
```bash
tinygo flash -target=feather-nrf52840 main.go
```

Show the output in the terminal:
```
# find out dev to use
ls -l /dev/cu.*
# use `Control+A Control+\ y [Enter]` to exit
screen /dev/cu.usbmodem144101 19200
```

It prints out measurements read from the sensor:

![screenshot of the console with measurement results](https://raw.githubusercontent.com/antonfisher/scd30/docs/images/example-console-output.jpg)

For more configuration options see sensor's official interface description
document and
[driver reference](https://godoc.org/github.com/antonfisher/scd30).

## Inspired by

- Sensirion's official driver: [https://github.com/Sensirion/embedded-scd/](https://github.com/Sensirion/embedded-scd/)
- SparkFun SCD30 Arduino driver: [https://github.com/sparkfun/SparkFun_SCD30_Arduino_Library](https://github.com/sparkfun/SparkFun_SCD30_Arduino_Library)
- @pvainio's Go driver: [https://github.com/pvainio/scd30](https://github.com/pvainio/scd30)

## License

MIT License

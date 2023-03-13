// Package scd30 provides a Go/TinyGo driver for the Sensirion SCD30:
// ambient CO2, humidity, and temperature sensor module.
//
// Driver works over the I2C interface and supports all available commands
// listed in the official sensor's interface description document.
//
// Datasheet: https://sensirion.com/media/documents/4EAF6AF8/61652C3C/Sensirion_CO2_Sensors_SCD30_Datasheet.pdf
// Interface description: https://sensirion.com/media/documents/D7CEEF4A/6165372F/Sensirion_CO2_Sensors_SCD30_Interface_Description.pdf
// Buy sensor: https://www.sparkfun.com/products/15112
//
// This driver inspired by:
//   - Sensirion's official driver: https://github.com/Sensirion/embedded-scd/
//   - SparkFun SCD30 Arduino driver: https://github.com/sparkfun/SparkFun_SCD30_Arduino_Library
//   - @pvainio's Go driver: https://github.com/pvainio/scd30
package scd30

import (
	"fmt"
	"math"
	"time"
)

// Device implements TinyGo driver for Sensirion SCD30: ambient CO2, humidity,
// and temperature sensor module.
type Device struct {
	Address uint8

	// From the interface description document:
	// >> clock stretching period in write- and read-frames is 30ms, however, due
	// >> to internal calibration processes a maximal clock stretching of 150 ms
	// >> may occur once per day.
	ClockStretching time.Duration

	// Maximal I2C speed is 100 kHz and the master has to support clock
	// stretching. Sensirion recommends to operate the SCD30 at a baud rate
	// of 50 kHz or smaller.
	bus I2C
}

// readResponse writes IC2 command and reads result to the provided response
// byte array.
func (d *Device) readResponse(command uint16, response []byte) (
	err error,
) {
	err = d.bus.Tx(
		uint16(d.Address),
		[]byte{uint8(command >> 8), uint8(command & 0xFF)},
		[]byte{},
	)
	if err != nil {
		return fmt.Errorf("failed to send data: %w", err)
	}

	time.Sleep(d.ClockStretching)

	err = d.bus.Tx(uint16(d.Address), []byte{}, response)
	if err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}

	return nil
}

// readAndCheckResponse reads the register and checks CRC8 of the response
// assuming the last byte is CRC8 checksum.
func (d *Device) readAndCheckResponse(command uint16, response []byte) (
	err error,
) {
	if len(response) != 3 {
		return fmt.Errorf(
			"readAndCheckResponse only supports 3 bytes long responses (uint16+CRC8)",
		)
	}

	err = d.readResponse(command, response)
	if err != nil {
		return err
	}

	return checkCRC8(response)
}

// writeValue writes setting value for the provided command.
func (d *Device) writeValue(command, value uint16) (err error) {
	err = d.bus.Tx(
		uint16(d.Address),
		[]byte{
			uint8(command >> 8),
			uint8(command & 0xFF),
			uint8(value >> 8),
			uint8(value & 0xFF),
			uint8(computeCRC8([]byte{uint8(value >> 8), uint8(value & 0xFF)}, 2)),
		},
		[]byte{},
	)
	if err != nil {
		return fmt.Errorf("failed to send data: %w", err)
	}

	return nil
}

// SoftReset resets the sensor.
func (d *Device) SoftReset() error {
	return d.writeValue(CMD_SOFT_RESET, RESET)
}

// GetSoftwareVersion reads software version of the connected device as
// a "major.minor" string.
func (d *Device) GetSoftwareVersion() (version string, err error) {
	result := make([]byte, 3)

	err = d.readAndCheckResponse(CMD_READ_VERSION, result)
	if err != nil {
		return version, err
	}

	return fmt.Sprintf("%d.%d", result[0], result[1]), nil
}

// GetMeasurementInterval returns current measurement interval [2-1800]s.
func (d *Device) GetMeasurementInterval() (interval uint16, err error) {
	result := make([]byte, 3)

	err = d.readAndCheckResponse(CMD_MEASUREMENT_INTERVAL, result)
	if err != nil {
		return interval, err
	}

	interval = uint16(result[0])<<8 | uint16(result[1])

	return interval, nil
}

// SetMeasurementInterval sets measurement interval, must be [2-1800]s.
func (d *Device) SetMeasurementInterval(interval uint16) error {
	if interval < 2 || interval > 1800 {
		return fmt.Errorf(
			"invalid measurement interval: %d, expected [2-1800]s",
			interval,
		)
	}

	return d.writeValue(CMD_MEASUREMENT_INTERVAL, interval)
}

// GetSelfCalibration returns enabled/disabled state of automatic
// self-calibration feature (ASC).
func (d *Device) GetSelfCalibration() (enabled bool, err error) {
	result := make([]byte, 3)

	err = d.readAndCheckResponse(CMD_SELF_CALIBRATION, result)
	if err != nil {
		return enabled, err
	}

	return uint16(result[1]) == SELF_CALIBRATION_ENABLED, nil
}

// SetSelfCalibration enables/disables automatic self-calibration feature (ASC).
//
// From the docs:
// >> To work properly SCD30 has to see fresh air on a regular basis. Optimal
// >> working conditions are given when the sensor sees fresh air for one hour
// >> every day so that ASC can constantly re-calibrate. ASC only works in
// >> continuous measurement mode.
func (d *Device) SetSelfCalibration(enabled bool) error {
	if enabled {
		return d.writeValue(CMD_SELF_CALIBRATION, SELF_CALIBRATION_ENABLED)
	}

	return d.writeValue(CMD_SELF_CALIBRATION, SELF_CALIBRATION_DISABLED)
}

// GetForcedRecalibrationValue returns value if set [400-2000ppm].
func (d *Device) GetForcedRecalibrationValue() (value uint16, err error) {
	result := make([]byte, 3)

	err = d.readAndCheckResponse(CMD_FORCED_RECALIBRATION_VALUE, result)
	if err != nil {
		return value, err
	}

	value = uint16(result[0])<<8 | uint16(result[1])

	return value, nil
}

// SetForcedRecalibrationValue sets forced recalibration value that overwrites
// value from automatic self-calibration. Must be [400-2000]ppm.
func (d *Device) SetForcedRecalibrationValue(value uint16) error {
	if value < 400 || value > 2000 {
		return fmt.Errorf(
			"invalid forced recalibration value: %d, expected [400-2000]",
			value,
		)
	}

	return d.writeValue(CMD_FORCED_RECALIBRATION_VALUE, value)
}

// GetTemperatureOffset returns value of the configured temperature offset in
// 1/100 °C.
func (d *Device) GetTemperatureOffset() (offset uint16, err error) {
	result := make([]byte, 3)

	err = d.readAndCheckResponse(CMD_TEMPERATURE_OFFSET, result)
	if err != nil {
		return offset, err
	}

	offset = uint16(result[0])<<8 | uint16(result[1])

	return offset, nil
}

// SetTemperatureOffset sets temperature offset in 1/100 °C.
func (d *Device) SetTemperatureOffset(offset uint16) error {
	return d.writeValue(CMD_TEMPERATURE_OFFSET, offset)
}

// GetAltitudeCompensation returns value of the configured altitude
// compensation in meters above the sea level.
func (d *Device) GetAltitudeCompensation() (altitude uint16, err error) {
	result := make([]byte, 3)

	err = d.readAndCheckResponse(CMD_ALTITUDE_COMPENSATION, result)
	if err != nil {
		return altitude, err
	}

	altitude = uint16(result[0])<<8 | uint16(result[1])

	return altitude, nil
}

// SetAltitudeCompensation sets altitude compensation in meters above the sea
// level.
func (d *Device) SetAltitudeCompensation(altitude uint16) error {
	return d.writeValue(CMD_ALTITUDE_COMPENSATION, altitude)
}

// StartContinuousMeasurement enables continuous measurement with provided
// ambient pressure: [700-1200]mBar or 0 (disabled).
func (d *Device) StartContinuousMeasurement(ambientPressure uint16) error {
	if ambientPressure != 0 && (ambientPressure < 700 || ambientPressure > 1200) {
		return fmt.Errorf(
			"invalid ambient pressure value: %d, expected 0 or [700-1200]",
			ambientPressure,
		)
	}

	return d.writeValue(CMD_START_CONTINUOUS_MEASUREMENT, ambientPressure)
}

// StopContinuousMeasurement stops continuous measurement.
func (d *Device) StopContinuousMeasurement() error {
	return d.writeValue(
		CMD_STOP_CONTINUOUS_MEASUREMENT,
		uint16(STOP_CONTINUOUS_MEASUREMENT),
	)
}

// HasDataReady returns true if there are measurements ready to be read.
func (d *Device) HasDataReady() (isReady bool, err error) {
	result := make([]byte, 3)

	err = d.readAndCheckResponse(CMD_READ_DATA_READY, result)
	if err != nil {
		return false, err
	}

	return uint16(result[1]) == HAS_DATA_READY, nil
}

// ReadMeasurement reads measurements from the device.
func (d *Device) ReadMeasurement() (measurement Measurement, err error) {
	result := make([]byte, 18)

	err = d.readResponse(CMD_READ_MEASUREMENT, result)
	if err != nil {
		return measurement, err
	}

	// 18 bytes response has three measurement results for CO2, temperature, and
	// humidity, each has: 4 bytes of the value + 2 CRC8's for every bytes couple.
	// Values come in BigEndian notation.
	//
	//                      CRC8
	//       + - - + - - - + - - + - - - + - - +
	//       |     |       |     |       |     |
	//       v     v       v     v       v     v
	// +-------------+-------------+-------------+
	// | X X C X X C | X X C X X C | X X C X X C |
	// +-------------+-------------+-------------+
	// |             |             |             |
	// |     CO2     | Temperature |  Humidity   |
	//
	i := 0
	chunk := make([]byte, 3)     // 2 data bytes + 1 CRC8
	var value uint32             // 4 bytes of a single measurement
	values := make([]float32, 3) // all 3 measurements as floats
	for v := 0; v < len(values); v++ {
		value = 0
		for c := 0; c < 2; c++ { // 2 chunks per one value
			chunk[0], chunk[1], chunk[2] = result[i], result[i+1], result[i+2]
			err = checkCRC8(chunk)
			if err != nil {
				return measurement, err
			}
			value <<= 8
			value |= uint32(result[i])
			value <<= 8
			value |= uint32(result[i+1])
			i += 3
		}
		values[v] = math.Float32frombits(value)
	}

	measurement.CO2 = values[0]
	measurement.Temperature = values[1]
	measurement.Humidity = values[2]

	return measurement, nil
}

// New create a new Sensirion SCD30 driver.
func New(bus I2C) (d Device) {
	d = Device{
		Address: I2C_ADDRESS,
		// Manual testing shows that 150ms is probably the most stable default.
		ClockStretching: 150 * time.Millisecond,
		bus:             bus,
	}

	return d
}

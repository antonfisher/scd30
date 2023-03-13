package scd30

// Registers. Names taken from the datasheet:
// https://sensirion.com/media/documents/D7CEEF4A/6165372F/Sensirion_CO2_Sensors_SCD30_Interface_Description.pdf
const (
	I2C_ADDRESS uint8 = 0x61

	// Read commands.
	CMD_READ_VERSION     uint16 = 0xD100
	CMD_READ_DATA_READY  uint16 = 0x0202
	CMD_READ_MEASUREMENT uint16 = 0x0300

	// Write commands.
	CMD_SOFT_RESET                   uint16 = 0xD304
	CMD_START_CONTINUOUS_MEASUREMENT uint16 = 0x0010
	CMD_STOP_CONTINUOUS_MEASUREMENT  uint16 = 0x0104

	// Read/write commands.
	CMD_MEASUREMENT_INTERVAL       uint16 = 0x4600
	CMD_ALTITUDE_COMPENSATION      uint16 = 0x5102
	CMD_FORCED_RECALIBRATION_VALUE uint16 = 0x5204
	CMD_SELF_CALIBRATION           uint16 = 0x5306
	CMD_TEMPERATURE_OFFSET         uint16 = 0x5403

	// Responses.
	HAS_DATA_READY            uint16 = 0x0001
	SELF_CALIBRATION_ENABLED  uint16 = 0x0001
	SELF_CALIBRATION_DISABLED uint16 = 0x0000

	// Command values.
	STOP_CONTINUOUS_MEASUREMENT uint16 = 0x0001
	RESET                       uint16 = 0x0001
)

package scd30

import "fmt"

// checkCRC8 checks CRC8 of a byte array assuming the last item is a checksum
func checkCRC8(data []byte) error {
	l := uint8(len(data))
	expectedCRC := uint8(data[l-1])
	crc := computeCRC8(data, l-1)

	if crc != expectedCRC {
		return fmt.Errorf(
			"CRC checksum failed, expected: '%x', got: '%x'",
			expectedCRC,
			crc,
		)
	}

	return nil
}

// computeCRC8 computes CRC8 of the l first bytes of the given byte array
func computeCRC8(data []byte, l uint8) uint8 {
	crc := uint8(0xFF)

	for x := uint8(0); x < l; x++ {
		crc ^= data[x]
		for i := 0; i < 8; i++ {
			if (crc & 0x80) != 0 {
				crc = uint8((crc << 1) ^ 0x31)
			} else {
				crc <<= 1
			}
		}
	}

	return crc
}

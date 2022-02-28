package scd30

// I2C interface that is compatable with TinyGo drivers I2C interface
type I2C interface {
	Tx(addr uint16, w, r []byte) error
}

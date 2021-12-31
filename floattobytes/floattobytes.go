package floattobytes

import (
	"encoding/binary"
	"math"
)

/*
	Quietly copied from https://newbedev.com/convert-byte-slice-uint8-to-float64-in-golang
*/

func Float64_from_bytes(bytes []byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes)
	float := math.Float64frombits(bits)
	return float
}

func Float64_to_bytes(float float64) []byte {
	bits := math.Float64bits(float)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)
	return bytes
}

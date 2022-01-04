package bytes

import (
	"encoding/binary"
	"errors"
	"math"
)

/*
	Quietly copied from https://newbedev.com/convert-byte-slice-uint8-to-float64-in-golang
*/

func Float64_from_bytes(bytes [8]byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes[:])
	float := math.Float64frombits(bits)
	return float
}

func Float64_to_bytes(float float64) [8]byte {
	var bytes [8]byte

	bits := math.Float64bits(float)
	binary.LittleEndian.PutUint64(bytes[:], bits)

	return bytes
}

func Float64Slice_from_bytesSlice(bytes []byte) []float64 {
	if len(bytes)%8 != 0 {
		panic(errors.New("Float64Slice_from_bytesSlice: bytes slice must be a multiple of 8"))
	}
	nFloats := len(bytes) / 8
	r := make([]float64, 0, nFloats)

	for i := 0; i < nFloats; i++ {
		var f [8]byte
		b := bytes[i*8 : (i+1)*8]
		for i, sb := range b {
			f[i] = sb
		}
		r = append(r, Float64_from_bytes(f))
	}
	return r
}

func Position_from_bytesarray(bytes [2][8]byte) [2]float64 {
	var r [2]float64

	for i, b := range bytes {
		r[i] = Float64_from_bytes(b)
	}

	return r
}

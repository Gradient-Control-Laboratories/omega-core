package float16

import (
	"encoding/base64"
	"encoding/binary"
	"math"
)

// EncodeSliceBase64 packs a []float32 as little-endian IEEE 754 binary16
// (half-precision) bytes and base64-encodes the result. Matches WebGPU's
// r16float texture upload format at 2 bytes per texel.
func EncodeSliceBase64(v []float32) string {
	raw := make([]byte, len(v)*2)
	for i, f := range v {
		binary.LittleEndian.PutUint16(raw[i*2:], FromFloat32(f))
	}
	return base64.StdEncoding.EncodeToString(raw)
}

// FromFloat32 converts an IEEE 754 binary32 to the bit pattern of an IEEE 754
// binary16 (half-precision). Handles normals, subnormals, zero, infinity, and
// NaN.
func FromFloat32(f float32) uint16 {
	bits := math.Float32bits(f)
	sign := uint16((bits >> 31) & 0x1)
	exp32 := int32((bits >> 23) & 0xFF)
	mant32 := bits & 0x7FFFFF

	// Special cases first.
	if exp32 == 0xFF { // Inf or NaN
		if mant32 == 0 {
			return (sign << 15) | (0x1F << 10) // ±Inf
		}
		// NaN: preserve a non-zero mantissa so it remains NaN after truncation.
		mant16 := uint16(mant32 >> 13)
		if mant16 == 0 {
			mant16 = 1
		}
		return (sign << 15) | (0x1F << 10) | mant16
	}
	if exp32 == 0 && mant32 == 0 {
		return sign << 15 // ±0
	}

	// Re-bias exponent into f16 range. f16 bias = 15, f32 bias = 127.
	exp16 := exp32 - 127 + 15

	if exp16 >= 0x1F {
		// Overflow → ±Inf
		return (sign << 15) | (0x1F << 10)
	}
	if exp16 <= 0 {
		// Subnormal or underflow.
		if exp16 < -10 {
			return sign << 15 // too small even as a subnormal → ±0
		}
		// Add the implicit leading 1 from the f32 mantissa, then shift right.
		mant := mant32 | 0x800000
		shift := uint32(14 - exp16) // amount to shift the f32 mantissa right
		half := uint16(mant >> shift)
		// Round on the bit just below the new LSB.
		if (mant>>(shift-1))&1 == 1 {
			half++
		}
		return (sign << 15) | half
	}

	// Normal case: pack 5-bit exponent and top-10 bits of the mantissa,
	// rounding on the truncated bit just below.
	half := (uint16(exp16) << 10) | uint16(mant32>>13)
	if mant32&0x1000 != 0 {
		half++ // may carry into the exponent — that's fine, encodes correctly
	}
	return (sign << 15) | half
}

// ToFloat32 converts an IEEE 754 binary16 bit pattern to a float32.
func ToFloat32(h uint16) float32 {
	sign := uint32(h>>15) << 31
	exp := uint32(h>>10) & 0x1f
	mant := uint32(h) & 0x3ff
	switch exp {
	case 0:
		if mant == 0 {
			return math.Float32frombits(sign) // +/-0
		}
		// Subnormal: shift until the implicit bit is set, then emit it as a
		// float32 normal. Each shift halves the exponent, and a half subnormal
		// is 2^-14 * mant/1024, so the biased float32 exponent is 113-shifts.
		shifts := uint32(0)
		for mant&0x400 == 0 {
			mant <<= 1
			shifts++
		}
		mant &= 0x3ff
		return math.Float32frombits(sign | (113-shifts)<<23 | mant<<13)
	case 0x1f:
		return math.Float32frombits(sign | 0xff<<23 | mant<<13) // Inf / NaN
	default:
		return math.Float32frombits(sign | (exp+127-15)<<23 | mant<<13)
	}
}

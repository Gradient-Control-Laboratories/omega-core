package float16

import (
	"encoding/base64"
	"encoding/binary"
	"math"
	"testing"
)

func TestToFloat32(t *testing.T) {
	cases := []struct {
		bits uint16
		want float32
		what string
	}{
		{0x0000, 0, "zero"},
		{0x8000, float32(math.Copysign(0, -1)), "negative zero"},
		{0x3C00, 1, "one"},
		{0xC000, -2, "minus two"},
		{0x3555, 0.333251953125, "a third, rounded to half"},
		{0x7BFF, 65504, "largest finite half"},
		{0x0400, 6.103515625e-05, "smallest normal (2^-14)"},
		{0x0200, 3.0517578125e-05, "subnormal 2^-15"},
		{0x0001, 5.960464477539063e-08, "smallest subnormal (2^-24)"},
		{0x7C00, float32(math.Inf(1)), "+Inf"},
		{0xFC00, float32(math.Inf(-1)), "-Inf"},
	}
	for _, c := range cases {
		got := ToFloat32(c.bits)
		if got != c.want || math.Signbit(float64(got)) != math.Signbit(float64(c.want)) {
			t.Errorf("ToFloat32(%#04x) [%s] = %v, want %v", c.bits, c.what, got, c.want)
		}
	}

	if !math.IsNaN(float64(ToFloat32(0x7E00))) {
		t.Errorf("ToFloat32(0x7E00) expected NaN")
	}
}

func TestFromFloat32(t *testing.T) {
	cases := []struct {
		in   float32
		want uint16
		what string
	}{
		{0, 0x0000, "zero"},
		{float32(math.Copysign(0, -1)), 0x8000, "negative zero"},
		{1, 0x3C00, "one"},
		{-2, 0xC000, "minus two"},
		{65504, 0x7BFF, "largest finite half"},
		{1e5, 0x7C00, "overflow to +Inf"},
		{-1e5, 0xFC00, "overflow to -Inf"},
		{float32(math.Inf(1)), 0x7C00, "+Inf"},
		{float32(math.Inf(-1)), 0xFC00, "-Inf"},
		{6.103515625e-05, 0x0400, "smallest normal (2^-14)"},
		{3.0517578125e-05, 0x0200, "subnormal (2^-15)"},
		{3.0517578125e-05 + 4.5e-08, 0x0201, "subnormal rounded up"},
		{1.0 + 0.0005, 0x3C01, "normal rounded up"},
		{5.960464477539063e-08, 0x0001, "smallest subnormal (2^-24)"},
		{1e-10, 0x0000, "underflow to +0"},
		{-1e-10, 0x8000, "underflow to -0"},
	}
	for _, c := range cases {
		if got := FromFloat32(c.in); got != c.want {
			t.Errorf("FromFloat32(%v) [%s] = %#04x, want %#04x", c.in, c.what, got, c.want)
		}
	}

	// NaN with small payload in low mantissa bits must still remain NaN (non-zero mant16).
	smallPayloadNaN := math.Float32frombits(0x7F800001)
	gotNaN := FromFloat32(smallPayloadNaN)
	if (gotNaN>>10)&0x1F != 0x1F || (gotNaN&0x3FF) == 0 {
		t.Errorf("FromFloat32(smallPayloadNaN) = %#04x, expected NaN bit pattern", gotNaN)
	}
}

func TestEncodeSliceBase64RoundTrip(t *testing.T) {
	in := []float32{0, -0.5, 1.25, -3.75, 65504}
	b64 := EncodeSliceBase64(in)
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		t.Fatalf("decode base64: %v", err)
	}
	if len(raw) != len(in)*2 {
		t.Fatalf("byte length = %d, want %d", len(raw), len(in)*2)
	}
	for i, want := range in {
		bits := binary.LittleEndian.Uint16(raw[i*2 : i*2+2])
		if got := ToFloat32(bits); got != want {
			t.Errorf("index %d: got %v, want %v", i, got, want)
		}
	}
}

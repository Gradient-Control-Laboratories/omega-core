package fidget_test

import (
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/backend/tape"
	"gradientcontrol.com/omega/core/internal/fidget"
)

// A material is four rows of four, and the tape keeps all sixteen trees of it,
// in order. It was once dropped without a word, because the serializer copied
// only a record of twelve.
func TestSerializeTapeWithAppearanceKeepsAFourRowMaterial(t *testing.T) {
	fidget.Lock()
	defer fidget.Unlock()
	fidget.Clear()

	field := fidget.Sub(fidget.NewX(), fidget.NewConst(1))
	material := make([]fidget.TreeIndex, 16)
	for i := range material {
		material[i] = fidget.NewConst(float64(i) + 0.5)
	}
	colour := []fidget.TreeIndex{fidget.NewConst(0.1), fidget.NewConst(0.2), fidget.NewConst(0.3)}

	bytes, err := fidget.SerializeTapeWithAppearance(
		[]fidget.TreeIndex{field, field},
		[]fidget.Appearance{{Color: colour, Material: material, Displaced: true}, {}},
	)
	if err != nil {
		t.Fatalf("SerializeTapeWithAppearance: %v", err)
	}
	var tp tape.Tape
	if err := proto.Unmarshal(bytes, &tp); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if len(tp.Appearance) != 2 {
		t.Fatalf("%d appearances, want 2", len(tp.Appearance))
	}

	a := tp.Appearance[0]
	if len(a.Color) != 3 {
		t.Errorf("colour has %d channels, want 3", len(a.Color))
	}
	if len(a.Material) != 16 {
		t.Fatalf("material has %d entries, want 16", len(a.Material))
	}
	for i, id := range a.Material {
		if got, want := tp.Nodes[id].GetConstant(), float64(i)+0.5; got != want {
			t.Errorf("material entry %d is the constant %v, want %v", i, got, want)
		}
	}
	if !a.Displaced {
		t.Error("the displaced flag was lost")
	}

	b := tp.Appearance[1]
	if len(b.Color) != 0 || len(b.Material) != 0 || b.Displaced {
		t.Errorf("an output with no appearance got colour %d, material %d, displaced %v", len(b.Color), len(b.Material), b.Displaced)
	}
}

// A colour or material of any other length is an error, not a quietly absent
// one: dropping it is how a material once vanished from every tape.
func TestSerializeTapeWithAppearanceRefusesOtherLengths(t *testing.T) {
	// The lengths are checked before any call into the library, so these
	// indices need not name trees.
	one := []fidget.TreeIndex{0}
	for _, n := range []int{1, 4, 12, 15, 17} {
		_, err := fidget.SerializeTapeWithAppearance(one, []fidget.Appearance{{Material: make([]fidget.TreeIndex, n)}})
		if err == nil || !strings.Contains(err.Error(), "material has") {
			t.Errorf("a material of %d trees: err = %v, want one naming the material", n, err)
		}
	}
	for _, n := range []int{1, 2, 4} {
		_, err := fidget.SerializeTapeWithAppearance(one, []fidget.Appearance{{Color: make([]fidget.TreeIndex, n)}})
		if err == nil || !strings.Contains(err.Error(), "colour has") {
			t.Errorf("a colour of %d trees: err = %v, want one naming the colour", n, err)
		}
	}
}

package main

import (
	"reflect"
	"testing"
)

// Every material in the manifest carries the colours it paints, and no finish
// does: the Bodies panel shows a finish over the body's own colour, so a colour
// recorded for one would be drawn where it does not belong.
func TestEveryMaterialHasItsColours(t *testing.T) {
	m, _ := readManifest(t)
	for _, e := range m.Entries {
		colors, ok := m.Colors[e.Package+"."+e.Symbol]
		switch {
		case e.Kind == "material" && (!ok || len(colors) == 0):
			t.Errorf("%s has no colours", e.Call())
		case e.Kind == "finish" && ok:
			t.Errorf("finish %s has colours %v", e.Call(), colors)
		}
	}
}

// The committed colours are the ones the libraries paint now. A spread: a
// metal, a liquid and a food, whose colours have nothing in common.
func TestCommittedColoursAreCurrent(t *testing.T) {
	m, _ := readManifest(t)
	for _, want := range []string{"metal.brass", "physical.coffee", "physical.ketchup"} {
		var entry Entry
		for _, e := range m.Entries {
			if e.Package+"."+e.Symbol == want {
				entry = e
			}
		}
		if entry.Symbol == "" {
			t.Fatalf("%s is not in the manifest", want)
		}
		fresh, err := surfaceColors(entry)
		if err != nil {
			t.Fatal(err)
		}
		if got := m.Colors[want]; !reflect.DeepEqual(got, fresh) {
			t.Errorf("%s: manifest has %v, the library paints %v (run `go run ./tool/materialswatch`)", want, got, fresh)
		}
	}
}

// A uniform colour is one colour, its mean; a varying one is three, darkest to
// lightest, so the chip can draw the range rather than one sample of it.
func TestSpanColours(t *testing.T) {
	uniform := [][3]float64{{0.5, 0.25, 1}, {0.5, 0.25, 1}, {0.5, 0.2506, 1}}
	if got, want := spanColors(uniform), [][3]float64{{0.5, 0.25, 1}}; !reflect.DeepEqual(got, want) {
		t.Errorf("uniform: %v, want %v", got, want)
	}

	var ramp [][3]float64
	for i := 10; i >= 0; i-- {
		v := float64(i) / 10
		ramp = append(ramp, [3]float64{v, v, v})
	}
	got := spanColors(ramp)
	want := [][3]float64{{0.1, 0.1, 0.1}, {0.5, 0.5, 0.5}, {0.9, 0.9, 0.9}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ramp: %v, want %v", got, want)
	}
}

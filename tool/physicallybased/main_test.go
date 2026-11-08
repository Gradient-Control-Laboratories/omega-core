package main

import (
	"os"
	"testing"
)

// TestGeneratedPackageIsCurrent regenerates the package from the committed
// snapshot and requires it to match the committed physical.gsl, so the
// package is never edited by hand or left stale after a snapshot refresh.
func TestGeneratedPackageIsCurrent(t *testing.T) {
	b, err := os.ReadFile("materials.json")
	if err != nil {
		t.Fatal(err)
	}
	s, dropped, err := parseSnapshot(b)
	if err != nil {
		t.Fatal(err)
	}
	got, err := generate(s, dropped)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("../../stdlib/material/bulk/physical/physical.gsl")
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Error("physical.gsl is out of date: run `go run .` in tool/physicallybased")
	}
}

func TestIdentifier(t *testing.T) {
	for in, want := range map[string]string{
		"Juice (Ruby Grapefruit)": "juiceRubyGrapefruit",
		"MIT Black":               "mitBlack",
		"Vodka (40%)":             "vodka40",
		"Plastic (PLA)":           "plasticPLA",
		"Glass (Soda-lime)":       "glassSodaLime",
		"Skin III":                "skinIII",
		"Aluminum":                "aluminum",
	} {
		if got := identifier(in); got != want {
			t.Errorf("identifier(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestTypicalDensity(t *testing.T) {
	for _, c := range []struct {
		in   []float64
		want float64
		ok   bool
	}{
		{[]float64{2700}, 2700, true},
		{[]float64{7500, 8000}, 7750, true},
		{[]float64{1380, 1450, 1400}, 1400, true},
		{nil, 0, false},
	} {
		got, ok := typicalDensity(c.in)
		if got != c.want || ok != c.ok {
			t.Errorf("typicalDensity(%v) = %v, %v; want %v, %v", c.in, got, ok, c.want, c.ok)
		}
	}
}

func TestEncodeSRGB(t *testing.T) {
	for _, c := range []struct{ in, want float64 }{
		{0, 0}, {1, 1}, {1.059, 1}, {-0.1, 0}, {0.5, 0.7354},
	} {
		if got := encodeSRGB(c.in); num(got) != num(c.want) {
			t.Errorf("encodeSRGB(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

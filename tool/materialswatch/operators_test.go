package main

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	assembler "gradientcontrol.com/omega/core/internal/assembler"
	"gradientcontrol.com/omega/core/internal/source"
)

// The operators are read off the @surface presets: the overload of each
// operator that takes only slots, its arity off its return list, and its
// adjust record from the operator call its body makes.

func readOperators(t *testing.T) map[string]*Operator {
	t.Helper()
	ops, err := ReadOperators()
	if err != nil {
		t.Fatal(err)
	}
	return ops
}

// The 66 presets the design lists (doc/material_composition.md, "Design:
// operators as presets"), one per operator, by the operator's own name.
func TestOperatorsAreThePresets(t *testing.T) {
	want := []string{
		"machined.turned", "machined.faced", "machined.faceMilled", "machined.ground", "machined.brushed", "machined.blasted", "machined.knurled",
		"machined.honed", "machined.engineTurned", "machined.ballMilled",
		"machined.scraped", "machined.lapped", "machined.sparkEroded", "machined.tumbled", "machined.polished",
		"coating.anodized", "coating.hardAnodized", "coating.blackOxide", "coating.phosphate", "coating.heatTint", "coating.anodizedTitanium",
		"coating.chippedPaint", "coating.clearCoat",
		"coating.powderCoat", "coating.wrinkleCoat", "coating.sandCoat", "coating.hammertone", "coating.galvanized",
		"plated.zinc", "plated.zincYellow", "plated.zincBlack", "plated.nickel", "plated.chrome", "plated.tin", "plated.cadmium", "plated.goldFlash", "plated.silver",
		"sheet.hotRolled", "sheet.coldRolled", "sheet.pickledOiled", "sheet.checkerPlate", "sheet.stuccoEmbossed",
		"cast.sandCast", "cast.dieCast", "cast.investmentCast", "cast.forged", "cast.shotPeened",
		"molded.textured", "molded.leatherGrain", "molded.grip",
		"printed.laserMelted", "printed.sintered",
		"ceramic.celadon", "ceramic.crackle", "ceramic.speckled", "ceramic.matteGlaze",
		"classic.scanlines",
		"weathering.weathered", "weathering.weatheredRunoff",
		"playful.glitter", "playful.glitterTwinkle",
		"material.withRoughness", "material.withMetallic", "material.withIor", "material.withEmissive", "material.withTransmission", "material.withLay",
	}
	sort.Strings(want)
	got := []string{}
	for name := range readOperators(t) {
		got = append(got, name)
	}
	sort.Strings(got)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("operators:\n got %v\nwant %v", got, want)
	}
}

// turned's preset takes all three slots, returns four values, and its record
// lifts the feed and nose radius from the frame form it calls, binding the
// layer's four values and assigning the first two.
func TestOperatorWithEverySlot(t *testing.T) {
	op := readOperators(t)["machined.turned"]
	if op == nil {
		t.Fatal("machined.turned has no operator")
	}
	if op.Call != "machined.turned" || op.Arity != 4 || op.Slots != (Slots{Material: 0, Color: 1, Frame: 2}) {
		t.Errorf("call %q arity %d slots %+v", op.Call, op.Arity, op.Slots)
	}
	if op.Adjust == nil {
		t.Fatal("no adjust record")
	}
	if want := []string{"machined.turned"}; !reflect.DeepEqual(op.Adjust.Calls, want) {
		t.Errorf("calls %v, want %v", op.Adjust.Calls, want)
	}
	want := []Parameter{
		{Name: "feed", Kind: "range", Min: f(0.05), Max: f(0.3), Default: []float64{0.15}, Call: "machined.turned", Arg: 3},
		{Name: "noseRadius", Kind: "range", Min: f(0.2), Max: f(1.6), Default: []float64{0.8}, Call: "machined.turned", Arg: 4},
	}
	if !reflect.DeepEqual(op.Adjust.Parameters, want) {
		t.Errorf("parameters:\n got %s\nwant %s", describe(op.Adjust.Parameters), describe(want))
	}
	wantTemplate := strings.Join([]string{
		"${layer}, ${layer}Color, ${layer}Height, ${layer}Slope = machined.turned(${material}, ${color}, ${frame}, ${feed}, ${noseRadius});",
		"${body}.material = ${layer};",
		"${body}.color = ${layer}Color;",
	}, "\n")
	if op.Adjust.Template != wantTemplate {
		t.Errorf("template:\n%s\nwant:\n%s", op.Adjust.Template, wantTemplate)
	}
}

// A coat replaces the body's colour, so its preset has one slot; a modifier
// returns the material alone, so its layer binds and assigns one value; and a
// preset that bakes only a plain argument has nothing to adjust.
func TestOperatorSlotsFollowThePreset(t *testing.T) {
	ops := readOperators(t)
	if op := ops["coating.powderCoat"]; op == nil || op.Arity != 4 || op.Slots != (Slots{0, -1, -1}) || op.Adjust == nil || op.Adjust.Parameters[0].Name != "coatColor" {
		t.Errorf("coating.powderCoat: %+v", op)
	}
	rough := ops["material.withRoughness"]
	if rough == nil || rough.Arity != 1 || rough.Slots != (Slots{0, -1, -1}) || rough.Adjust == nil {
		t.Fatalf("material.withRoughness: %+v", rough)
	}
	wantTemplate := "${layer} = material.withRoughness(${material}, ${rough});\n${body}.material = ${layer};"
	if rough.Adjust.Template != wantTemplate {
		t.Errorf("withRoughness template:\n%s\nwant:\n%s", rough.Adjust.Template, wantTemplate)
	}
	if op := ops["material.withIor"]; op == nil || op.Arity != 1 || op.Adjust != nil {
		t.Errorf("material.withIor: %+v, want arity 1 and nothing to adjust", op)
	}
}

// The knurl's teeth, the floor plate's lugs and the stucco's pebbles are
// the heights a layer has to apply; every other operator's relief rides in
// its material and returns 0.
func TestOnlyTheKnurlReturnsRelief(t *testing.T) {
	heights := map[string]bool{"machined.knurled": true, "sheet.checkerPlate": true, "sheet.stuccoEmbossed": true, "molded.grip": true}
	for name, op := range readOperators(t) {
		if want := heights[name]; op.Relief != want {
			t.Errorf("%s: relief %v, want %v", name, op.Relief, want)
		}
	}
}

// Every preset is proven the way the panel will use it: applied to steel on a
// sphere, with the sphere's frame for a frame slot, destructured to its arity.
func TestEveryOperatorCompiles(t *testing.T) {
	for name, op := range readOperators(t) {
		doc := ProbeDoc(op)
		if _, err := assembler.Assemble([]string{doc}, &source.MockManager{}); err != nil {
			t.Errorf("%s: %v\n%s", name, err, doc)
		}
	}
}

// The committed manifest carries the operators the libraries declare.
func TestCommittedOperatorsAreCurrent(t *testing.T) {
	m, _ := readManifest(t)
	if ops := readOperators(t); !reflect.DeepEqual(ops, m.Operators) {
		t.Errorf("the manifest's operators differ from the libraries' — run `go run ./tool/materialswatch -manifest-only`")
	}
}

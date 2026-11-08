package main

import (
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// What the panel can adjust is read off the constructors' own parameter
// annotations, so these check the reading against presets whose shape is
// known: a destructured call, calls returned directly, and a preset whose
// constructor declares nothing adjustable.

func f(v float64) *float64 { return &v }

func readAdjust(t *testing.T) map[string]*Adjustable {
	t.Helper()
	adjust, err := ReadAdjust()
	if err != nil {
		t.Fatal(err)
	}
	return adjust
}

// redPowderCoat destructures powderCoat's four returns and returns two of
// them. Its colour, gloss and peel are the constructor's annotated parameters,
// in call order, with the preset's own arguments as defaults.
func TestAdjustLiftsADestructuredCall(t *testing.T) {
	adj := readAdjust(t)["coating.redPowderCoat"]
	if adj == nil {
		t.Fatal("coating.redPowderCoat has no adjust record")
	}
	if want := []string{"coating.powderCoat"}; !reflect.DeepEqual(adj.Calls, want) {
		t.Errorf("calls %v, want %v", adj.Calls, want)
	}
	want := []Parameter{
		{Name: "coatColor", Kind: "color", Default: []float64{0.62, 0.06, 0.05}, Call: "coating.powderCoat", Arg: 0},
		{Name: "glossUnits", Kind: "range", Min: f(0), Max: f(100), Default: []float64{85}, Call: "coating.powderCoat", Arg: 1},
		{Name: "peel", Kind: "range", Min: f(0), Max: f(1), Default: []float64{0.7}, Call: "coating.powderCoat", Arg: 2},
	}
	if !reflect.DeepEqual(adj.Parameters, want) {
		t.Errorf("parameters:\n got %s\nwant %s", describe(adj.Parameters), describe(want))
	}
	wantTemplate := strings.Join([]string{
		"${body}Surface, ${body}SurfaceColor, ${body}Height, ${body}Slope = coating.powderCoat(${coatColor}, ${glossUnits}, ${peel});",
		"${body}.material = ${body}Surface;",
		"${body}.color = ${body}SurfaceColor;",
	}, "\n")
	if adj.Template != wantTemplate {
		t.Errorf("template:\n%s\nwant:\n%s", adj.Template, wantTemplate)
	}
}

// greyGranite returns its two calls directly, and granite's colours are
// library constants: they are evaluated, and the calls land in the property
// lines qualified as a document writes them. Its grain size carries a @range,
// so it is lifted with the colours.
func TestAdjustInlinesReturnedCalls(t *testing.T) {
	adj := readAdjust(t)["stone.greyGranite"]
	if adj == nil {
		t.Fatal("stone.greyGranite has no adjust record")
	}
	if want := []string{"stone.stoneFinish", "stone.granite"}; !reflect.DeepEqual(adj.Calls, want) {
		t.Errorf("calls %v, want %v", adj.Calls, want)
	}
	names := []string{}
	for _, p := range adj.Parameters {
		names = append(names, p.Name+":"+p.Kind)
	}
	if want := []string{"rough:range", "cellSize:range", "baseColor:color", "fleckColorA:color", "fleckColorB:color"}; !reflect.DeepEqual(names, want) {
		t.Errorf("parameters %v, want %v", names, want)
	}
	// stone.graniteFeldspar, evaluated.
	if got, want := adj.Parameters[2].Default, []float64{0.665, 0.66, 0.65}; len(got) != 3 || math.Abs(got[0]-want[0]) > 1e-6 || math.Abs(got[1]-want[1]) > 1e-6 || math.Abs(got[2]-want[2]) > 1e-6 {
		t.Errorf("baseColor default %v is not stone.graniteFeldspar %v", got, want)
	}
	wantTemplate := strings.Join([]string{
		"${body}.material = stone.stoneFinish(${rough});",
		"${body}.color = stone.granite(${cellSize}, ${baseColor}, ${fleckColorA}, ${fleckColorB});",
	}, "\n")
	if adj.Template != wantTemplate {
		t.Errorf("template:\n%s\nwant:\n%s", adj.Template, wantTemplate)
	}
}

// A metal's finish(rough) declares nothing adjustable, so the metal has no
// record and the panel offers nothing to adjust.
func TestAdjustSkipsUnannotatedConstructors(t *testing.T) {
	adj := readAdjust(t)
	if _, ok := adj["metal.steel"]; ok {
		t.Errorf("metal.steel is adjustable, but finish() declares no parameter kinds")
	}
}

// A local the preset binds is renamed after the body, and a qualified name
// that happens to end in a local's name is not: rustySteel binds `steel` and
// its template must not touch it elsewhere.
func TestAdjustRenamesLocalsOnly(t *testing.T) {
	adj := readAdjust(t)["weathering.rustySteel"]
	if adj == nil {
		t.Fatal("weathering.rustySteel has no adjust record")
	}
	wantTemplate := strings.Join([]string{
		"${body}Surface, ${body}SurfaceColor = weathering.rustedSteel(${amount});",
		"${body}.material = ${body}Surface;",
		"${body}.color = ${body}SurfaceColor;",
	}, "\n")
	if adj.Template != wantTemplate {
		t.Errorf("template:\n%s\nwant:\n%s", adj.Template, wantTemplate)
	}
	if renamed := renameWords("metal.steel() + steel", map[string]string{"steel": "X"}); renamed != "metal.steel() + X" {
		t.Errorf("renameWords: %q", renamed)
	}
}

func describe(params []Parameter) string {
	var sb strings.Builder
	for _, p := range params {
		sb.WriteString(p.Name + ":" + p.Kind)
		if p.Min != nil {
			sb.WriteString(" min=" + strconv.FormatFloat(*p.Min, 'g', -1, 64))
		}
		if p.Max != nil {
			sb.WriteString(" max=" + strconv.FormatFloat(*p.Max, 'g', -1, 64))
		}
		if p.Anchor != "" {
			sb.WriteString(" anchor=" + p.Anchor)
		}
		sb.WriteString(" default=")
		for i, v := range p.Default {
			if i > 0 {
				sb.WriteString(",")
			}
			sb.WriteString(strconv.FormatFloat(v, 'g', -1, 64))
		}
		sb.WriteString("; ")
	}
	return sb.String()
}

// A pattern's sizes lift as ranges with their bounds: the terrain preset's
// scale and peak height, both in millimetres, with its snowBias left as the
// literal it is.
func TestAdjustLiftsATerrainsSizes(t *testing.T) {
	adj := readAdjust(t)["natural.terrain"]
	if adj == nil {
		t.Fatal("natural.terrain has no adjust record")
	}
	if want := []string{"natural.raisedTerrain"}; !reflect.DeepEqual(adj.Calls, want) {
		t.Errorf("calls %v, want %v", adj.Calls, want)
	}
	want := []Parameter{
		{Name: "scale", Kind: "range", Min: f(2), Max: f(200), Default: []float64{40}, Call: "natural.raisedTerrain", Arg: 0},
		{Name: "peakHeight", Kind: "range", Min: f(0.1), Max: f(4), Default: []float64{2.5}, Call: "natural.raisedTerrain", Arg: 1},
	}
	if !reflect.DeepEqual(adj.Parameters, want) {
		t.Errorf("parameters:\n got %s\nwant %s", describe(adj.Parameters), describe(want))
	}
	wantTemplate := strings.Join([]string{
		"${body}Surface, ${body}SurfaceColor = natural.raisedTerrain(${scale}, ${peakHeight}, 0);",
		"${body}.material = ${body}Surface;",
		"${body}.color = ${body}SurfaceColor;",
	}, "\n")
	if adj.Template != wantTemplate {
		t.Errorf("template:\n%s\nwant:\n%s", adj.Template, wantTemplate)
	}
}

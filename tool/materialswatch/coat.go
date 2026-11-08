package main

import (
	"encoding/binary"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// The coats' recolouring basis. A preset whose constructors take one @color
// parameter — the powder coats, the nylons — paints that colour wherever it
// paints anything, so its swatch is linear in it as a finish's swatch is in
// its albedo (basis.go, examples/viewer/panels/finish-tint.js). Two renders of each, the
// colour at 0 and at 1 with every other parameter at the preset's default,
// give the swatch of an adjusted body in whatever coat colour its document
// declares, with the preset's other parameters as they were. A preset with
// two colours or more (a granite's three minerals, lava's crust and glow)
// mixes them over the surface, which is no one albedo, and gets no basis.

// swatchSettings is how a swatch is lit and framed, the flags main takes.
type swatchSettings struct {
	extent     float64
	env        string
	light      bool
	ground     bool
	groundGrey float64
	checker    float64
}

// coatColorIndex is the index in adj.Parameters of the preset's one colour
// parameter, or -1 when it has none or several.
func coatColorIndex(adj *Adjustable) int {
	at := -1
	for i, p := range adj.Parameters {
		if p.Kind != "color" {
			continue
		}
		if at >= 0 {
			return -1
		}
		at = i
	}
	return at
}

var templatePlaceholderRE = regexp.MustCompile(`\$\{(\w+)\}`)

// fillTemplate fills an adjust template for a body, with each parameter's
// placeholder replaced by the name in values.
func fillTemplate(template, body string, values map[string]string) string {
	return templatePlaceholderRE.ReplaceAllStringFunc(template, func(m string) string {
		key := m[2 : len(m)-1]
		if key == "body" {
			return body
		}
		if v, ok := values[key]; ok {
			return v
		}
		return m
	})
}

// gslValue writes a parameter value as a GSL literal: a number, a list, or a
// matrix as a list of four columns.
func gslValue(vals []float64) string {
	num := func(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }
	switch len(vals) {
	case 1:
		return num(vals[0])
	case 16:
		cols := make([]string, 4)
		for c := range cols {
			cols[c] = gslValue(vals[4*c : 4*c+4])
		}
		return "[" + strings.Join(cols, ", ") + "]"
	}
	parts := make([]string, len(vals))
	for i, v := range vals {
		parts[i] = num(v)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// coatDoc is the swatch document of an adjustable preset with its colour
// parameter at albedo and every other at its default: the variables the panel
// would write, without their controls, the template filled for `ball`, and
// the swatch's ground.
func coatDoc(adj *Adjustable, albedo [3]float64, s swatchSettings) string {
	var sb strings.Builder
	values := map[string]string{}
	colorAt := coatColorIndex(adj)
	for i, p := range adj.Parameters {
		name := "swatch" + strings.ToUpper(p.Name[:1]) + p.Name[1:]
		vals := p.Default
		if i == colorAt {
			vals = albedo[:]
		}
		fmt.Fprintf(&sb, "%s = %s;\n", name, gslValue(vals))
		values[p.Name] = name
	}
	fmt.Fprintf(&sb, "ball = sphere(%g);\n", swatchRadius)
	sb.WriteString(fillTemplate(adj.Template, "ball", values))
	sb.WriteString("\n")
	return swatchTail(sb.String(), s.ground, s.groundGrey, s.checker)
}

// renderCoatBasis appends a coat's two renders to out, as half floats, laid
// out as a finish's are in basis.go.
func renderCoatBasis(out []byte, adj *Adjustable, width, height int, s swatchSettings) ([]byte, error) {
	for _, albedo := range basisAlbedos {
		_, radiance, err := renderRadianceDoc(coatDoc(adj, albedo, s), width, height, s.extent, s.env, s.light)
		if err != nil {
			return nil, err
		}
		for _, v := range radiance {
			out = binary.LittleEndian.AppendUint16(out, floatToHalf(v))
		}
	}
	return out, nil
}

// buildCoatBasis renders the coat basis of every entry with one colour
// parameter, in catalog order. It returns nil when no entry has one.
func buildCoatBasis(entries []Entry, adjust map[string]*Adjustable, width, height int, s swatchSettings) (*FinishBasis, []byte, error) {
	basis := &FinishBasis{File: "coat-basis.bin", Width: width, Height: height}
	var raw []byte
	for _, e := range entries {
		name := e.Package + "." + e.Symbol
		adj := adjust[name]
		if adj == nil || coatColorIndex(adj) < 0 {
			continue
		}
		var err error
		if raw, err = renderCoatBasis(raw, adj, width, height, s); err != nil {
			return nil, nil, fmt.Errorf("%s coat basis: %w", name, err)
		}
		basis.Order = append(basis.Order, name)
	}
	if len(basis.Order) == 0 {
		return nil, nil, nil
	}
	return basis, raw, nil
}

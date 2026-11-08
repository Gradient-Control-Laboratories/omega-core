package main

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// snapshot is the subset of the Physically Based v2 materials response the
// generator reads. Unlisted fields are ignored (see droppedFields).
type snapshot struct {
	Header struct {
		SchemaVersion float64 `json:"schemaVersion"`
		Updated       int64   `json:"updated"`
		License       string  `json:"license"`
		Origin        string  `json:"origin"`
	} `json:"header"`
	Data []pbMaterial `json:"data"`
}

type pbMaterial struct {
	Name     string   `json:"name"`
	Category []string `json:"category"`
	Color    []struct {
		ColorSpace string    `json:"colorSpace"`
		Color      []float64 `json:"color"`
	} `json:"color"`
	Metalness         float64   `json:"metalness"`
	Roughness         float64   `json:"roughness"`
	IOR               *float64  `json:"ior"`
	Transmission      *float64  `json:"transmission"`
	TransmissionDepth *float64  `json:"transmissionDepth"`
	Density           []float64 `json:"density"`
	Description       string    `json:"description"`
	References        []struct {
		Title string `json:"title"`
	} `json:"references"`
}

// droppedFields are Physically Based fields with no place in GSL's material
// record; the package header lists the ones present in the snapshot.
var droppedFields = map[string]string{
	"complexIor":             "complex IOR (metals)",
	"specularColor":          "specular color / F82 tint (metals)",
	"subsurfaceRadius":       "subsurface radius",
	"thinFilmThickness":      "thin-film thickness",
	"thinFilmIor":            "thin-film IOR",
	"transmissionDispersion": "dispersion (Abbe number)",
	"volumeCoefficients":     "volume scattering coefficients",
	"diffuseRoughness":       "diffuse roughness",
	"viscosity":              "viscosity",
	"surfaceTension":         "surface tension",
	"acousticAbsorption":     "acoustic absorption",
}

// cmToMM converts transmissionDepth. The schema gives no unit; centimetres
// is inferred from the data (blood 0.08 and coffee 0.1 are only opaque at
// the depths real blood and coffee are in centimetres, not metres).
const cmToMM = 10.0

// defaultDepthMM is the absorption distance for a transmissive material the
// database gives no depth. GSL reads 0 as "no tint", which would render
// amber, sapphire and honey clear; 1 cm is the depth the database itself
// uses for most tinted liquids.
const defaultDepthMM = 10.0

func parseSnapshot(b []byte) (*snapshot, []string, error) {
	var s snapshot
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, nil, fmt.Errorf("parse snapshot: %w", err)
	}
	var raw struct {
		Data []map[string]json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, nil, err
	}
	seen := map[string]bool{}
	for _, m := range raw.Data {
		for k := range m {
			if desc, ok := droppedFields[k]; ok {
				seen[desc] = true
			}
		}
	}
	var dropped []string
	for d := range seen {
		dropped = append(dropped, d)
	}
	sort.Strings(dropped)
	return &s, dropped, nil
}

// identifier turns a Physically Based name into a GSL identifier:
// "Juice (Ruby Grapefruit)" -> juiceRubyGrapefruit, "MIT Black" -> mitBlack,
// "Vodka (40%)" -> vodka40, "Plastic (PLA)" -> plasticPLA.
func identifier(name string) string {
	words := strings.FieldsFunc(name, func(r rune) bool {
		return !(r < 128 && (unicode.IsLetter(r) || unicode.IsDigit(r)))
	})
	var b strings.Builder
	for i, w := range words {
		switch {
		case i == 0 && strings.ToUpper(w) == w:
			b.WriteString(strings.ToLower(w))
		case i == 0:
			b.WriteString(strings.ToLower(w[:1]) + w[1:])
		default:
			b.WriteString(strings.ToUpper(w[:1]) + w[1:])
		}
	}
	return b.String()
}

// typicalDensity reads the density array: one value, a [min, max] range
// (its midpoint), or [min, max, most common].
func typicalDensity(d []float64) (float64, bool) {
	switch len(d) {
	case 1:
		return d[0], true
	case 2:
		return (d[0] + d[1]) / 2, true
	case 3:
		return d[2], true
	}
	return 0, false
}

func linearColor(m pbMaterial) ([]float64, error) {
	for _, c := range m.Color {
		if c.ColorSpace == "srgb-linear" && len(c.Color) == 3 {
			return c.Color, nil
		}
	}
	return nil, fmt.Errorf("%s: no srgb-linear color", m.Name)
}

// encodeSRGB clamps a linear channel to [0, 1] and applies the sRGB transfer
// function, the representation `name.color` expects.
func encodeSRGB(l float64) float64 {
	l = math.Min(math.Max(l, 0), 1)
	if l <= 0.0031308 {
		return 12.92 * l
	}
	return 1.055*math.Pow(l, 1/2.4) - 0.055
}

func num(f float64) string {
	s := strconv.FormatFloat(math.Round(f*10000)/10000, 'f', -1, 64)
	if s == "-0" {
		return "0"
	}
	return s
}

// generate renders the physical GSL package from a snapshot.
func generate(s *snapshot, dropped []string) (string, error) {
	var b strings.Builder
	date := strconv.FormatInt(s.Header.Updated, 10)
	if len(date) >= 8 {
		date = date[:4] + "-" + date[4:6] + "-" + date[6:8]
	}
	fmt.Fprintf(&b, `
// Physically Based Materials (material/bulk/physical)
//
// GENERATED by tool/physicallybased from
// %s
// (snapshot %s, schema %s, %s).
// Do not edit by hand: change the generator or refresh the snapshot, then
// regenerate.
//
// Measured values for real-world materials. Like the metal and glass
// libraries, each one is a zero-argument function returning a PAIR: its
// material record and its colour, in that order.
//   cupMat, cupColor = physical.porcelain();
//   cup.material = cupMat;
//   cup.color = cupColor;
//   grams = volumeMM3 * physical.porcelainDensity * 1e-6;
// The colour is sRGB in [0, 1] (the database gives linear light;
// out-of-gamut values are clamped, then encoded). Density stays a separate
// constant, XDensity in kg/m^3 (the midpoint where the database gives only a
// range; absent where the database has no density): it is a property of the
// stuff rather than of its surface, so it is not part of the pair.
//
// Roughness is the database's reference value, often 0 (polished) for
// metals and liquids; vary it with material.withRoughness. Metals carry no
// IOR (their colour is the reflectance), so they get 1.5. Transmission
// depth is converted to the record's absorption distance in mm assuming the
// database's depths are centimetres: the schema gives no unit, and blood
// (0.08) and coffee (0.1) are only right at centimetre depths. A
// transmissive material with no depth gets 10 mm, so its colour still tints
// what comes through (amber, sapphire, honey).
//
%s//
// Self-contained (no library dependencies): records are spelled out as the
// rows material.make produces. Doc convention follows builtin/base.gsl:
// `+"`///`"+` lines are public docs.
`, s.Header.Origin, date, strconv.FormatFloat(s.Header.SchemaVersion, 'f', -1, 64), s.Header.License,
		wrap("// ", "Not carried over (no place in GSL's material record): "+strings.Join(dropped, ", ")+"."))

	byCategory := map[string][]pbMaterial{}
	var cats []string
	for _, m := range s.Data {
		c := "Other"
		if len(m.Category) > 0 {
			c = m.Category[0]
		}
		if section, ok := sectionOf[m.Name]; ok {
			c = section
		}
		if _, ok := byCategory[c]; !ok {
			cats = append(cats, c)
		}
		byCategory[c] = append(byCategory[c], m)
	}
	sort.Strings(cats)

	used := map[string]string{}
	for _, c := range cats {
		ms := byCategory[c]
		sort.Slice(ms, func(i, j int) bool { return ms[i].Name < ms[j].Name })
		fmt.Fprintf(&b, "\n// ===========================================================================\n// %s\n// ===========================================================================\n", c)
		for _, m := range ms {
			id := identifier(m.Name)
			if id == "" {
				return "", fmt.Errorf("material %q has no identifier", m.Name)
			}
			if prev, dup := used[id]; dup {
				return "", fmt.Errorf("materials %q and %q both map to %s", prev, m.Name, id)
			}
			used[id] = m.Name
			lin, err := linearColor(m)
			if err != nil {
				return "", err
			}
			ior := 1.5
			if m.IOR != nil {
				ior = *m.IOR
			}
			var trans, depthMM float64
			if m.Transmission != nil {
				trans = *m.Transmission
			}
			if trans > 0 {
				depthMM = defaultDepthMM
				if m.TransmissionDepth != nil {
					depthMM = *m.TransmissionDepth * cmToMM
				}
			}

			doc := m.Name + "."
			if m.Description != "" {
				doc += " " + strings.TrimSpace(m.Description)
			}
			if len(m.References) > 0 && m.References[0].Title != "" {
				doc += " Source: " + m.References[0].Title + "."
			}
			doc += " Returns material and colour."
			fmt.Fprintf(&b, "\n%s", wrap("/// ", doc))
			// @material declares the pair browsable. The record and the
			// colour leave together as one return, so a consumer cannot take
			// one without the other, and the annotation package checks the
			// shape of what comes back rather than a naming convention.
			fmt.Fprintf(&b, "@material %s() {\n", id)
			fmt.Fprintf(&b, "  return [[%s, %s, %s, 0], [0, 0, 0, 0], [%s, %s, 0, 0], [0, 0, 0, 0]], [%s, %s, %s];\n",
				num(m.Roughness), num(m.Metalness), num(ior), num(trans), num(depthMM),
				num(encodeSRGB(lin[0])), num(encodeSRGB(lin[1])), num(encodeSRGB(lin[2])))
			fmt.Fprintf(&b, "}\n")
			if d, ok := typicalDensity(m.Density); ok {
				fmt.Fprintf(&b, "/// Density of %s, kg/m^3.\n", strings.ToLower(m.Name[:1])+m.Name[1:])
				fmt.Fprintf(&b, "%sDensity = %s;\n", id, num(d))
			}
		}
	}
	return b.String(), nil
}

// wrap formats text as prefixed comment lines of at most 79 columns.
func wrap(prefix, text string) string {
	var b, line strings.Builder
	for _, w := range strings.Fields(text) {
		if line.Len() > 0 && len(prefix)+line.Len()+1+len(w) > 79 {
			b.WriteString(prefix + line.String() + "\n")
			line.Reset()
		}
		if line.Len() > 0 {
			line.WriteByte(' ')
		}
		line.WriteString(w)
	}
	if line.Len() > 0 {
		b.WriteString(prefix + line.String() + "\n")
	}
	return b.String()
}

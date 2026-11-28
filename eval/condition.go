//go:build cgo

package eval

import (
	"fmt"

	"github.com/Gradient-Control-Laboratories/omega-core/internal/fidget/convert"
)

// Condition is the state of the surfaces a query reads: where each layered
// surface stands, given its materials' builds (the films coats stand proud by)
// and their relief (the marks on them).
type Condition string

const (
	// AsBuilt is the field as the document builds it, no surface moved: what
	// every query read before conditions existed, and the default.
	AsBuilt Condition = "as-built"
	// Nominal moves each surface by its materials' builds and not their relief:
	// the part plated, its finish smoothed to the peaks' datum.
	Nominal Condition = "nominal"
	// AsMade moves each surface by its builds and its relief: the part as the
	// CPU renderer draws it, and as a displaced mesh meshes it.
	AsMade Condition = "as-made"
	// MaxMaterial is the part at its most material: each surface moved by its
	// builds, its relief at its peaks, 0, since relief is measured down from
	// them. Until spreads exist (each build at the top of its range) it reads
	// the field as Nominal does.
	MaxMaterial Condition = "max-material"
	// MinMaterial is the part at its least material, each relief at its depth
	// (material.reliefDepth). A feature's depth is a constant, so a
	// least-material claim is its band shifted by that depth, which the claims
	// generated from .tolerance do; SetCondition refuses it as a field.
	MinMaterial Condition = "min-material"
)

// ParseCondition reads a condition by name; "" is AsBuilt.
func ParseCondition(name string) (Condition, error) {
	switch c := Condition(name); c {
	case "", AsBuilt:
		return AsBuilt, nil
	case Nominal, AsMade, MaxMaterial:
		return c, nil
	case MinMaterial:
		return "", fmt.Errorf("condition %q is not a field: shift the claim's band by the surface's relief depth (material.reliefDepth) and read it at max-material or nominal", name)
	default:
		return "", fmt.Errorf("condition %q: one of as-built, nominal, as-made or max-material", name)
	}
}

// SetCondition sets the condition the evaluator's field queries read the
// document at: Eval, Extremum, Certify (and CertifyParams, CertifyOverSliders),
// Bounds, Enclose, Volume, SurfacePoint and Mesh. Render, DescribeAssets and
// Shader draw or describe the document and do not take it. AsBuilt, the
// default, builds every field exactly as before.
func (e *Evaluator) SetCondition(c Condition) error {
	c, err := ParseCondition(string(c))
	if err != nil {
		return err
	}
	e.surface = c.surface()
	return nil
}

// surface is where c puts the layered surfaces.
func (c Condition) surface() convert.Surface {
	switch c {
	case Nominal, MaxMaterial:
		return convert.Surface{Build: 1}
	case AsMade:
		return convert.Surface{Build: 1, Relief: 1}
	default:
		return convert.Surface{}
	}
}

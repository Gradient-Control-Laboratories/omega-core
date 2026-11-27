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
	// MaxMaterial and MinMaterial are the part at its most and least material.
	// They need each material's relief envelope and each layer's spread, which
	// documents cannot declare yet, so SetCondition refuses them.
	MaxMaterial Condition = "max-material"
	MinMaterial Condition = "min-material"
)

// ParseCondition reads a condition by name; "" is AsBuilt.
func ParseCondition(name string) (Condition, error) {
	switch c := Condition(name); c {
	case "", AsBuilt:
		return AsBuilt, nil
	case Nominal, AsMade:
		return c, nil
	case MaxMaterial, MinMaterial:
		return "", fmt.Errorf("condition %q needs each material's relief envelope and each layer's spread, which documents cannot declare yet; use as-built, nominal or as-made", name)
	default:
		return "", fmt.Errorf("condition %q: one of as-built, nominal or as-made", name)
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
	case Nominal:
		return convert.Surface{Build: 1}
	case AsMade:
		return convert.Surface{Build: 1, Relief: 1}
	default:
		return convert.Surface{}
	}
}

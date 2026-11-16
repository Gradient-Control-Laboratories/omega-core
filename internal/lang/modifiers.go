// Package lang defines shared GSL language rules and tables used across
// parser, assembler, and analysis passes.
package lang

// ModifierKind classifies a postfix modifier in a GSL expression.
type ModifierKind int

const (
	ModNone ModifierKind = iota
	ModIndex
	ModSwizzleScalar
	ModSwizzleVec
	ModDeriv
	ModBox
	ModSecond
	ModSurfaceFrame
)

// String returns a stable human-readable name for the modifier kind.
func (k ModifierKind) String() string {
	switch k {
	case ModNone:
		return "none"
	case ModIndex:
		return "index"
	case ModSwizzleScalar:
		return "swizzleScalar"
	case ModSwizzleVec:
		return "swizzleVec"
	case ModDeriv:
		return "deriv"
	case ModBox:
		return "box"
	case ModSecond:
		return "second"
	case ModSurfaceFrame:
		return "surfaceFrame"
	default:
		return "unknown"
	}
}

// ModifierTransition represents a valid (Prev, Cur) postfix modifier pair.
type ModifierTransition struct {
	Prev ModifierKind
	Cur  ModifierKind
}

// allowedModifierTransitions defines the language's valid left-to-right
// postfix modifier transitions, matching doc/gsl_lang_spec_current.md.
var allowedModifierTransitions = map[ModifierKind][]ModifierKind{
	ModIndex: {
		ModIndex,
		ModSwizzleScalar,
		ModSwizzleVec,
		ModDeriv,
		ModBox,
		ModSecond,
		ModSurfaceFrame,
	},
	ModSwizzleScalar: {
		ModSwizzleScalar,
		ModSwizzleVec,
		ModDeriv,
	},
	ModSwizzleVec: {
		ModIndex,
		ModSwizzleScalar,
		ModSwizzleVec,
		ModDeriv,
	},
	ModDeriv: {
		ModIndex,
		ModSwizzleScalar,
		ModSwizzleVec,
		ModDeriv,
	},
	ModBox: {
		ModIndex,
		ModDeriv,
	},
	ModSecond: {
		ModSurfaceFrame,
		ModDeriv,
	},
	ModSurfaceFrame: {
		ModDeriv,
	},
}

var allowedTransitionSet = func() map[ModifierTransition]bool {
	set := make(map[ModifierTransition]bool)
	for prev, curs := range allowedModifierTransitions {
		for _, cur := range curs {
			set[ModifierTransition{Prev: prev, Cur: cur}] = true
		}
	}
	return set
}()

// IsAllowedModifierTransition reports whether cur may immediately follow prev.
func IsAllowedModifierTransition(prev, cur ModifierKind) bool {
	return allowedTransitionSet[ModifierTransition{Prev: prev, Cur: cur}]
}

// AllowedModifierTransitions returns a copy of the valid modifier transition
// table so callers can inspect or iterate over it without mutating it.
func AllowedModifierTransitions() map[ModifierKind][]ModifierKind {
	out := make(map[ModifierKind][]ModifierKind, len(allowedModifierTransitions))
	for prev, curs := range allowedModifierTransitions {
		cp := make([]ModifierKind, len(curs))
		copy(cp, curs)
		out[prev] = cp
	}
	return out
}

package lang_test

import (
	"testing"

	"gradientcontrol.com/omega/core/internal/lang"
)

func TestModifierKindString(t *testing.T) {
	cases := []struct {
		kind lang.ModifierKind
		want string
	}{
		{lang.ModNone, "none"},
		{lang.ModIndex, "index"},
		{lang.ModSwizzleScalar, "swizzleScalar"},
		{lang.ModSwizzleVec, "swizzleVec"},
		{lang.ModDeriv, "deriv"},
		{lang.ModBox, "box"},
		{lang.ModSecond, "second"},
		{lang.ModSurfaceFrame, "surfaceFrame"},
		{lang.ModifierKind(999), "unknown"},
	}
	for _, tc := range cases {
		if got := tc.kind.String(); got != tc.want {
			t.Errorf("ModifierKind(%d).String() = %q, want %q", tc.kind, got, tc.want)
		}
	}
}

func TestAllowedModifierTransitions(t *testing.T) {
	table := lang.AllowedModifierTransitions()
	if len(table) == 0 {
		t.Fatal("AllowedModifierTransitions() returned empty map")
	}

	for prev, curs := range table {
		for _, cur := range curs {
			if !lang.IsAllowedModifierTransition(prev, cur) {
				t.Errorf("IsAllowedModifierTransition(%s, %s) = false, want true", prev, cur)
			}
		}
	}

	if lang.IsAllowedModifierTransition(lang.ModNone, lang.ModIndex) {
		t.Error("expected ModNone -> ModIndex to be false in transition table")
	}
	if lang.IsAllowedModifierTransition(lang.ModSurfaceFrame, lang.ModSecond) {
		t.Error("expected ModSurfaceFrame -> ModSecond to be disallowed")
	}

	// Verify AllowedModifierTransitions returns a defensive copy.
	table[lang.ModSurfaceFrame] = nil
	delete(table, lang.ModBox)
	fresh := lang.AllowedModifierTransitions()
	if len(fresh[lang.ModSurfaceFrame]) == 0 {
		t.Error("mutating returned slice mutated internal transition table")
	}
	if len(fresh[lang.ModBox]) == 0 {
		t.Error("deleting map key mutated internal transition table")
	}
}

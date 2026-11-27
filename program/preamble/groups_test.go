package preamble

import (
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/program"
)

// TestBuiltinGroupsNameRealBuiltins: a group lists only described built-ins,
// each in one group at most, and never one the index already hides.
func TestBuiltinGroupsNameRealBuiltins(t *testing.T) {
	seen := map[string]string{}
	for _, g := range builtinGroups {
		for _, n := range g.Names {
			if _, ok := builtinDescriptions[n]; !ok {
				t.Errorf("%s lists %s, which is not a described built-in", g.Path, n)
			}
			if annotationBuiltins[n] {
				t.Errorf("%s lists %s, which the docs leave to its annotation", g.Path, n)
			}
			if other, dup := seen[n]; dup {
				t.Errorf("%s is in both %s and %s", n, other, g.Path)
			}
			seen[n] = g.Path
		}
	}
}

// TestIndexListsCoreBuiltinsAndNamesGroups: the index's built-in list leaves
// the grouped ones out but names every group; the full list keeps them all.
func TestIndexListsCoreBuiltinsAndNamesGroups(t *testing.T) {
	core := coreBuiltinDescriptiveText()
	full := PrimitiveAndBuiltinDescriptiveText()
	for _, g := range builtinGroups {
		if !strings.Contains(core, "- **`"+g.Path+"`**") {
			t.Errorf("the index does not name group %s", g.Path)
		}
		for _, n := range g.Names {
			if strings.Contains(core, "* **`"+n+"`**") {
				t.Errorf("the index lists grouped built-in %s", n)
			}
			if !strings.Contains(full, "* **`"+n+"`**") {
				t.Errorf("the full reference lacks %s", n)
			}
		}
	}
	for _, n := range []string{"length", "clamp", "normalize"} {
		if !strings.Contains(core, "* **`"+n+"`**") {
			t.Errorf("the index lacks core built-in %s", n)
		}
	}
}

// TestReservedRuleExamplesAreReserved: the words the rule cites as examples
// are ones the compiler rejects.
func TestReservedRuleExamplesAreReserved(t *testing.T) {
	for _, w := range []string{"f32", "u8", "bool", "async", "chan"} {
		if !program.IsReservedWord(w) {
			t.Errorf("the reserved-word rule cites %q, which is not reserved", w)
		}
		if !strings.Contains(reservedWordsRuleText(), "`"+w+"`") {
			t.Errorf("the rule no longer cites %q; update this test's list", w)
		}
	}
}

// TestSkillIndexUsesTheRuleAndTheBrief: the index carries the reserved-word
// rule rather than the list, and the builtin package in brief.
func TestSkillIndexUsesTheRuleAndTheBrief(t *testing.T) {
	index := GSLSkillIndex()
	if strings.Contains(index, ReservedWordsDescriptiveText()) {
		t.Error("the index carries the full reserved-word list")
	}
	if !strings.Contains(index, reservedWordsRuleText()) {
		t.Error("the index lacks the reserved-word rule")
	}
	if !strings.Contains(index, "## **The `builtin` Package, in Brief**") {
		t.Error("the index lacks the builtin brief")
	}
	if !strings.Contains(GSLAuthorSkill(), ReservedWordsDescriptiveText()) {
		t.Error("--all no longer lists every reserved word")
	}
}

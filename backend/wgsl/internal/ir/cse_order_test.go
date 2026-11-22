package ir

import (
	"strings"
	"testing"
)

// A let that rule 2 moves up to the first use of its value must not read a
// name declared between there and where it stood: builtin's
// profileCornerStepXY repeats length(start - previous) before
// `previousChord = start - previous` and `previousLength =
// length(previousChord)`, and moving previousLength up as written read
// previousChord before its declaration. The moved let takes the first use's
// text instead, the same value.
func TestCSE_MovedLetReadsOnlyEarlierNames(t *testing.T) {
	got := cseFormat(t, cseFn(
		let("u", "length(start - previous) + 1.0"),
		let("chord", "start - previous"),
		let("len", "length(chord)"),
		ret("u + len * length(start - previous)")))
	lines := strings.Split(got, "\n")
	declaredAt := map[string]int{}
	for i, line := range lines {
		if f := strings.Fields(line); len(f) > 1 && f[0] == "let" {
			declaredAt[f[1]] = i
		}
	}
	for i, line := range lines {
		for name, at := range declaredAt {
			if at > i && strings.Contains(line, name+")") || at > i && strings.Contains(line, name+" ") && !strings.Contains(line, "let "+name) {
				t.Errorf("line %d reads %s, declared on line %d:\n%s", i, name, at, got)
			}
		}
	}
}

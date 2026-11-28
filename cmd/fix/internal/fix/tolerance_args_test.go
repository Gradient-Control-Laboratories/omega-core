package fix

import (
	"strings"
	"testing"
)

// The removed three-argument condition forms are rewritten to the size they
// gave, exactly: the nominal plus that condition's deviation, a shaft's upper
// at maximum material and lower at minimum, a hole's the other way round, and
// a bearing's seat and housing through its bore and outside diameter.
func TestStdlibFix_ConditionArgumentsBecomeDeviations(t *testing.T) {
	in := `a = tolerance.shaftDia(25.0, tolerance.k5, tolerance.maxMaterial) / 2.0;
b = tolerance.shaftDia(25.0, tolerance.k5, tolerance.minMaterial);
c = tolerance.holeDia(52.0, tolerance.H7, tolerance.maxMaterial);
d = tolerance.holeDia(52.0, tolerance.H7, tolerance.minMaterial);
e = bearing.seatDia(6205.0, tolerance.k5, tolerance.maxMaterial);
f = bearing.housingDia(6205.0, tolerance.H7, tolerance.minMaterial);
return sphere(a);
`
	want := `a = (25.0 + tolerance.shaftUpper(25.0, tolerance.k5)) / 2.0;
b = (25.0 + tolerance.shaftLower(25.0, tolerance.k5));
c = (52.0 + tolerance.holeLower(52.0, tolerance.H7));
d = (52.0 + tolerance.holeUpper(52.0, tolerance.H7));
e = (bearing.bore(6205.0) + tolerance.shaftUpper(bearing.bore(6205.0), tolerance.k5));
f = (bearing.outerDia(6205.0) + tolerance.holeUpper(bearing.outerDia(6205.0), tolerance.H7));
return sphere(a);
`
	got, changed, notes, err := stdlibFixer.ApplyWithNotes(in)
	if err != nil {
		t.Fatalf("ApplyWithNotes: %v", err)
	}
	if !changed || len(notes) != 0 {
		t.Fatalf("changed %v, notes %+v", changed, notes)
	}
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
	if again, changed2, _, _ := stdlibFixer.ApplyWithNotes(got); changed2 || again != want {
		t.Fatalf("not idempotent:\n%s", again)
	}
}

// A condition the fixer cannot read is left, with a note naming the
// deviations to write it with.
func TestStdlibFix_UnreadableConditionIsReported(t *testing.T) {
	in := "r = tolerance.shaftDia(25.0, tolerance.k5, cond);\nreturn sphere(r);\n"
	got, changed, notes, err := stdlibFixer.ApplyWithNotes(in)
	if err != nil {
		t.Fatalf("ApplyWithNotes: %v", err)
	}
	if changed || got != in {
		t.Fatalf("changed the call it cannot read:\n%s", got)
	}
	if len(notes) != 1 || notes[0].Line != 1 || !strings.Contains(notes[0].Message, "tolerance.shaftUpper or tolerance.shaftLower") {
		t.Fatalf("notes %+v", notes)
	}
}

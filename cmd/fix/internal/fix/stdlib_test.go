package fix

import (
	"strings"
	"testing"
)

func TestStdlibFix_RoundExactTwoArgThunks(t *testing.T) {
	in := `a = sphere([-0.5, 0, 0], 1);
b = sphere([0.5, 0, 0], 1);
u = unionRoundExact(() { return a; }, () { return b; });
i = intersectRoundExact(() { return a; }, () { return b; });
return u;
`
	want := `a = sphere([-0.5, 0, 0], 1);
b = sphere([0.5, 0, 0], 1);
u = unionRoundExact(a, b, 0);
i = intersectRoundExact(a, b, 0);
return u;
`
	got, changed, notes, err := stdlibFixer.ApplyWithNotes(in)
	if err != nil {
		t.Fatalf("ApplyWithNotes: %v", err)
	}
	if !changed {
		t.Fatal("expected changed=true")
	}
	if len(notes) != 0 {
		t.Fatalf("unexpected notes: %+v", notes)
	}
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}

	// Idempotent on second run.
	got2, changed2, _, err := stdlibFixer.ApplyWithNotes(got)
	if err != nil {
		t.Fatalf("second ApplyWithNotes: %v", err)
	}
	if changed2 || got2 != want {
		t.Fatalf("expected idempotent result, changed=%v\ngot:\n%s", changed2, got2)
	}
}

func TestStdlibFix_RoundExactFourArgThunksAndNamedThunk(t *testing.T) {
	in := `s1() { return sphere([0, 0, 0], 1); }
n1 = normalize(p);
n2 = normalize(p - [1.2, 0, 0]);
lens = intersectRoundExact(
  s1,
  () { return sphere([1.2, 0, 0], 1); },
  n1, n2);
return lens;
`
	want := `s1() { return sphere([0, 0, 0], 1); }
n1 = normalize(p);
n2 = normalize(p - [1.2, 0, 0]);
lens = intersectRoundExact(
  s1(),
  sphere([1.2, 0, 0], 1),
  0, n1, n2);
return lens;
`
	got, changed, _, err := stdlibFixer.ApplyWithNotes(in)
	if err != nil {
		t.Fatalf("ApplyWithNotes: %v", err)
	}
	if !changed {
		t.Fatal("expected changed=true")
	}
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestStdlibFix_LeavesNewRoundExactFormsUntouched(t *testing.T) {
	in := `a = sphere([0, 0, 0], 1);
b = box([1, 1, 1]);
c = cylinder(0.5, 2);
u1 = unionRoundExact([a, b], 0.2);
u2 = unionRoundExact(a, b, c, 0.2);
u3 = intersectRoundExact(a, b, 0, a.dp, b.dp);
return u1;
`
	got, changed, notes, err := stdlibFixer.ApplyWithNotes(in)
	if err != nil {
		t.Fatalf("ApplyWithNotes: %v", err)
	}
	if changed || len(notes) != 0 || got != in {
		t.Fatalf("expected no changes/notes, got changed=%v notes=%+v", changed, notes)
	}
}

func TestStdlibFix_RemovedSweepLoftEmitsSuggestions(t *testing.T) {
	in := `s = sweepLoft(circle(1), circle(2), [0, 0, 0], [5, 0, 5], [10, 0, 0]);
t = builtin.bezierTubeDir([0, 0, 0], [5, 0, 5], [10, 0, 0], 1);
return min(s, t);
`
	got, changed, notes, err := stdlibFixer.ApplyWithNotes(in)
	if err != nil {
		t.Fatalf("ApplyWithNotes: %v", err)
	}
	if changed || got != in {
		t.Fatalf("expected file unchanged, got changed=%v", changed)
	}
	if len(notes) != 2 {
		t.Fatalf("got %d notes, want 2: %+v", len(notes), notes)
	}
	if notes[0].Line != 1 || !strings.Contains(notes[0].Message, "sweepLoft") || !strings.Contains(notes[0].Message, "sweepOpenEdges") {
		t.Errorf("unexpected note[0]: %+v", notes[0])
	}
	if notes[1].Line != 2 || !strings.Contains(notes[1].Message, "bezierTubeDir") {
		t.Errorf("unexpected note[1]: %+v", notes[1])
	}
}

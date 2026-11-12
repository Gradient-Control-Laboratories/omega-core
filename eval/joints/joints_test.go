package joints

import (
	"strings"
	"testing"
)

const twoJoints = `
// a joint in a comment is not a joint: a, b, c = joint.bolted(x);
d = 6.0;
screw1, lid1, base1 = joint.bolted(lidBody, baseBody, [0.0, 0.0, 0.0],
                                   [0.0, 0.0, -1.0], d, 7.5, fastener.fitMedium);
screw2, lid2, base2 = joint.bolted(lid1, base1, seat2, dir2, d, joint.boltedLength(d, 7.5) - 12.5, fit);
cover = box([-20.0, -20.0, 4.0], [20.0, 20.0, 8.0]);
return screw1, lid2, base2, screw2, min(cover, other), max(screw1, cover);
`

// TestFind: both top-level joints are found with their names, arguments and
// lines; a helper call is not a joint; the top-level return is split.
func TestFind(t *testing.T) {
	found, returns, skipped := Find(twoJoints)
	if len(skipped) != 0 {
		t.Errorf("skipped = %v, want none", skipped)
	}
	if len(found) != 2 {
		t.Fatalf("found %d joints, want 2: %+v", len(found), found)
	}
	j := found[0]
	if j.Kind != "bolted" || strings.Join(j.Names, ",") != "screw1,lid1,base1" || j.Line != 4 {
		t.Errorf("joint 1 = %+v", j)
	}
	if len(j.Args) != 7 || j.Args[2] != "[0.0, 0.0, 0.0]" || j.Args[6] != "fastener.fitMedium" {
		t.Errorf("joint 1 args = %q", j.Args)
	}
	if found[1].Args[5] != "joint.boltedLength(d, 7.5) - 12.5" || found[1].Line != 6 {
		t.Errorf("joint 2 = %+v", found[1])
	}
	want := []string{"screw1", "lid2", "base2", "screw2", "min(cover, other)", "max(screw1, cover)"}
	if strings.Join(returns, "|") != strings.Join(want, "|") {
		t.Errorf("returns = %q, want %q", returns, want)
	}
}

// TestFindSkipped: a joint in a function body, one not destructured into
// three parts and one with the wrong number of arguments are each reported,
// never silently passed.
func TestFindSkipped(t *testing.T) {
	src := `
part(x) {
  s, l, b = joint.bolted(x, x, [0.0, 0.0, 0.0], [0.0, 0.0, -1.0], 6.0, 7.5, 2.0);
  return min(s, l, b);
}
all = joint.bolted(a, b, s, n, 6.0, 7.5, 2.0);
s, l, b = joint.bolted(a, b, s, n, 6.0);
return part(1.0);
`
	found, returns, skipped := Find(src)
	if len(found) != 0 {
		t.Errorf("found %+v, want none", found)
	}
	if len(skipped) != 3 {
		t.Fatalf("skipped = %q, want 3 notes", skipped)
	}
	for i, want := range []string{"line 3: a joint call inside a function body", "line 6: joint.bolted is not destructured", "line 7: joint.bolted takes 7 arguments"} {
		if !strings.HasPrefix(skipped[i], want) {
			t.Errorf("skipped[%d] = %q, want prefix %q", i, skipped[i], want)
		}
	}
	if len(returns) != 1 || returns[0] != "part(1.0)" {
		t.Errorf("returns = %q: the function body's return is not the document's", returns)
	}
}

// TestOthers: the screw is held clear of every returned part, a min(...)
// split into its parts, less the joint's own names and anything naming the
// screw.
func TestOthers(t *testing.T) {
	found, returns, _ := Find(twoJoints)
	if got, want := strings.Join(Others(found[0], returns), "|"), "lid2|base2|screw2|cover|other"; got != want {
		t.Errorf("joint 1 others = %q, want %q", got, want)
	}
	if got, want := strings.Join(Others(found[1], returns), "|"), "screw1|cover|other|max(screw1, cover)"; got != want {
		t.Errorf("joint 2 others = %q, want %q", got, want)
	}
	if got := splitMins([]string{"min(a, min(b, c), max(d, e))", "f"}); strings.Join(got, "|") != "a|b|c|max(d, e)|f" {
		t.Errorf("splitMins = %q", got)
	}
}

// TestClaims: the four claims of the joint itself, then the screw against
// the rest of the assembly, with both allowances.
func TestClaims(t *testing.T) {
	found, returns, _ := Find(twoJoints)
	claims := Claims(found[0], Others(found[0], returns))
	if len(claims) != 5 {
		t.Fatalf("got %d claims, want 5: %+v", len(claims), claims)
	}
	if want := "max(screw1, lid1, joint.boltedSeat(([0.0, 0.0, 0.0]), ([0.0, 0.0, -1.0]), (d), (fastener.fitMedium)))"; claims[0].Expr != want {
		t.Errorf("claim 0 = %s\nwant %s", claims[0].Expr, want)
	}
	if !strings.HasSuffix(claims[2].Expr, "neg((lidBody)))") || !strings.HasSuffix(claims[3].Expr, "neg((baseBody)))") {
		t.Errorf("context claims should read the bodies the joint was handed: %s / %s", claims[2].Expr, claims[3].Expr)
	}
	rest := claims[4].Expr
	if !strings.HasPrefix(rest, "max(screw1, min(lid2, base2, screw2, cover, other), joint.boltedSeat(") ||
		!strings.Contains(rest, "joint.boltedEngaged(") {
		t.Errorf("rest-of-assembly claim = %s", rest)
	}
	for _, c := range claims {
		if c.Hint == "" {
			t.Errorf("claim %q has no repair hint", c.Name)
		}
	}
	if len(Claims(found[0], nil)) != 4 {
		t.Error("with nothing else returned there is no assembly claim")
	}
}

// TestRegion: the box runs from a diameter above the seat (plus 1) to a
// diameter past the tip (plus 1), padded by the counterbore radius plus 1.
func TestRegion(t *testing.T) {
	box, err := Region([3]float64{0, 0, 0}, [3]float64{0, 0, -2}, 6, 20, 11)
	if err != nil {
		t.Fatal(err)
	}
	want := [6]float64{-6.5, -6.5, -33.5, 6.5, 6.5, 13.5}
	if box != want {
		t.Errorf("box = %v, want %v", box, want)
	}
	if _, err := Region([3]float64{}, [3]float64{0, 0, -1}, 6, 0, 11); err == nil {
		t.Error("a zero length (no catalog screw that long) should be an error")
	}
}

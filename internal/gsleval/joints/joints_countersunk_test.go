package joints_test

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/internal/gsleval/joints"
)

// countersunkDoc is a 3 mm cover screwed flush to a plate with one ISO 10642
// M3: the seat on the cover's top face, the shank driven down through the
// cover into the plate.
const countersunkDoc = `d = 3.0;
fit = fastener.fitMedium;
grip = 3.0;
coverBody = box([-10.0, -10.0, 0.0 - grip], [10.0, 10.0, 0.0]);
plateBody = box([-10.0, -10.0, -15.0], [10.0, 10.0, 0.0 - grip]);
screw, cover, plate = joint.countersunk(coverBody, plateBody, [0.0, 0.0, 0.0], [0.0, 0.0, -1.0], d, grip, fit);
return screw, cover, plate;
`

// TestCountersunkJoint: a countersunk joint is found like a bolted one and
// proves its claims. A cover thinner than the head is deep (1.86 for an M3)
// sinks the head into the plate, which the engagement claim disproves; a seat
// so near the cover's edge that the countersink breaks out of it is disproven
// by the claim that the cover holds the head round its cone.
func TestCountersunkJoint(t *testing.T) {
	rep := joints.Run(countersunkDoc, prover(t, countersunkDoc))
	if len(rep.Joints) != 1 || rep.Joints[0].Kind != "countersunk" || len(rep.Skipped) != 0 {
		t.Fatalf("found %+v, skipped %v", rep.Joints, rep.Skipped)
	}
	if !rep.AllProven {
		t.Fatalf("every claim should prove: %q", verdicts(rep))
	}
	if n := len(rep.Joints[0].Claims); n != 4 {
		t.Errorf("%d claims, want 4 (nothing else is returned)", n)
	}

	cases := []struct{ name, from, to, claim string }{
		{"cover thinner than the head", "grip = 3.0;", "grip = 1.5;", "screw overlaps the threaded part only in its thread"},
		{"countersink breaks out of an edge", "[0.0, 0.0, 0.0], [0.0, 0.0, -1.0]", "[8.5, 0.0, 0.0], [0.0, 0.0, -1.0]",
			"the clamped part holds the head all round its cone"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			broken := strings.Replace(countersunkDoc, c.from, c.to, 1)
			if broken == countersunkDoc {
				t.Fatalf("edit %q did not apply", c.from)
			}
			rep := joints.Run(broken, prover(t, broken))
			if rep.AllProven {
				t.Fatalf("should fail: %q", verdicts(rep))
			}
			for _, cl := range rep.Joints[0].Claims {
				if cl.Name == c.claim && !cl.Disproven {
					t.Errorf("%q should be disproven: %q", c.claim, verdicts(rep)[0])
				}
			}
		})
	}
}

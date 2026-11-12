//go:build cgo

package joints_test

import (
	"context"
	"math"
	"reflect"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/eval"
	"gradientcontrol.com/omega/core/eval/joints"
	"gradientcontrol.com/omega/core/internal/source"
)

// TestInterfaceValuesMatchFusedJoints: an interface declared as one value
// builds what the call that takes its parts loose builds — joint.insert what
// joint.bolted and joint.countersunk do, for each head, and the planar and
// prismatic accessors what joint.planar and joint.slide return.
func TestInterfaceValuesMatchFusedJoints(t *testing.T) {
	doc := `grip = 3.0;
coverBody = box([-10.0, -10.0, 0.0 - grip], [10.0, 10.0, 0.0]);
plateBody = box([-10.0, -10.0, -15.0], [10.0, 10.0, 0.0 - grip]);
bolt = joint.cylindrical([0.0, 0.0, 0.0], [0.0, 0.0, -1.0], 3.0, fastener.fitMedium);
sF, cF, pF = joint.insert(bolt, joint.flatHead, coverBody, plateBody, grip);
sC, cC, pC = joint.countersunk(coverBody, plateBody, [0.0, 0.0, 0.0], [0.0, 0.0, -1.0], 3.0, grip, fastener.fitMedium);
sS, cS, pS = joint.insert(bolt, joint.socketHead, coverBody, plateBody, grip);
sB, cB, pB = joint.bolted(coverBody, plateBody, [0.0, 0.0, 0.0], [0.0, 0.0, -1.0], 3.0, grip, fastener.fitMedium);
face = joint.planarFace([0.0, 0.0, -7.5], vector.dirZ);
above, below = joint.planar([0.0, 0.0, -7.5], vector.dirZ);
guide = joint.prismatic([0.0, 17.5, -4.0], vector.dirX, vector.dirZ, 6.0, 5.0, 0.3);
channel, envelope = joint.slide([0.0, 17.5, -4.0], vector.dirX, vector.dirZ, 6.0, 5.0, 0.3);
return sF, cF, pF;
`
	// On the axis in the cover and the plate, in the countersink and the
	// counterbore, in the clearance, outside everything, and in the channel.
	points := [][3]float64{
		{0, 0, -0.5}, {1.2, 0, -0.2}, {0, 0, -6}, {3, 0, -1}, {2.5, 0, -0.1},
		{1.6, 0, -2}, {0, 0, 2}, {10, 20.8, -4}, {3, 17, -1},
	}
	pairs := [][2]string{
		{"sF", "sC"}, {"cF", "cC"}, {"pF", "pC"},
		{"sS", "sB"}, {"cS", "cB"}, {"pS", "pB"},
		{"joint.above(face)", "above"}, {"joint.below(face)", "below"},
		{"joint.channel(guide)", "channel"}, {"joint.envelope(guide)", "envelope"},
	}
	var exprs []string
	for _, pr := range pairs {
		exprs = append(exprs, pr[0]+" - "+pr[1])
	}
	e := eval.New(&source.MockManager{}, nil)
	groups, err := e.EvalGsl(context.Background(), nil, doc, "", exprs, points)
	if err != nil {
		t.Fatal(err)
	}
	for i, ex := range exprs {
		for j, pt := range points {
			if got := float64(groups[i].Samples[0][j].Value); math.Abs(got) > 1e-6 {
				t.Errorf("%s at %v = %g, want 0", ex, pt, got)
			}
		}
	}
}

// insertDoc is a cover screwed to a plate through two declared interfaces: a
// flat head flush in the cover's top face and a socket head seated on it.
const insertDoc = `grip = 3.0;
coverBody = box([-20.0, -10.0, 0.0 - grip], [20.0, 10.0, 0.0]);
plateBody = box([-20.0, -10.0, -15.0], [20.0, 10.0, 0.0 - grip]);
flat = joint.cylindrical([-8.0, 0.0, 0.0], [0.0, 0.0, -1.0], 3.0, fastener.fitMedium);
socket = joint.cylindrical([8.0, 0.0, 0.0], [0.0, 0.0, -1.0], 3.0, fastener.fitMedium);
screwF, cover1, plate1 = joint.insert(flat, joint.flatHead, coverBody, plateBody, grip);
screwS, cover2, plate2 = joint.insert(socket, joint.socketHead, cover1, plate1, grip);
return screwF, screwS, cover2, plate2;
`

// TestInsertJoint: a fastener inserted into an interface is proven as the
// joint its head makes, with the claims and verdicts of the same joint written
// with joint.countersunk and joint.bolted. A cover thinner than the flat head
// is deep is disproven as it is there, and a head that is not one of
// joint.insert's is an error, not a pass.
func TestInsertJoint(t *testing.T) {
	rep := joints.Run(insertDoc, prover(t, insertDoc))
	if len(rep.Joints) != 2 || len(rep.Skipped) != 0 {
		t.Fatalf("found %+v, skipped %v", rep.Joints, rep.Skipped)
	}
	if k := []string{rep.Joints[0].Kind, rep.Joints[1].Kind}; !reflect.DeepEqual(k, []string{"countersunk", "bolted"}) {
		t.Errorf("kinds %v, want the head's: [countersunk bolted]", k)
	}
	if c := rep.Joints[0].Call(); !strings.Contains(c, "joint.insert(flat, joint.flatHead,") {
		t.Errorf("the report should name the call as written, got %q", c)
	}
	if !rep.AllProven {
		t.Fatalf("every claim should prove: %q", verdicts(rep))
	}

	fused := insertDoc
	for _, r := range [][2]string{
		{"joint.insert(flat, joint.flatHead, coverBody, plateBody, grip)",
			"joint.countersunk(coverBody, plateBody, [-8.0, 0.0, 0.0], [0.0, 0.0, -1.0], 3.0, grip, fastener.fitMedium)"},
		{"joint.insert(socket, joint.socketHead, cover1, plate1, grip)",
			"joint.bolted(cover1, plate1, [8.0, 0.0, 0.0], [0.0, 0.0, -1.0], 3.0, grip, fastener.fitMedium)"},
	} {
		if !strings.Contains(fused, r[0]) {
			t.Fatalf("edit %q did not apply", r[0])
		}
		fused = strings.Replace(fused, r[0], r[1], 1)
	}
	want := joints.Run(fused, prover(t, fused))
	if got := verdicts(rep); !reflect.DeepEqual(got, verdicts(want)) {
		t.Errorf("verdicts %q, the fused joints' %q", got, verdicts(want))
	}

	t.Run("cover thinner than the head", func(t *testing.T) {
		broken := strings.Replace(insertDoc, "grip = 3.0;", "grip = 1.5;", 1)
		rep := joints.Run(broken, prover(t, broken))
		if rep.AllProven {
			t.Fatalf("should fail: %q", verdicts(rep))
		}
		for _, cl := range rep.Joints[0].Claims {
			if cl.Name == "screw overlaps the threaded part only in its thread" && !cl.Disproven {
				t.Errorf("%q should be disproven: %q", cl.Name, verdicts(rep)[0])
			}
		}
	})

	t.Run("not a head", func(t *testing.T) {
		broken := strings.Replace(insertDoc, "joint.insert(flat, joint.flatHead,", "joint.insert(flat, 3.0,", 1)
		rep := joints.Run(broken, prover(t, broken))
		if rep.AllProven || rep.Joints[0].Error == "" || len(rep.Joints[0].Claims) != 0 {
			t.Fatalf("an unknown head should be an error with no claims: %+v", rep.Joints[0])
		}
	})
}

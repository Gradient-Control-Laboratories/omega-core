//go:build cgo

package joints_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/eval"
	"gradientcontrol.com/omega/core/eval/joints"
	"gradientcontrol.com/omega/core/internal/source"
)

// prover proves joint claims with eval, as the gsl CLI does.
func prover(t *testing.T, code string) joints.Prover {
	e := eval.New(&source.MockManager{}, nil)
	ctx := context.Background()
	return joints.Prover{
		Eval: func(exprs []string, at [3]float64) ([][]float64, error) {
			groups, err := e.EvalGsl(ctx, nil, code, "", exprs, [][3]float64{at})
			if err != nil {
				return nil, err
			}
			vals := make([][]float64, len(groups))
			for i, g := range groups {
				for _, comp := range g.Samples {
					vals[i] = append(vals[i], float64(comp[0].Value))
				}
			}
			return vals, nil
		},
		Certify: func(expr string, region [6]float64) (joints.Verdict, error) {
			r, err := e.Certify(ctx, nil, code, "", expr, region, 0, 0)
			if err != nil {
				return joints.Verdict{}, err
			}
			return joints.Verdict{Proven: r.Proven, Disproven: r.Disproven, Indeterminate: r.Indeterminate,
				Reason: r.Reason, CounterBox: r.CounterBox, Evals: r.Evals}, nil
		},
	}
}

func boltedJoint(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile("../../internal/testdata/example/Bolted Joint.gsl")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// verdicts is each joint's claims as "name: P|D|I".
func verdicts(rep joints.Report) [][]string {
	var out [][]string
	for _, j := range rep.Joints {
		var vs []string
		for _, c := range j.Claims {
			v := "I"
			if c.Proven {
				v = "P"
			} else if c.Disproven {
				v = "D"
			}
			if c.Error != "" {
				v = "E " + c.Error
			}
			vs = append(vs, c.Name+": "+v)
		}
		if j.Error != "" {
			vs = append(vs, "error: "+j.Error)
		}
		out = append(out, vs)
	}
	return out
}

// TestBoltedJointExample: both joints of the Bolted Joint example, the upright
// one and the one square to a 30-degree face, prove every claim at nominal.
func TestBoltedJointExample(t *testing.T) {
	src := boltedJoint(t)
	rep := joints.Run(src, prover(t, src))
	if len(rep.Joints) != 2 || len(rep.Skipped) != 0 {
		t.Fatalf("found %d joints, skipped %v", len(rep.Joints), rep.Skipped)
	}
	if !rep.AllProven {
		t.Fatalf("every claim should prove: %q", verdicts(rep))
	}
	for _, j := range rep.Joints {
		if len(j.Claims) != 5 {
			t.Errorf("joint at line %d has %d claims, want 5", j.Line, len(j.Claims))
		}
	}
}

// TestBoltedJointBroken: each way a joint goes wrong in context is disproven
// by the claim that names it, on the upright joint, and the tilted joint still
// proves. A fault may break more than one claim: a grip told too long leaves
// the base untapped where the shank crosses it, which the fit claims see too.
func TestBoltedJointBroken(t *testing.T) {
	src := boltedJoint(t)
	upright := "seat1, dir1, d, grip, fit);"
	if !strings.Contains(src, upright) {
		t.Fatal("the example's upright joint call has changed; update this test's edits")
	}
	cases := []struct {
		name, from, to, claim string
	}{
		// Told a grip longer than the lid: the head would bear on air.
		{"grip too long", upright, "seat1, dir1, d, 9.0, fit);", "the clamped part fills the grip under the head"},
		// Told a grip shorter than the lid: the screw ends a millimetre and a
		// half early, inside the lid's thickness rather than the base's.
		{"grip too short", upright, "seat1, dir1, d, 6.0, fit);", "the threaded part holds 1.5 d of thread with a wall round it"},
		// A base 8 wide round an M6 thread: no wall.
		{"thin wall", "baseLocal = box([-15.0, -15.0, -30.0], [15.0, 15.0, 0.0 - grip]);",
			"baseLocal = box([-4.0, -15.0, -30.0], [4.0, 15.0, 0.0 - grip]);", "the threaded part holds 1.5 d of thread with a wall round it"},
		// A cover laid over the upright screw's head.
		{"collision", "return screw1, lid1, base1, screw2, lid2, base2;",
			"cover = box([-35.0, -10.0, 4.0], [-9.0, 10.0, 8.0]);\nreturn screw1, lid1, base1, screw2, lid2, base2, cover;",
			"screw clears the rest of the assembly"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			broken := strings.Replace(src, c.from, c.to, 1)
			if broken == src {
				t.Fatalf("edit %q did not apply", c.from)
			}
			rep := joints.Run(broken, prover(t, broken))
			if rep.AllProven || len(rep.Joints) == 0 {
				t.Fatalf("a broken joint proved: %q", verdicts(rep))
			}
			for _, cl := range rep.Joints[0].Claims {
				if cl.Name == c.claim && !cl.Disproven {
					t.Errorf("%q should be disproven: %q", c.claim, verdicts(rep)[0])
				}
			}
			for _, cl := range rep.Joints[1].Claims {
				if !cl.Proven {
					t.Errorf("the other joint is untouched, but %q did not prove: %q", cl.Name, verdicts(rep)[1])
				}
			}
		})
	}
}

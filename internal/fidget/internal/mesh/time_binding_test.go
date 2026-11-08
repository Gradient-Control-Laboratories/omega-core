package mesh

import (
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/internal/fidget/internal/convert"
	"gradientcontrol.com/omega/core/internal/fidget/internal/fidget"
)

// timeAtPoints evaluates a program's single return at the given points with
// the clock set to timeAt.
func timeAtPoints(t *testing.T, src string, timeAt float64, xs, ys, zs []float32) []float32 {
	t.Helper()
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(testPkgMgr))
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	var out []float32
	err = convert.ToTreesAtTime(res.Program, timeAt, func(trees []fidget.TreeIndex, _ []string) error {
		grads, err := fidget.EvalGrad(fidget.BuildCanvas(trees[0]), xs, ys, zs)
		if err != nil {
			return err
		}
		for _, s := range grads {
			out = append(out, s.Value)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	return out
}

// TestTimeIsNotBoundToY locks the fix for a build-order bug: the implied vars
// are created before BuildPackage clears the tree arena, so the tree indices
// they captured were stale. p survived by luck (x, y, z land back on the same
// indices); time's constant did not, and its index aliased the y tree — every
// `time` reference on the server evaluated as p.y, silently diverging from the
// viewer, which has a real clock uniform.
func TestTimeIsNotBoundToY(t *testing.T) {
	const src = "a = time + 0.0 * p.x;\nreturn a;\n"
	xs := []float32{7, 0, 0}
	ys := []float32{-3, 9, -1.5}
	zs := []float32{2, 0, 0}

	got := timeAtPoints(t, src, 0, xs, ys, zs)
	for i, v := range got {
		if v != 0 {
			t.Errorf("time at y=%v evaluated to %v, want 0 (aliasing p.y?)", ys[i], v)
		}
	}

	// And the clock is honored when set, not merely pinned to zero.
	got = timeAtPoints(t, src, 2.5, xs, ys, zs)
	for i, v := range got {
		if v != 2.5 {
			t.Errorf("time at y=%v evaluated to %v, want 2.5", ys[i], v)
		}
	}
}

// TestImpliedPointSurvivesArenaClear guards the other half of the same hazard:
// p must still be x, y, z after the clear.
func TestImpliedPointSurvivesArenaClear(t *testing.T) {
	const src = "a = p.x * 100.0 + p.y * 10.0 + p.z;\nreturn a;\n"
	got := timeAtPoints(t, src, 0, []float32{1, 2}, []float32{3, 4}, []float32{5, 6})
	want := []float32{135, 246}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("p encoding at sample %d = %v, want %v", i, got[i], want[i])
		}
	}
}

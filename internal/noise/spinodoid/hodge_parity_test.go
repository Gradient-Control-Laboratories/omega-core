package spinodoid_test

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/internal/noise/spinodoid"
)

// What hodge's Gabor spinodoid is, measured against hodge itself.
//
// The Go baker reproduces hodge-release's spinodoid_gabor: the same SplitMix64
// kernel hash, the same numpy-drawn direction pool, the same jitter, envelope
// and normalization, with its scale and orientation drivers sampled at the
// kernel centres. That claim is only worth anything if something checks it, so
// these tests run hodge's own spinodoid.py and compare.
//
// hodge is found as repoHodgeDir says: HODGE_PATH, the sibling hodge-release
// checkout omega's other hodge tooling uses, by default. They skip unless a
// Python with numpy and scipy can import it (hodge's level threshold needs
// scipy): HODGE_PYTHON names the interpreter — the hodge conda env's python —
// and "python" then "python3" are tried after it.
//
// Parity holds inside the table's range. hodge generates kernels on demand for
// any cell, while the table holds [-TableCells/2, TableCells/2) and wraps, so
// the points below stay well inside it.

const (
	parityPeriod      = 6.0
	parityPacketWidth = 3.0
	parityImpulses    = 8
	parityPool        = 4096
	parityLevelBias   = 0.5
)

// The pool is drawn by rejection from numpy's normals, so it either matches
// bit for bit or not at all; a stray ULP would mean a different draw.
const poolTolerance = 1e-15

// Both sides sum the same kernels in f64 but not in the same order, so the
// field agrees to rounding, not to the bit: measured 2e-14 on the cases below,
// at a period of 6. The tolerance is on the pseudo-distance, which is in
// length units, and is loose on purpose — any real divergence (a different
// hash, jitter, pool draw or envelope) moves the field by order 1, not by
// ULPs.
const fieldTolerance = 1e-9

func TestWavesMatchHodge(t *testing.T) {
	py := hodgePython(t)
	thetaDeg := []float64{90, 60, 0}

	got, err := spinodoid.Waves(thetaDeg, parityPool, 0, nil)
	if err != nil {
		t.Fatalf("Waves: %v", err)
	}
	var want [][3]float64
	py.run(t, fmt.Sprintf(`
import json, numpy as np, spinodoid as sp
dirs, _phases = sp.spinodoid_waves(%s, %d, seed=0)
print(json.dumps(np.asarray(dirs).tolist()))
`, pyList(thetaDeg), parityPool), &want)

	if len(got) != len(want) {
		t.Fatalf("pool has %d directions, hodge drew %d", len(got), len(want))
	}
	worst := 0.0
	for i := range got {
		for a := range 3 {
			worst = math.Max(worst, math.Abs(got[i][a]-want[i][a]))
		}
	}
	if worst > poolTolerance {
		t.Errorf("direction pool differs from hodge's by %g, want <= %g", worst, poolTolerance)
	}
	t.Logf("pool of %d directions, worst component differs by %g (theta %v deg)", len(got), worst, thetaDeg)
}

// A stationary field: no drivers at all, which is what hodge's None means.
func TestBakeMatchesHodgeStationary(t *testing.T) {
	checkFieldParity(t, spinodoid.Params{
		Theta:       []float64{math.Pi / 2, math.Pi / 2, math.Pi / 2},
		Period:      parityPeriod,
		PacketWidth: parityPacketWidth,
		Impulses:    parityImpulses,
		Pool:        parityPool,
	}, "", "")
}

// An anisotropic field: cones of different half-angles, which is what makes
// the pool draw and the per-kernel direction choice matter.
func TestBakeMatchesHodgeAnisotropic(t *testing.T) {
	checkFieldParity(t, spinodoid.Params{
		Theta:       []float64{math.Pi / 12, math.Pi / 6, math.Pi / 2},
		Period:      parityPeriod,
		PacketWidth: parityPacketWidth,
		Impulses:    parityImpulses,
		Pool:        parityPool,
		Seed:        7,
	}, "", "")
}

// A graded field: the scale driver is read once at every kernel centre, so it
// only agrees with hodge if the centres do — jitter included.
func TestBakeMatchesHodgeGradedScale(t *testing.T) {
	scale := func(centers [][3]float64) ([]float64, error) {
		out := make([]float64, len(centers))
		for i, c := range centers {
			out[i] = 1 + 0.3*math.Sin(c[0]/25)
		}
		return out, nil
	}
	checkFieldParity(t, spinodoid.Params{
		Theta:       []float64{math.Pi / 2, math.Pi / 2, math.Pi / 2},
		Period:      parityPeriod,
		PacketWidth: parityPacketWidth,
		Impulses:    parityImpulses,
		Pool:        parityPool,
		Scale:       scale,
	}, "", "lambda pts: 1 + 0.3 * np.sin(pts[:, 0] / 25)")
}

// A rotated field: the orientation driver turns each kernel's direction, so
// this checks the rotation is applied as hodge applies it (R . n, per kernel).
func TestBakeMatchesHodgeGradedOrientation(t *testing.T) {
	rotate := func(centers [][3]float64) ([][3][3]float64, error) {
		out := make([][3][3]float64, len(centers))
		for i, c := range centers {
			s, co := math.Sincos(c[1] / 40)
			out[i] = [3][3]float64{{co, -s, 0}, {s, co, 0}, {0, 0, 1}}
		}
		return out, nil
	}
	checkFieldParity(t, spinodoid.Params{
		Theta:       []float64{math.Pi / 12, math.Pi / 12, math.Pi / 2},
		Period:      parityPeriod,
		PacketWidth: parityPacketWidth,
		Impulses:    parityImpulses,
		Pool:        parityPool,
		Orientation: rotate,
	}, `lambda pts: _zrot(pts[:, 1] / 40)`, "")
}

// checkFieldParity bakes p, evaluates the pseudo-distance at the probe points,
// and compares against hodge's spinodoid_gabor with the same arguments.
// orientationPy and scalePy are the Python spellings of p's drivers, or "".
func checkFieldParity(t *testing.T, p spinodoid.Params, orientationPy, scalePy string) {
	t.Helper()
	py := hodgePython(t)

	level, err := spinodoid.Level(parityLevelBias)
	if err != nil {
		t.Fatalf("Level: %v", err)
	}
	table, err := spinodoid.HodgeBake(p)
	if err != nil {
		t.Fatalf("HodgeBake: %v", err)
	}
	points := probePoints()
	got := make([]float64, len(points))
	for i, q := range points {
		got[i] = table.Distance(q, level, true)
	}

	thetaDeg := make([]float64, len(p.Theta))
	for i, th := range p.Theta {
		thetaDeg[i] = th * 180 / math.Pi
	}
	var want []float64
	py.run(t, fmt.Sprintf(`
import json, numpy as np, spinodoid as sp

def _zrot(angle):
    s, c = np.sin(angle), np.cos(angle)
    z, o = np.zeros_like(s), np.ones_like(s)
    return np.stack([np.stack([c, -s, z], -1),
                     np.stack([s, c, z], -1),
                     np.stack([z, z, o], -1)], -2)

pts = np.asarray(%s, dtype=np.float64)
dirs, _phases = sp.spinodoid_waves(%s, %d, seed=%d)
d = sp.spinodoid_gabor(pts, %.17g, dirs, density=%.17g, packet_width=%.17g,
                       impulses_per_cell=%d, seed=%d, normalize="local",
                       orientation=%s, scale=%s)
print(json.dumps(np.asarray(d).tolist()))
`, pyPoints(points), pyList(thetaDeg), p.Pool, p.Seed,
		p.Period, parityLevelBias, p.PacketWidth, p.Impulses, p.Seed,
		orNone(orientationPy), orNone(scalePy)), &want)

	if len(want) != len(got) {
		t.Fatalf("hodge returned %d values for %d points", len(want), len(got))
	}
	worst, at := 0.0, 0
	for i := range got {
		if e := math.Abs(got[i] - want[i]); e > worst {
			worst, at = e, i
		}
	}
	if worst > fieldTolerance {
		t.Errorf("pseudo-distance differs from hodge's by %g at %v (got %.12f, hodge %.12f), want <= %g",
			worst, points[at], got[at], want[at], fieldTolerance)
	}
	t.Logf("%d points, worst difference %g (period %g, cell %g)",
		len(points), worst, p.Period, table.CellSize)
}

// probePoints are the points both sides evaluate: a few on the kernel grid's
// own lattice, where a cell boundary falls, and a spread of off-grid ones,
// including negative coordinates (the cell floor and the table's origin offset
// are easy to get wrong by one there).
func probePoints() [][3]float64 {
	pts := [][3]float64{
		{0, 0, 0},
		{1e-9, 1e-9, 1e-9},
		{-1e-9, -1e-9, -1e-9},
		{3.5, -7.25, 11.125},
		{-21.79317, 21.79317, 0}, // one cell either way at period 6
		{43.5, -43.5, 43.5},
		{-60, 17, 5.5},
		{59.9, -59.9, 59.9},
	}
	// A deterministic spread, pseudo-random but fixed: no RNG to seed and no
	// dependence on map order.
	for i := range 24 {
		f := float64(i)
		pts = append(pts, [3]float64{
			40 * math.Sin(1.7*f+0.3),
			40 * math.Sin(2.3*f+1.1),
			40 * math.Sin(3.1*f+2.7),
		})
	}
	return pts
}

// hodgeRunner is the interpreter and environment that can import the repo's
// hodge, or a skip.
type hodgeRunner struct {
	exe     string
	hodgeIn string
}

func (py hodgeRunner) run(t *testing.T, script string, out any) {
	t.Helper()
	cmd := exec.Command(py.exe, "-c", script)
	cmd.Env = append(os.Environ(),
		"PYTHONPATH="+py.hodgeIn,
		"PYTHONDONTWRITEBYTECODE=1")
	stdout, err := cmd.Output()
	if err != nil {
		var stderr string
		if ee, ok := err.(*exec.ExitError); ok {
			stderr = string(ee.Stderr)
		}
		t.Fatalf("hodge python failed: %v\n%s", err, stderr)
	}
	// numpy's OpenMP can print warnings before the JSON; take the last line.
	lines := strings.Split(strings.TrimSpace(string(stdout)), "\n")
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), out); err != nil {
		t.Fatalf("hodge python output: %v\n%s", err, stdout)
	}
}

func hodgePython(t *testing.T) hodgeRunner {
	t.Helper()
	dir := repoHodgeDir(t)
	candidates := []string{"python", "python3"}
	if named := os.Getenv("HODGE_PYTHON"); named != "" {
		candidates = append([]string{named}, candidates...)
	}
	var tried []string
	for _, exe := range candidates {
		path, err := exec.LookPath(exe)
		if err != nil {
			tried = append(tried, exe+": not on PATH")
			continue
		}
		// hodge's level threshold reaches for scipy's erfinv, and it does so
		// lazily, so importing the module is not enough to know it is there.
		probe := exec.Command(path, "-c", "import numpy, spinodoid; from scipy.special import erfinv")
		probe.Env = append(os.Environ(), "PYTHONPATH="+dir, "PYTHONDONTWRITEBYTECODE=1")
		if out, err := probe.CombinedOutput(); err != nil {
			tried = append(tried, fmt.Sprintf("%s: %v (%s)", path, err, firstLine(out)))
			continue
		}
		return hodgeRunner{exe: path, hodgeIn: dir}
	}
	t.Skipf("no Python here can import hodge's spinodoid with numpy and scipy, so this parity "+
		"check cannot run; set HODGE_PYTHON to one that can. Tried %v", tried)
	return hodgeRunner{}
}

// repoHodgeDir finds hodge's spinodoid.py: HODGE_PATH if set (omega's
// convention for its sibling hodge-release checkout), else a hodge/ directory
// above the package (a monorepo that carries hodge as source), else the
// default sibling ~/github/hodge-release.
func repoHodgeDir(t *testing.T) string {
	t.Helper()
	has := func(dir string) bool {
		_, err := os.Stat(filepath.Join(dir, "spinodoid.py"))
		return err == nil
	}
	if named := os.Getenv("HODGE_PATH"); named != "" {
		if !has(named) {
			t.Fatalf("HODGE_PATH=%s has no spinodoid.py", named)
		}
		return named
	}
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	for {
		if candidate := filepath.Join(dir, "hodge"); has(candidate) {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if home, err := os.UserHomeDir(); err == nil {
		if sibling := filepath.Join(home, "github", "hodge-release"); has(sibling) {
			return sibling
		}
	}
	t.Skip("no hodge checkout: set HODGE_PATH to a hodge-release checkout")
	return ""
}

func pyList(vs []float64) string {
	parts := make([]string, len(vs))
	for i, v := range vs {
		parts[i] = fmt.Sprintf("%.17g", v)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func pyPoints(pts [][3]float64) string {
	rows := make([]string, len(pts))
	for i, p := range pts {
		rows[i] = fmt.Sprintf("[%.17g, %.17g, %.17g]", p[0], p[1], p[2])
	}
	return "[" + strings.Join(rows, ", ") + "]"
}

func orNone(expr string) string {
	if expr == "" {
		return "None"
	}
	return expr
}

// firstLine is the useful part of a failed probe's output.
func firstLine(out []byte) string {
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	return lines[len(lines)-1]
}

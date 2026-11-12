//go:build cgo

// Package libboxtest proves the boxes the mechanical libraries declare. It is a
// package of its own so its interval proofs run in their own test binary,
// under their own timeout, apart from gsleval's.
package libboxtest

import (
	"context"
	"fmt"
	"io/fs"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"gradientcontrol.com/omega/core/eval"
	"gradientcontrol.com/omega/core/internal/source"
)

// packageRoot is the GSL package tree.
const packageRoot = "../../../stdlib"

// unboundedAxis is what an axis .box cannot bound reads as: builtin's nothing.
const unboundedAxis = 1e5

func finiteAxis(v float64) bool { return math.Abs(v) < unboundedAxis/2 }

// boxEval evaluates each expression's .box in the document and returns
// [lo.x, lo.y, lo.z, hi.x, hi.y, hi.z] for each.
func boxEval(t *testing.T, e *eval.Evaluator, code string, exprs []string) [][6]float64 {
	t.Helper()
	queries := make([]string, len(exprs))
	for i, x := range exprs {
		queries[i] = "(" + x + ").box"
	}
	groups, err := e.EvalGsl(context.Background(), nil, code, "", queries, [][3]float64{{0, 0, 0}})
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	out := make([][6]float64, len(exprs))
	for i := range exprs {
		if len(groups[i].Samples) != 6 {
			t.Fatalf("%s.box: %d components, want 6", exprs[i], len(groups[i].Samples))
		}
		for k := range out[i] {
			out[i][k] = float64(groups[i].Samples[k][0].Value)
		}
	}
	return out
}

// declaredBoxLibraries are the packages whose parts declare their own boxes.
var declaredBoxLibraries = []string{
	"mechanical/fastener/fastener.gsl",
	"mechanical/driver/driver.gsl",
	"mechanical/gears/gears.gsl",
}

// boxDeclaration matches one name.box = [...]; statement, over as many lines
// as it takes.
var boxDeclaration = regexp.MustCompile(`(?m)^[ \t]*\w+\.box = [^;]*;[ \t]*\n`)

// TestBoundingBoxLibraryDeclared checks every box the mechanical libraries
// declare. A library's box is not cropped to, only read, so nothing but this
// test holds it to the part it bounds. Each part is proved against a copy of
// the libraries with the declarations taken out: outside its box, grown by a
// hair less than boxPad, the bare part has no material; and inside, within a
// few percent of each face, it has some, so the box is tight as well as sound.
//
// Where interval arithmetic cannot decide a face — a bevel gear's teeth wrap
// their phase across the -x axis, and no interval straddling the seam decides —
// the band outside that face is sampled densely instead, which finds material
// there if the box is too small but does not prove it is not.
func TestBoundingBoxLibraryDeclared(t *testing.T) {
	if testing.Short() {
		t.Skip("interval proofs over every declared library box")
	}
	const pad = 0.01  // builtin's boxPad
	const hair = 2e-3 // less than pad, so a proof covers the crop
	ctx := context.Background()

	bare := eval.New(source.NewLocalManager(undeclaredPackages(t)), nil)
	declared := eval.New(source.NewLocalManager(packageRoot), nil)

	code := `
softBody, softPad = fastener.setScrewSoftTip(6.0, 10.0, 3.0, 5.5, 2.0);
dogBody, dogTip = fastener.setScrewSoftPoint(5.0, 8.0, 2.5, 4.0, 3.5, 2.5);
drillBody, drillTip = fastener.setScrewDrillPoint(4.0, 12.0, 2.0, 3.0, 4.0, radians(118.0), 1.0);
spurA, spurB = gears.gearPair(20.0, 26.0, 0.4, 0.26, 3.5, 0.5, 3.0, 4.0, 2.5, 0.1, 1.2, 1.6, 0.25, 0.025, 0.3);
bevelA, bevelB = gears.bevelPair(15.0, 25.0, 2.0, 0.35, 1.5707963267948966, 0.6, 9.0, 1.0, 1.0, 0.1, 11.0, 16.0, 3.0, 0.4, 6.0, 8.0, 0.25, 0.05, 0.2);
skewA, skewB = gears.bevelPair(18.0, 24.0, 1.5, 0.35, 1.2, 0.0, 8.0, 0.5, 0.5, 0.1, 8.0, 10.0, 2.0, 0.3, 4.0, 5.0, 0.25, 0.05, 0.0);
return sphere(1);
`
	exprs := []string{
		// heads, including a button taller than it is wide and a flange
		// narrower than its crown
		"fastener.cylindricalHead(8.5, 5.0)",
		"fastener.countersunkHead(11.2, 5.0)",
		"fastener.countersunkHead(8.0, 3.0, radians(100.0))",
		"fastener.buttonHead(9.5, 2.75)",
		"fastener.buttonHead(6.0, 4.5)",
		"fastener.panHead(8.0, 3.1)",
		"fastener.panHead(8.0, 3.1, 5.0)",
		"fastener.hexHead(8.0, 3.5)",
		"fastener.squareHead(8.0, 3.5)",
		"fastener.pentHead(8.0, 3.5)",
		"fastener.flangeHead(8.0, 4.0, 11.0, 1.2)",
		"fastener.flangeHead(12.0, 4.0, 8.0, 1.2)",
		// shanks, including a thread longer than its shank and a thread
		// wider than its shoulder
		"fastener.plainShank(5.0, 20.0)",
		"fastener.threadedShank(5.0, 0.8, 20.0)",
		"fastener.sinShank(5.0, 0.8, 20.0)",
		"fastener.partialShank(8.0, 1.25, 40.0, 22.0)",
		"fastener.partialShank(8.0, 1.25, 10.0, 22.0)",
		"fastener.shoulderShank(8.0, 12.0, 6.0, 1.0, 10.0)",
		"fastener.shoulderShank(5.0, 12.0, 6.0, 1.0, 10.0)",
		// ends that add material
		"fastener.dogPoint(4.0, -10.0, 3.0)",
		"fastener.softTip(4.9, -10.0, 2.0)",
		"fastener.selfDrillPoint(4.1, -16.0, 5.0, radians(118.0), 1.2)",
		// the set screws' two halves
		"softBody", "softPad", "dogBody", "dogTip", "drillBody", "drillTip",
		// standard screws across the catalog, and a size it has no row for
		"fastener.shcs(1.6, 5.0)",
		"fastener.shcs(3.0, 12.0)",
		"fastener.shcs(8.0, 30.0)",
		"fastener.shcs(12.0, 40.0)",
		"fastener.shcsThreaded(4.0, 10.0)",
		"fastener.shcsThreadedSin(4.0, 10.0)",
		"fastener.shcsTorx(5.0, 12.0)",
		"fastener.countersunk(3.0, 10.0)",
		"fastener.countersunk(8.0, 25.0)",
		"fastener.countersunkThreaded(4.0, 12.0)",
		"fastener.panPhillips(2.0, 6.0)",
		"fastener.panPhillips(6.0, 16.0)",
		"fastener.panPhillipsGauge(2.0, 1.0)",
		"fastener.panPhillipsGauge(6.0, 3.0)",
		"fastener.panPozidriv(4.0, 10.0)",
		"fastener.panTorx(3.0, 8.0)",
		"fastener.panTorx(5.0, 12.0)",
		"fastener.button(3.0, 8.0)",
		"fastener.button(5.0, 10.0)",
		"fastener.button(16.0, 40.0)",
		"fastener.hexScrew(2.0, 8.0)",
		"fastener.hexScrew(12.0, 40.0)",
		"fastener.hexScrew(24.0, 60.0)",
		"fastener.hexBolt(8.0, 40.0)",
		// placed, shank down and shank up
		"fastener.shcsSeated(4.0, 12.0, [8.0, 26.75, -4.5], [0.0, 0.0, -1.0])",
		"fastener.shcsSeated(4.0, 12.0, [8.0, 26.75, -4.5], [0.0, 0.0, 1.0])",
		"fastener.shcsSeatedThreaded(3.0, 8.0, [1.0, 2.0, 3.0], [0.0, 0.0, 1.0])",
		"fastener.countersunkSeated(5.0, 16.0, [0.0, 5.0, 2.0], [0.0, 0.0, -1.0])",
		"fastener.countersunkSeated(5.0, 16.0, [0.0, 5.0, 2.0], [0.0, 0.0, 1.0])",
		// nuts, washers and pins
		"fastener.nut(5.0)",
		"fastener.nut(12.0)",
		"fastener.washer(5.0)",
		"fastener.washer(10.0)",
		"fastener.dowel(4.0, 16.0)",
		// driver tips, shanks and screwdrivers
		"driver.phillipsTip(0.0)",
		"driver.phillipsTip(2.0)",
		"driver.phillipsTip(4.0)",
		"driver.pozidrivTip(1.0)",
		"driver.pozidrivTip(3.0)",
		"driver.torxTip(10.0, 4.0)",
		"driver.torxTip(25.0, 6.0)",
		"driver.hexTip(3.0, 8.0)",
		"driver.robertsonTip(2.0, 6.0)",
		"driver.slotTip(0.6, 5.5, 8.0)",
		"driver.slotTip(1.2, 8.0, 1.0)",
		"driver.insertShank(4.37, 0.8)",
		"driver.powerShank(4.37, 0.8)",
		"driver.microShank(2.78, 0.5)",
		"driver.screwdriverBlade(4.37, 0.8, 6.0, 100.0)",
		"driver.screwdriverHandle(129.37, 36.0, 118.0)",
		// gears: a helical one with a standoff, one without, one whose
		// standoff is shorter than its face, the pairs placed and turned
		"gears.gear(20.0, 0.4, 0.26, 0.0, 3.5, 0.5, 3.0, 2.5, 0.1, 1.2, 0.25, 0.025)",
		"gears.gear(12.0, 1.0, 0.35, 0.0, 4.0, 0.0, 0.0, 0.0, 0.0, 2.0, 0.25, 0.02)",
		"gears.gear(16.0, 0.8, 0.35, 0.1, 6.0, 0.0, 5.0, 1.0, 0.3, 2.0, 0.25, 0.02)",
		"spurA", "spurB",
		"gears.rack(1.0, 0.35, 0.0, 0.1, 20.0, 4.0, 3.0, 0.25, 0.02)",
		"gears.bevelGear(20.0, 2.0, 0.35, 0.7853981633974483, 0.0, 0.0, 12.0, 1.0, 1.0, 0.0, 14.0, 4.0, 0.4, 8.0, 0.25, 0.05)",
		"gears.bevelGear(20.0, 2.0, 0.35, 0.7853981633974483, 0.0, 0.0, 12.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 8.0, 0.25, 0.05)",
		"bevelA", "bevelB", "skewA", "skewB",
	}
	// A size the catalog has no row for has no head, so its box's top is
	// loose; a bevel gear turned onto its shaft has the box around its own box
	// turned, loose wherever the gear does not fill a corner. These only have
	// to be sound.
	soundOnly := map[string]bool{
		"fastener.shcs(18.0, 40.0)": true,
		"bevelB":                    true,
		"skewB":                     true,
	}
	exprs = append(exprs, "fastener.shcs(18.0, 40.0)")

	boxes := boxEval(t, declared, code, exprs)
	for i, expr := range exprs {
		bx := boxes[i]
		var box [6]float64
		for k := 0; k < 3; k++ {
			box[k], box[k+3] = bx[k]+pad, bx[k+3]-pad
		}
		size := 0.0
		for k := 0; k < 3; k++ {
			if !finiteAxis(box[k]) || !finiteAxis(box[k+3]) {
				t.Errorf("%s: declared box %v is unbounded", expr, bx)
				size = -1
				break
			}
			size = math.Max(size, box[k+3]-box[k])
		}
		if size < 0 {
			continue
		}
		reach := size + 5
		var region [6]float64
		for k := 0; k < 3; k++ {
			region[k], region[k+3] = box[k]-reach, box[k+3]+reach
		}
		for k := 0; k < 3; k++ {
			band := 0.05*(box[k+3]-box[k]) + 0.05
			for side := 0; side < 2; side++ {
				// Outside this face: no material.
				out := region
				in := box
				if side == 0 {
					out[k+3] = box[k] - hair
					in[k+3] = box[k] + band
				} else {
					out[k] = box[k+3] + hair
					in[k] = box[k+3] - band
				}
				res, err := bare.Certify(ctx, nil, code, "", expr, out, 0, 1e-4)
				if err != nil {
					t.Fatalf("%s: certify outside %v: %v", expr, out, err)
				}
				face := fmt.Sprintf("%d%s", k, [2]string{"-", "+"}[side])
				if res.Disproven {
					t.Errorf("%s: material outside its box %v, at %v", expr, box, res.CounterBox)
				} else if res.Indeterminate {
					if at, found := materialNear(t, bare, code, expr, box, k, side, hair); found {
						t.Errorf("%s: material outside face %s of its box %v, at %v", expr, face, box, at)
					} else {
						t.Logf("%s: face %s sampled, not proven (%s)", expr, face, res.Reason)
					}
				}
				// Just inside it: some.
				if soundOnly[expr] {
					continue
				}
				res, err = bare.Certify(ctx, nil, code, "", expr, in, 0, 1e-4)
				if err != nil {
					t.Fatalf("%s: certify inside %v: %v", expr, in, err)
				}
				if res.Indeterminate {
					t.Logf("%s: whether anything is within %.3g of face %s is undecided (%s)",
						expr, band, face, res.Reason)
				} else if !res.Disproven {
					t.Errorf("%s: box %v is loose, nothing within %.3g of face %s",
						expr, box, band, face)
				}
			}
		}
	}
}

// materialNear samples the band just outside face (k, side) of box, as wide
// as a tenth of the box and running a tenth past it on the other axes, and
// returns the first point where the part has material.
func materialNear(t *testing.T, e *eval.Evaluator, code, expr string, box [6]float64, k, side int, hair float64) ([3]float64, bool) {
	t.Helper()
	rng := rand.New(rand.NewSource(int64(k*2 + side + 1)))
	pts := make([][3]float64, 40000)
	for i := range pts {
		for j := 0; j < 3; j++ {
			lo, hi := box[j], box[j+3]
			span := hi - lo
			if j == k {
				depth := hair + rng.Float64()*(0.1*span+0.5)
				if side == 0 {
					pts[i][j] = lo - depth
				} else {
					pts[i][j] = hi + depth
				}
				continue
			}
			pts[i][j] = lo - 0.1*span + rng.Float64()*1.2*span
		}
	}
	groups, err := e.EvalGsl(context.Background(), nil, code, "", []string{expr}, pts)
	if err != nil {
		t.Fatalf("%s: sample: %v", expr, err)
	}
	for i, s := range groups[0].Samples[0] {
		if s.Value <= 0 {
			return pts[i], true
		}
	}
	return [3]float64{}, false
}

// undeclaredPackages copies the package tree to a temporary directory with the
// libraries' box declarations taken out, and returns its root.
func undeclaredPackages(t *testing.T) string {
	t.Helper()
	src := packageRoot
	dst := t.TempDir()
	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, lib := range declaredBoxLibraries {
			if filepath.ToSlash(rel) == lib {
				stripped := boxDeclaration.ReplaceAll(body, nil)
				if len(stripped) == len(body) {
					t.Fatalf("%s declares no box", lib)
				}
				body = stripped
			}
		}
		return os.WriteFile(target, body, 0o644)
	})
	if err != nil {
		t.Fatalf("copy packages: %v", err)
	}
	return dst
}

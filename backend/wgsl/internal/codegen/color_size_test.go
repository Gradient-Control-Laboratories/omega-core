package codegen_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

// The colour field is generated, not written, so nothing in the source says how
// big it is allowed to get. Two defects lived there, and neither changed a
// single evaluated colour — no existing test could see either one:
//
//   - nearestColor accumulated its running nearest distance with a select that
//     named the accumulator twice, so the expression doubled on every coloured
//     shape folded in. Eight spheres synthesized 392 functions.
//   - On a sharp union the blend weights are one-hot, so they sum to 1, the
//     leftover weight max(1-sum, 0) is identically 0, and the nearest-colour
//     fill it multiplies is dead. It was emitted in full anyway, carrying a
//     distance remap per unit.
//
// Both are in common/program/property/color.go; the measurements have to be
// here, because only the backend knows how big the result got.
//
// These tests assert SHAPE, not size. A byte count is a golden in disguise, and
// moves whenever codegen changes. Every bound below is a comparison between two
// documents that differ in exactly one thing — the number of shapes, or whether
// one shape has a colour — so unrelated codegen work moves both sides together.

// colouredUnion is n coloured spheres and ONE UNCOLOURED one in a flat sharp
// union: the cheapest document that runs the colour blend once per shape AND
// the nearest-colour fill across all of them.
//
// The uncoloured sphere is what makes it a test of the fill. With every sphere
// coloured the weights sum to 1, the fill is elided as dead, and nearestValue —
// where the exponential defect lived — is never called: the document this used
// to build passed with that defect put back.
func colouredUnion(n int) []byte {
	var b strings.Builder
	names := make([]string, n+1)
	for i := 0; i < n; i++ {
		names[i] = fmt.Sprintf("s%d", i)
		fmt.Fprintf(&b, "%s = sphere([%d.0, 0.0, 0.0], 1.0);\n", names[i], 3*i)
		fmt.Fprintf(&b, "%s.color = [%0.1f, %0.1f, 0.5];\n",
			names[i], float64(i%3)/2, float64(i%2))
	}
	names[n] = "bare"
	fmt.Fprintf(&b, "bare = sphere([0.0, 3.0, 0.0], 1.0);\n")
	fmt.Fprintf(&b, "return min(%s);\n", strings.Join(names, ", "))
	return []byte(b.String())
}

func countFns(wgsl string) int {
	n := strings.Count(wgsl, "\nfn ")
	if strings.HasPrefix(wgsl, "fn ") {
		n++
	}
	return n
}

// remapCalls counts CALLS to synthesized remap helpers — every occurrence of a
// helper's name followed by an argument list, less the definitions. The codegen
// shares byte-identical helpers, so a duplicated subexpression no longer shows
// up as more definitions; it still shows up as more calls, one per copy.
func remapCalls(wgsl string) int {
	n := 0
	for _, line := range strings.Split(wgsl, "\n") {
		n += len(remapCallName.FindAllStringIndex(line, -1))
		if strings.HasPrefix(line, "fn gsl_") && remapCallName.MatchString(line) {
			n-- // the definition's own name
		}
	}
	return n
}

var remapCallName = regexp.MustCompile(`\bgsl_\w*remap(?:_\d+)?\(`)

// scalarRemaps counts synthesized remaps that return a number rather than a
// position or a colour. In the colour field those are the per-unit DISTANCES,
// and a per-unit distance has exactly one purpose: ranking the coloured shapes
// to find the nearest one for the fill.
func scalarRemaps(wgsl string) int {
	n := 0
	for _, line := range strings.Split(wgsl, "\n") {
		if strings.HasPrefix(line, "fn gsl_remap") && strings.HasSuffix(line, "-> f32 {") {
			n++
		}
	}
	return n
}

// TestColorUnionGrowthIsNotExponential pins the growth ORDER of the colour
// field in the number of coloured shapes.
//
// Measured against the compiler before the fix, synthesized functions went
// 28 at 4 spheres, 392 at 8, 6156 at 12 — doubling per shape. So the step from
// 4 shapes to 8 is 14x when the defect is present, and the bound below sits
// well above the healthy ratio.
//
// Remap CALLS are the measure, not bytes and no longer function definitions.
// Bytes separate the two states far less cleanly, because the duplicated
// subexpressions are hoisted into helpers rather than written out in line. The
// definitions used to be the sharp instrument, but the codegen now shares
// byte-identical helpers, and the copies the defect makes are exactly that:
// with the defect put back, definitions went 10 -> 16 and passed. The copies
// still call their helpers, one call each — 26 -> 390 with the defect, 15x,
// against 17 -> 51 healthy, 3x.
func TestColorUnionGrowthIsNotExponential(t *testing.T) {
	const small, large = 4, 8

	resSmall, err := runTestBuildOne(t, "", colouredUnion(small))
	if err != nil {
		t.Fatalf("build %d spheres: %v", small, err)
	}
	resLarge, err := runTestBuildOne(t, "", colouredUnion(large))
	if err != nil {
		t.Fatalf("build %d spheres: %v", large, err)
	}

	fSmall, fLarge := countFns(resSmall.WGSL), countFns(resLarge.WGSL)
	cSmall, cLarge := remapCalls(resSmall.WGSL), remapCalls(resLarge.WGSL)
	t.Logf("%d spheres: %d fns, %d remap calls, %d bytes", small, fSmall, cSmall, len(resSmall.WGSL))
	t.Logf("%d spheres: %d fns, %d remap calls, %d bytes", large, fLarge, cLarge, len(resLarge.WGSL))

	// Doubling per shape is 2**(large-small) = 16x. Quadratic is 4x.
	const maxRatio = 6
	if cLarge > maxRatio*cSmall {
		t.Errorf("doubling the coloured shapes multiplied the remap calls by %.1fx (%d -> %d).\n"+
			"The colour field is growing exponentially in the number of coloured shapes, which happens "+
			"when an accumulator is named more than once in its own update.",
			float64(cLarge)/float64(cSmall), cSmall, cLarge)
	}
}

// TestColorSharpUnionOmitsDeadFill pins the other half. On a sharp union of
// coloured units the weights are one-hot — min passes its whole derivative to
// the active branch and none to the others — so they sum to 1 everywhere, and
// the nearest-colour fill is multiplied by zero at every point.
//
// The fill is NOT dead as soon as any leaf is uncoloured: that leaf takes the
// derivative wherever it is nearest, the sum over the coloured units is 0
// there, and the fill is the only thing painting that region. So the two
// documents here are the same but for one colour assignment, and they must
// generate measurably different colour fields. A predicate that answers "no
// fill needed" unconditionally passes the first half and fails the second.
//
// Before the fix the all-coloured document emitted 7 per-unit distance remaps
// and 15 functions; after, 0 and 5.
func TestColorSharpUnionOmitsDeadFill(t *testing.T) {
	const allColoured = `a = sphere([0.0, 0.0, 0.0], 1.0);
a.color = [1.0, 0.0, 0.0];
b = sphere([3.0, 0.0, 0.0], 1.0);
b.color = [0.0, 0.0, 1.0];
c = sphere([6.0, 0.0, 0.0], 1.0);
c.color = [0.0, 1.0, 0.0];
return min(a, b, c);`

	// The same document less one colour. c is now an uncoloured leaf, so the
	// region where c is nearest has no colour of its own and has to inherit one.
	const oneBare = `a = sphere([0.0, 0.0, 0.0], 1.0);
a.color = [1.0, 0.0, 0.0];
b = sphere([3.0, 0.0, 0.0], 1.0);
b.color = [0.0, 0.0, 1.0];
c = sphere([6.0, 0.0, 0.0], 1.0);
return min(a, b, c);`

	resAll, err := runTestBuildOne(t, "", []byte(allColoured))
	if err != nil {
		t.Fatalf("build all-coloured: %v", err)
	}
	resBare, err := runTestBuildOne(t, "", []byte(oneBare))
	if err != nil {
		t.Fatalf("build one-uncoloured: %v", err)
	}

	fAll, fBare := countFns(resAll.WGSL), countFns(resBare.WGSL)
	dAll, dBare := scalarRemaps(resAll.WGSL), scalarRemaps(resBare.WGSL)
	t.Logf("all coloured:   %d fns, %d per-unit distances, %d bytes", fAll, dAll, len(resAll.WGSL))
	t.Logf("one uncoloured: %d fns, %d per-unit distances, %d bytes", fBare, dBare, len(resBare.WGSL))

	if dAll != 0 {
		t.Errorf("a sharp union of three COLOURED spheres emitted %d per-unit distance remaps.\n"+
			"Those distances rank the shapes to pick a fill colour, and on a sharp union every weight "+
			"the fill could occupy is already taken: it is multiplied by zero.", dAll)
	}
	if dBare == 0 {
		t.Errorf("the same union with one shape left UNCOLOURED emitted no per-unit distance remaps, " +
			"so nothing is finding the nearest coloured shape.\nThat region has no colour of its own " +
			"and would be left unpainted — eliding the fill there is a behaviour change, not a saving.")
	}
	if fAll >= fBare {
		t.Errorf("three coloured shapes generated %d functions; the same union with one shape "+
			"uncoloured generated %d. The all-coloured case needs no fill and must generate strictly less.",
			fAll, fBare)
	}
}

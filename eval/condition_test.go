//go:build cgo

package eval_test

import (
	"io"
	"math"
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/eval"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/source"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/require"
)

// coated is a 20 mm cube under a coat t mm thick (a slider, 0.1 as written),
// and rough a 20 mm cube from x = 20 to 40 whose material has 0.2 mm of relief
// and no build.
const coated = `
@range(0.05, 0.2) t = 0.1;
blank = boxCentered(vector.zero, [20.0, 20.0, 20.0]);
blank.material = material.make(0.5, 0, 1.5);
coat = blank;
coat.material = material.withBuild(material.make(0.3, 0, 1.5), t);
rough = boxCentered([30.0, 0, 0], [20.0, 20.0, 20.0]);
rough.material = material.withDisplacement(material.make(0.5, 0, 1.5), 0.2);
return coat, rough;
`

// conditioned is an evaluator that reads its fields at c.
func conditioned(t *testing.T, c eval.Condition) *eval.Evaluator {
	t.Helper()
	e := eval.New(&source.MockManager{}, nil)
	require.NoError(t, e.SetCondition(c))
	return e
}

// values is each expression's value at each point.
func values(t *testing.T, e *eval.Evaluator, exprs []string, points [][3]float64) [][]float64 {
	t.Helper()
	groups, err := e.EvalGsl(t.Context(), nil, coated, "", exprs, points)
	require.NoError(t, err)
	out := make([][]float64, len(groups))
	for i, g := range groups {
		for _, s := range g.Samples[0] {
			out[i] = append(out[i], float64(s.Value))
		}
	}
	return out
}

func nearly(t *testing.T, what string, got, want, tol float64) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Errorf("%s: %g, want %g within %g", what, got, want, tol)
	}
}

// As built nothing moves; nominal moves the coat out by its build and leaves
// relief alone; as made moves both.
func TestCondition_MovesBuildsAndRelief(t *testing.T) {
	points := [][3]float64{{10.05, 0, 0}, {40.05, 0, 0}}
	for _, c := range []struct {
		condition   eval.Condition
		coat, rough float64
	}{
		{eval.AsBuilt, 0.05, 0.05},
		{eval.Nominal, -0.05, 0.05},
		{eval.AsMade, -0.05, -0.15},
		// The most material: builds on, relief at its peaks, as nominal until spreads exist.
		{eval.MaxMaterial, -0.05, 0.05},
	} {
		got := values(t, conditioned(t, c.condition), []string{"coat", "rough"}, points)
		nearly(t, string(c.condition)+" coat", got[0][0], c.coat, 1e-5)
		nearly(t, string(c.condition)+" rough", got[1][1], c.rough, 1e-5)
	}
}

// A shape returned bare beside a return that moves it inside its booleans is
// moved by its own build everywhere, inside the blank and just above it too,
// not by the material its layer shows there.
func TestCondition_BareReturnBesideItsLayer(t *testing.T) {
	got := values(t, conditioned(t, eval.Nominal), []string{"coat", "max(coat, 9.95 - length(p))"},
		[][3]float64{{9.9, 0, 0}, {10.0005, 0, 0}})
	nearly(t, "coat inside the blank", got[0][0], -0.2, 1e-5)
	nearly(t, "coat just above the blank", got[0][1], -0.0995, 1e-5)
}

// A claim that holds as built fails nominal once the coat is on, and a claim
// over the coat's thickness names the thicknesses it fails for.
func TestCondition_Certify(t *testing.T) {
	near := [6]float64{10.02, -1, -1, 10.08, 1, 1}
	r, err := conditioned(t, eval.AsBuilt).Certify(t.Context(), nil, coated, "", "coat", near, 0, 0)
	require.NoError(t, err)
	require.True(t, r.Proven, "as built: %+v", r)
	r, err = conditioned(t, eval.Nominal).Certify(t.Context(), nil, coated, "", "coat", near, 0, 0)
	require.NoError(t, err)
	require.True(t, r.Disproven, "nominal: %+v", r)

	nominal := conditioned(t, eval.Nominal)
	thickness := map[string][2]float64{"t": {0.05, 0.2}}
	r, err = nominal.CertifyParams(t.Context(), nil, coated, "", "coat", [6]float64{10.3, -1, -1, 10.4, 1, 1}, 0, 0, thickness)
	require.NoError(t, err)
	require.True(t, r.Proven, "clear of every coat: %+v", r)
	r, err = nominal.CertifyParams(t.Context(), nil, coated, "", "coat", [6]float64{10.15, -1, -1, 10.18, 1, 1}, 0, 0, thickness)
	require.NoError(t, err)
	require.True(t, r.Disproven, "inside the thick coats: %+v", r)
	if lo := r.CounterParams["t"][0]; lo < 0.15 {
		t.Errorf("fails only for coats past 0.15 mm, named from %g", lo)
	}
}

// journal is a 25 k5 journal, its seat turned (0.1 mm feed, 0.4 mm nose, cusps
// 3.14 µm deep) and machined 5 µm under its maximum-material size, then plated
// 5 µm over the turning.
const journal = `
rMax = (25.0 + tolerance.shaftUpper(25.0, tolerance.k5)) / 2.0;
steelMat, steelColor = metal.steel();
fMat, fColor, fHeight, fSlope = machined.turned(steelMat, steelColor, [0, 0, 0], [0, 0, 1], 0.1, 0.4);
r = sqrt(p.x * p.x + p.y * p.y);
seat = r - (rMax - 0.005);
seat.material = fMat;
plated = seat;
plated.material = material.withBuild(fMat, 0.005);
envelope = max(rMax - r, plated);
return plated;
`

// As made, the plated journal stays inside its maximum-material envelope, and
// the proof costs what the field written out by hand costs: the plating over
// the turning reads the turning's relief once, not once beneath it and twice
// in its increment.
func TestCondition_PlatingOverTurningProves(t *testing.T) {
	patch := [6]float64{12.495, -0.05, 0, 12.512, 0.05, 0.5}
	r, err := conditioned(t, eval.AsMade).Certify(t.Context(), nil, journal, "", "envelope", patch, -0.0001, 0)
	require.NoError(t, err)
	require.True(t, r.Proven, "%+v", r)
	if r.Evals > 20000 {
		t.Errorf("proven in %d interval evaluations; written out by hand it takes about 5,000", r.Evals)
	}
}

// The proven box, a surface point and the extremum see the coat's build.
func TestCondition_EncloseSurfaceExtremum(t *testing.T) {
	region := [6]float64{-12, -12, -12, 12, 12, 12}
	for _, c := range []struct {
		condition eval.Condition
		face      float64
	}{
		{eval.AsBuilt, 10},
		{eval.Nominal, 10.1},
	} {
		e := conditioned(t, c.condition)
		b, err := e.Enclose(t.Context(), nil, coated, "", "coat", region, 32)
		require.NoError(t, err)
		require.True(t, b.Proven, "%s: %+v", c.condition, b)
		nearly(t, string(c.condition)+" proven box", b.BboxMax[0], c.face, 0.01)

		s, err := e.SurfacePoint(t.Context(), nil, coated, "", "coat", [][3]float64{{12, 0, 0}}, 0, 0)
		require.NoError(t, err)
		nearly(t, string(c.condition)+" surface point", s[0].Point[0], c.face, 1e-3)

		// The least value, -face at the centre, lies between the bound and the least value sampled.
		x, err := e.Extremum(t.Context(), nil, coated, "", "coat", [6]float64{-11, -11, -11, 11, 11, 11}, "min", 0, 1)
		require.NoError(t, err)
		if x.MinLower > -c.face || x.Min < -c.face {
			t.Errorf("%s: least value %g not in [%g, %g]", c.condition, -c.face, x.MinLower, x.Min)
		}
		if c.condition == eval.Nominal && x.Min >= -10 {
			t.Errorf("nominal: least value sampled %g, no lower than the bare cube's -10", x.Min)
		}
	}
}

// The volume nominal brackets the cube offset by its build, a^3 + 6 a^2 b +
// 3 pi a b^2 + 4/3 pi b^3, and not the bare cube.
func TestCondition_Volume(t *testing.T) {
	const a, b = 20.0, 0.1
	built := a*a*a + 6*a*a*b + 3*math.Pi*a*b*b + 4.0/3.0*math.Pi*b*b*b
	v, err := conditioned(t, eval.Nominal).VolumeOf(t.Context(), nil, coated, "", "coat", [6]float64{-11, -11, -11, 11, 11, 11}, 0.01)
	require.NoError(t, err)
	if v.Lower > built || v.Upper < built || v.Lower <= a*a*a {
		t.Errorf("%+v does not bracket %g above the bare %g", v, built, a*a*a)
	}
}

// Mesh meshes at the condition as its Coatings and Relief flags mesh.
func TestCondition_Mesh(t *testing.T) {
	const a, b = 20.0, 0.1
	built := a*a*a + 6*a*a*b + 3*math.Pi*a*b*b + 4.0/3.0*math.Pi*b*b*b
	doc := strings.Replace(coated, "return coat, rough;", "return coat;", 1)
	box := [6]float64{-11, -11, -11, 11, 11, 11}
	for _, c := range []struct {
		condition eval.Condition
		volume    float64
	}{
		{eval.AsBuilt, a * a * a},
		{eval.Nominal, built},
	} {
		r, err := conditioned(t, c.condition).Mesh(t.Context(), nil, doc, "", io.Discard,
			eval.MeshOptions{Subdivision: 8, Algorithm: "mdc", Region: &box})
		require.NoError(t, err)
		require.Equal(t, 1, len(r.Bodies))
		if rel := math.Abs(r.Bodies[0].VolumeMm3-c.volume) / c.volume; rel > 1e-4 {
			t.Errorf("%s: meshed volume %.3f, want %.3f", c.condition, r.Bodies[0].VolumeMm3, c.volume)
		}
	}
}

// The minimum material condition is a band shifted by a relief depth, not a
// field, and a name that is no condition is refused.
func TestCondition_Refused(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	for _, c := range []eval.Condition{eval.MinMaterial, "as-drawn"} {
		if err := e.SetCondition(c); err == nil {
			t.Errorf("%s: accepted", c)
		}
	}
	c, err := eval.ParseCondition("")
	require.NoError(t, err)
	require.Equal(t, eval.AsBuilt, c)
}

// fit is a block with a 25 mm bore, zinc plated, beside a shaft: the bore's
// radius as bored, the zinc's thickness and the shaft's radius are working
// dimensions a tolerance would range.
const fit = `
boreR = 12.51175;
zinc = 0.0065;
shaftR = 12.49325;
steelMat, steelColor = metal.steel();
block = boxCentered([0, 0, 0], [40.0, 40.0, 20.0]);
holed = max(block, 0.0 - (length(p.xy) - boreR));
holed.material = steelMat;
plated = holed;
plated.material = material.withBuild(steelMat, zinc);
shaft = max(length(p.xy) - shaftR, abs(p.z) - 15.0);
return plated, shaft;
`

// A query's derivative is taken of the field its condition reads: as made the
// plated surface moves out by its zinc on every face, as built not at all,
// and the bore's radius moves only the bore.
func TestCondition_DerivativesSeeTheCoats(t *testing.T) {
	points := [][3]float64{{12.505, 0, 0}, {20.01, 0, 0}, {0, 20.01, 5}, {13, 0, 10.01}}
	for _, c := range []struct {
		condition eval.Condition
		zinc      float64
	}{
		{eval.AsBuilt, 0},
		{eval.AsMade, -1},
	} {
		groups, err := conditioned(t, c.condition).EvalGsl(t.Context(), nil, fit, "", []string{"plated.d(zinc)", "plated.d(boreR)", "plated.dp"}, points)
		require.NoError(t, err)
		for i := range points {
			nearly(t, string(c.condition)+" d(zinc)", float64(groups[0].Samples[0][i].Value), c.zinc, 1e-5)
		}
		for i, want := range []float64{1, 0, 0, 0} {
			nearly(t, string(c.condition)+" d(boreR)", float64(groups[1].Samples[0][i].Value), want, 1e-5)
		}
		// The gradient points out of the solid: into the bore at the first
		// point, along +x at the second, +y and +z at the others.
		for i, want := range [][3]float64{{-1, 0, 0}, {1, 0, 0}, {0, 1, 0}, {0, 0, 1}} {
			for k := 0; k < 3; k++ {
				nearly(t, string(c.condition)+" dp", float64(groups[2].Samples[k][i].Value), want[k], 1e-5)
			}
		}
	}
}

// The gradient of a value read from a material whose rows only partly vary.
func TestDerivative_ThroughAMaterialRecord(t *testing.T) {
	doc := fit + "\nwavy = material.withDisplacement(steelMat, 0.001 * sin(p.z));\n"
	groups, err := eval.New(&source.MockManager{}, nil).EvalGsl(t.Context(), nil, doc, "", []string{"material.displacement(wavy).dp"}, [][3]float64{{0, 0, 0.5}})
	require.NoError(t, err)
	require.Equal(t, 3, len(groups[0].Samples))
	nearly(t, "d/dx", float64(groups[0].Samples[0][0].Value), 0, 1e-7)
	nearly(t, "d/dz", float64(groups[0].Samples[2][0].Value), 0.001*math.Cos(0.5), 1e-7)
}

// fitToleranced is fit with working tolerances on its three dimensions.
var fitToleranced = strings.NewReplacer(
	"boreR = 12.51175;", "boreR = 12.51175;\nboreR.tolerance = tolerance.symmetric(0.003);",
	"zinc = 0.0065;", "zinc = 0.0065;\nzinc.tolerance = tolerance.limits(-0.0015, 0.0015);",
	"shaftR = 12.49325;", "shaftR = 12.49325;\nshaftR.tolerance = tolerance.symmetric(0.0016);",
).Replace(fit)

// A working tolerance reads back as the offsets it was set to.
func TestTolerance_ReadsItsLimits(t *testing.T) {
	groups, err := eval.New(&source.MockManager{}, nil).EvalGsl(t.Context(), nil, fitToleranced, "", []string{"zinc.tolerance", "boreR + boreR.tolerance[0]"}, [][3]float64{{0, 0, 0}})
	require.NoError(t, err)
	nearly(t, "zinc low", float64(groups[0].Samples[0][0].Value), -0.0015, 1e-9)
	nearly(t, "zinc high", float64(groups[0].Samples[1][0].Value), 0.0015, 1e-9)
	nearly(t, "the bore's least radius", float64(groups[1].Samples[0][0].Value), 12.50875, 1e-6)
}

// Certified over its tolerances as made, the plated bore's clearance to the
// shaft falls below 3.2 µm, at the corner where the bore is bored smallest
// and the zinc and the shaft run largest.
func TestTolerance_CertifyFindsTheWorstCorner(t *testing.T) {
	gap := [6]float64{12.48, -0.2, -5, 12.53, 0.2, 5}
	r, err := conditioned(t, eval.AsMade).CertifyOverTolerances(t.Context(), nil, fitToleranced, "", "max(plated, shaft)", gap, 0.0032, 0)
	require.NoError(t, err)
	require.True(t, r.Disproven, "%+v", r)
	if r.CounterParams["boreR"][1] > 0.1 || r.CounterParams["zinc"][0] < 0.9 || r.CounterParams["shaftR"][0] < 0.9 {
		t.Errorf("disproven away from the worst corner: %+v", r.CounterParams)
	}
}

// The journal, its 5 ± 1 µm of zinc a working tolerance: machined 6.5 µm
// under, every zinc keeps it inside its envelope; 5 µm under, the thickest
// do not.
func TestTolerance_CertifyProvesThePlatedJournal(t *testing.T) {
	doc := strings.NewReplacer(
		"seat = r - (rMax - 0.005);", "seat = r - (rMax - cut);",
		"plated.material = material.withBuild(fMat, 0.005);", "zinc = 0.005;\nzinc.tolerance = tolerance.symmetric(0.001);\nplated.material = material.withBuild(fMat, zinc);",
	).Replace("cut = 0.0065;\n" + journal)
	patch := [6]float64{12.495, -0.05, 0, 12.512, 0.05, 0.5}
	asMade := conditioned(t, eval.AsMade)
	r, err := asMade.CertifyOverTolerances(t.Context(), nil, doc, "", "envelope", patch, -0.0001, 0)
	require.NoError(t, err)
	require.True(t, r.Proven, "6.5 µm under: %+v", r)
	r, err = asMade.CertifyOverTolerances(t.Context(), nil, doc, "cut = 0.005;", "envelope", patch, -0.0001, 0)
	require.NoError(t, err)
	require.True(t, r.Disproven, "5 µm under: %+v", r)
	if r.CounterParams["zinc"][0] < 0.5 {
		t.Errorf("fails for zinc thinner than 5.1 µm: %+v", r.CounterParams)
	}
}

// A document that declares no tolerance has none to certify over.
func TestTolerance_CertifyNeedsATolerance(t *testing.T) {
	_, err := eval.New(&source.MockManager{}, nil).CertifyOverTolerances(t.Context(), nil, fit, "", "plated", [6]float64{-1, -1, -1, 1, 1, 1}, 0, 0)
	if err == nil || !strings.Contains(err.Error(), "declares no tolerances") {
		t.Errorf("got %v", err)
	}
}

// A value that declares no tolerance reads its own, to first order from the
// working tolerances it is made from: a sum of two, and a shape's, which
// differs face by face and, as made, counts the coat.
func TestTolerance_Propagates(t *testing.T) {
	points := [][3]float64{{12.505, 0, 0}, {20.01, 0, 0}, {0, 20.01, 5}}
	for _, c := range []struct {
		condition    eval.Condition
		bore, others float64
	}{
		{eval.AsBuilt, 0.003, 0},
		{eval.AsMade, 0.0045, 0.0015},
	} {
		groups, err := conditioned(t, c.condition).EvalGsl(t.Context(), nil, fitToleranced, "finished = boreR - zinc;",
			[]string{"finished.tolerance", "plated.tolerance", "shaft.tolerance"}, points)
		require.NoError(t, err)
		for i := range points {
			nearly(t, "finished low", float64(groups[0].Samples[0][i].Value), -0.0045, 1e-7)
			nearly(t, "finished high", float64(groups[0].Samples[1][i].Value), 0.0045, 1e-7)
			want := c.others
			if i == 0 {
				want = c.bore
			}
			nearly(t, string(c.condition)+" plated low", float64(groups[1].Samples[0][i].Value), -want, 1e-7)
			nearly(t, string(c.condition)+" plated high", float64(groups[1].Samples[1][i].Value), want, 1e-7)
			nearly(t, "shaft high", float64(groups[2].Samples[1][i].Value), 0.0016, 1e-7)
		}
	}

	// Nothing toleranced, nothing varies.
	groups, err := eval.New(&source.MockManager{}, nil).EvalGsl(t.Context(), nil, fit, "", []string{"plated.tolerance"}, points[:1])
	require.NoError(t, err)
	nearly(t, "untoleranced low", float64(groups[0].Samples[0][0].Value), 0, 0)
	nearly(t, "untoleranced high", float64(groups[0].Samples[1][0].Value), 0, 0)
}

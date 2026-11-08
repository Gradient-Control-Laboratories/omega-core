package gsleval

import (
	"context"
	"fmt"
	"math"
	"sort"

	fidgetapi "gradientcontrol.com/omega/core/internal/fidget"
	"gradientcontrol.com/omega/core/internal/fidget/lib"
	"gradientcontrol.com/omega/core/internal/slider"
	parser "gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

// CertifyResult is the outcome of a sound interval-arithmetic proof that a
// scalar expression is at least a threshold over an entire region. Unlike
// Extremum (Lipschitz branch-and-bound, whose certificate is only as sound as
// the caller's Lipschitz constant), Certify needs no constant: interval
// arithmetic bounds each sub-box's true range directly, so a verdict of Proven
// or Disproven is sound to floating-point precision.
//
// Exactly one of Proven / Disproven / Indeterminate is set. Proven: expr >=
// threshold everywhere in the region. Disproven: a sub-box (CounterBox) where
// expr < threshold everywhere is proven (CounterUpper < threshold is its
// certified upper bound) — a concrete counterexample region. Indeterminate:
// the search hit its budget or tolerance with sub-boxes still straddling the
// threshold (WorstBox / WorstLower is the undecided sub-box closest to
// violating).
type CertifyResult struct {
	Expr      string
	Threshold float64
	Tolerance float64
	Evals     int

	// RegionLower/RegionUpper is the interval bound over the whole region — a
	// one-shot bracket, before any subdivision.
	RegionLower float64
	RegionUpper float64

	Proven        bool
	Disproven     bool
	Indeterminate bool
	// Reason says WHY an Indeterminate verdict stopped short, because two
	// different failures wear it and call for opposite repairs: ReasonBudget
	// (the evaluation budget ran out with cells still straddling — the field
	// grazes the threshold too closely to resolve at this scale; open the
	// margin) versus ReasonTolerance (every straddling cell reached the
	// tolerance — the margin is real but thinner than the tolerance can
	// see; lower the tolerance). ReasonPool is the frontier outgrowing the
	// cell pool. Empty when Proven or Disproven.
	Reason string

	// Set when Disproven: a box where expr < threshold is proven.
	CounterBox   [6]float64
	CounterUpper float64

	// Set when Indeterminate: of the boxes left undecided when the search
	// stopped, the one with the least lower bound. A box whose children the
	// budget never reached is reported whole, with its own bound.
	WorstBox   [6]float64
	WorstLower float64

	// Params records the parameter ranges a parametric proof held over —
	// nil for a point proof. See CertifyParams.
	Params map[string][2]float64
	// CounterParams, set with a parametric Disproven, is the parameter
	// sub-range over which the counterexample box fails for EVERY value —
	// "fails when rad is in [6.4, 6.5]". WorstParams accompanies a
	// parametric Indeterminate's WorstBox the same way.
	CounterParams map[string][2]float64
	WorstParams   map[string][2]float64
}

// The Reason values an Indeterminate CertifyResult can carry.
const (
	ReasonBudget    = "budget"
	ReasonPool      = "pool"
	ReasonTolerance = "tolerance"
)

const (
	// certChunk bounds how many boxes are interval-evaluated per FFI call, to
	// cap the flat transfer buffer.
	certChunk = 4096
	// certEvalBudget bounds total interval evaluations. Interval width shrinks
	// linearly with box size, so a field with margin proves far below this; the
	// budget only bites fields that genuinely graze the threshold.
	certEvalBudget = 200000
	// certMaxActive caps the live box pool; overflow ends the search as
	// Indeterminate rather than exhausting memory.
	certMaxActive = 200000
	// certMaxActiveCap bounds the pool even under a raised budget — cells
	// cost ~120 bytes, so this holds the pool near a GB. A joint proof
	// over space × motion × sliders legitimately carries a frontier in
	// the millions while the graze band resolves.
	certMaxActiveCap = 8000000
)

// budgetAndPool resolves the interval-eval budget (the evaluator's override
// or the default) and the frontier pool cap scaled to it.
func (e *Evaluator) budgetAndPool() (budget, pool int) {
	budget = certEvalBudget
	if e.certifyBudget > 0 {
		budget = e.certifyBudget
	}
	pool = certMaxActive
	if budget > pool {
		pool = budget
	}
	if pool > certMaxActiveCap {
		pool = certMaxActiveCap
	}
	return budget, pool
}

// Certify proves, soundly, whether expr >= threshold holds over the whole
// region by interval branch-and-bound: each box is bounded by oracle-aware
// interval arithmetic; a box whose lower bound clears the threshold is retired,
// a box whose upper bound falls below it is a proven counterexample, and a box
// that straddles is subdivided until it decides or shrinks below tolerance.
//
// tolerance <= 0 names NO tolerance, and then the budget alone ends the
// search: cells subdivide down to precisionFloor, where float32 boxes stop
// being boxes. A default tolerance used to sit at 1e-3 of the region
// diagonal, and it gave up on exactly the claims worth making — a driver bit
// seated in its recess touches the screw, so a proof that it goes no deeper
// than 0.01 needs cells finer than 0.01, and over an 18 mm region the default
// stopped at 0.018 with most of the budget unspent. It reported "lower the
// tolerance" to a caller who never set one.
func (e *Evaluator) Certify(ctx context.Context, manifest map[string]string, code, defs, expr string, region [6]float64, threshold, tolerance float64) (*CertifyResult, error) {
	if err := startQuery(ctx); err != nil {
		return nil, err
	}
	defer endQuery()
	if region == ([6]float64{}) {
		region = [6]float64{
			-defaultBoundsRegion, -defaultBoundsRegion, -defaultBoundsRegion,
			defaultBoundsRegion, defaultBoundsRegion, defaultBoundsRegion,
		}
	}
	for k := 0; k < 3; k++ {
		if region[k+3] <= region[k] {
			return nil, fmt.Errorf("certify region max must exceed min on every axis")
		}
	}
	if tolerance <= 0 {
		tolerance = precisionFloor(region)
	}

	prog, err := e.prepareQuery(ctx, manifest, code, defs, []string{expr})
	if err != nil {
		return nil, err
	}
	evalRow := func(boxes [][6]float32) ([]fidgetapi.IntervalSample, error) {
		ivals, err := lib.EvalInterval(prog, boxes)
		if err != nil {
			return nil, err
		}
		if len(ivals) != 1 {
			return nil, fmt.Errorf("certify requires a single scalar expression, got %d values", len(ivals))
		}
		return ivals[0], nil
	}
	return e.certifySearch(ctx, expr, region, threshold, tolerance, evalRow)
}

// CertifyParams is Certify with named document parameters ranging over
// intervals: the verdict is sound over the whole region AND every value of
// every listed parameter simultaneously. Proven here means "for all p in the
// box and all parameter values in range"; Disproven means a sub-box where the
// predicate fails for EVERY parameter value in range. A predicate that fails
// only for SOME parameter values subdivides down to them: the search bisects
// parameters alongside space, so the disproof names the failing sub-range
// (CounterParams).
//
// The reserved name "time" (program.TimeVarName) ranges the implicit
// animation clock rather than a slider: a claim proven with time in params
// holds at every instant of the range — a motion envelope. It subdivides
// like any other parameter, so a failure narrows to the failing time span.
func (e *Evaluator) CertifyParams(ctx context.Context, manifest map[string]string, code, defs, expr string, region [6]float64, threshold, tolerance float64, params map[string][2]float64) (*CertifyResult, error) {
	if len(params) == 0 {
		return e.Certify(ctx, manifest, code, defs, expr, region, threshold, tolerance)
	}
	if err := startQuery(ctx); err != nil {
		return nil, err
	}
	defer endQuery()
	for name, b := range params {
		if !(b[0] <= b[1]) {
			return nil, fmt.Errorf("parameter %q has an empty range [%g, %g]", name, b[0], b[1])
		}
	}
	// Validate the names against the DOCUMENT, not the built program. A
	// parameter this expression does not reference is legitimate — the claim
	// simply holds for every value of it — but dead-code elimination drops
	// such a declaration before the build, so by then an unreferenced slider
	// and a name that was never declared look identical. Only the document
	// text still tells them apart.
	if err := e.checkParamNames(code, params); err != nil {
		return nil, err
	}
	if region == ([6]float64{}) {
		region = [6]float64{
			-defaultBoundsRegion, -defaultBoundsRegion, -defaultBoundsRegion,
			defaultBoundsRegion, defaultBoundsRegion, defaultBoundsRegion,
		}
	}
	for k := 0; k < 3; k++ {
		if region[k+3] <= region[k] {
			return nil, fmt.Errorf("certify region max must exceed min on every axis")
		}
	}
	// The tolerance does two jobs here, and with none named they come apart.
	// It NORMALIZES the dimensions against one another — which of space or a
	// parameter to split next — and 1e-3 of the region diagonal is still the
	// scale that balance was tuned at. It also STOPS the search, and that is
	// the precision floor, as in Certify: the budget decides when to give up.
	floor := tolerance
	if tolerance <= 0 {
		dx, dy, dz := region[3]-region[0], region[4]-region[1], region[5]-region[2]
		tolerance = 1e-3 * math.Sqrt(dx*dx+dy*dy+dz*dz)
		floor = precisionFloor(region)
	}
	prog, err := e.prepareQuery(ctx, manifest, code, defs, []string{expr})
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(params))
	for n := range params {
		names = append(names, n)
	}
	sort.Strings(names)
	ranges := make([][2]float64, len(names))
	for i, n := range names {
		ranges[i] = params[n]
	}
	evalBoxes := func(spatial [][6]float32, bounds [][][2]float64) ([]fidgetapi.IntervalSample, error) {
		return lib.EvalIntervalParamBoxes(prog, names, spatial, bounds)
	}
	result, err := e.certifySearchParams(ctx, expr, region, threshold, tolerance, floor, names, ranges, evalBoxes)
	if err != nil {
		return nil, err
	}
	result.Params = params
	return result, nil
}

// estimateParamLeverage probes how strongly each parameter moves the field:
// the max over sampled spatial points of (field variation across the
// parameter's range) / (range width), other parameters held at their
// midpoints. Point boxes make the interval evaluator an exact point
// evaluator, so the whole probe is one cheap batch. An underestimate only
// costs search time (the parameter splits later); it never affects
// soundness, since every verdict still comes from interval bounds.
func (e *Evaluator) estimateParamLeverage(
	region [6]float64,
	ranges [][2]float64,
	evalBoxes func([][6]float32, [][][2]float64) ([]fidgetapi.IntervalSample, error),
	result *CertifyResult,
) []float64 {
	nParams := len(ranges)
	leverage := make([]float64, nParams)
	const paramSamples = 5
	// Region corners + center: enough spatial spread to catch a parameter
	// whose effect is localized (a moving part far from the region center).
	points := make([][3]float64, 0, 9)
	for _, x := range [2]float64{region[0], region[3]} {
		for _, y := range [2]float64{region[1], region[4]} {
			for _, z := range [2]float64{region[2], region[5]} {
				points = append(points, [3]float64{x, y, z})
			}
		}
	}
	points = append(points, [3]float64{
		(region[0] + region[3]) / 2, (region[1] + region[4]) / 2, (region[2] + region[5]) / 2,
	})

	mids := make([]float64, nParams)
	for j, r := range ranges {
		mids[j] = (r[0] + r[1]) / 2
	}
	var boxes [][6]float32
	var bounds [][][2]float64
	for j, r := range ranges {
		for s := 0; s < paramSamples; s++ {
			v := r[0] + (r[1]-r[0])*float64(s)/(paramSamples-1)
			row := make([][2]float64, nParams)
			for k := range row {
				row[k] = [2]float64{mids[k], mids[k]}
			}
			row[j] = [2]float64{v, v}
			for _, pt := range points {
				boxes = append(boxes, [6]float32{
					float32(pt[0]), float32(pt[1]), float32(pt[2]),
					float32(pt[0]), float32(pt[1]), float32(pt[2]),
				})
				bounds = append(bounds, row)
			}
		}
	}
	row, err := evalBoxes(boxes, bounds)
	if err != nil || len(row) != len(boxes) {
		return leverage // zeros: fall back to spatial-tolerance behavior
	}
	result.Evals += len(boxes)
	idx := 0
	for j, r := range ranges {
		width := math.Max(r[1]-r[0], 1e-12)
		perPoint := make([][2]float64, len(points))
		for i := range perPoint {
			perPoint[i] = [2]float64{math.Inf(1), math.Inf(-1)}
		}
		for s := 0; s < paramSamples; s++ {
			for i := range points {
				v := float64(row[idx].Lower)
				idx++
				if math.IsNaN(v) || math.IsInf(v, 0) {
					continue
				}
				if v < perPoint[i][0] {
					perPoint[i][0] = v
				}
				if v > perPoint[i][1] {
					perPoint[i][1] = v
				}
			}
		}
		for i := range perPoint {
			if v := (perPoint[i][1] - perPoint[i][0]) / width; v > leverage[j] {
				leverage[j] = v
			}
		}
	}
	return leverage
}

// paramCell is one node of the parametric search: a spatial box together
// with the parameter ranges it is being proven over.
type paramCell struct {
	s [6]float64
	p [][2]float64
}

// pendingParamCell is a cell awaiting evaluation with the straddling parent
// it was split from, which is all the search knows about it if the budget
// runs out first — see certCell.
type pendingParamCell struct {
	c        paramCell
	parent   paramCell
	parentLo float64
}

// certifySearchParams is branch-and-bound over space AND parameters: a
// straddling cell bisects its widest dimension by normalized width, whether
// that dimension is a spatial axis or a parameter — so a predicate failing
// only at some parameter values narrows to a counterexample IN PARAMETER
// SPACE ("fails when rad is in [6.4, 6.5]") instead of an indeterminate
// shrug. Widths are normalized to the geometric `tolerance` for space and to
// that tolerance divided by the parameter's leverage for a parameter (with a
// 1e-4-of-range floor so a tiny-range slider stays splittable). A cell stops
// splitting once every dimension is under the same measure taken at `floor`
// instead — equal to tolerance when the caller named one.
func (e *Evaluator) certifySearchParams(
	ctx context.Context,
	expr string,
	region [6]float64,
	threshold, tolerance, floor float64,
	names []string,
	ranges [][2]float64,
	evalBoxes func([][6]float32, [][][2]float64) ([]fidgetapi.IntervalSample, error),
) (*CertifyResult, error) {
	result := &CertifyResult{Expr: expr, Threshold: threshold, Tolerance: floor}
	nParams := len(names)
	// A parametric proof must exhaust a frontier that subdivides parameter
	// dimensions as well as space, so the budget scales with the parameter
	// count; interval evals on the VM path are microseconds, so even the
	// scaled budget is seconds of work.
	baseBudget, maxActive := e.budgetAndPool()
	evalBudget := baseBudget * (1 + 2*nParams)
	// A parameter resolves at the geometric tolerance divided by its
	// LEVERAGE — how many field units one unit of the parameter moves the
	// field, probed by point evaluation. A slider that IS a length (a
	// radius, a travel: leverage ~1) resolves at the spatial tolerance,
	// so a certificate does not need micro-slivers of its range — and
	// normalizing such a parameter finer than space sends the search into
	// a thousand-leaf fan-out that no budget survives. But a parameter
	// with geometric gain (the animation clock: one unit of `time` can
	// swing a crank pin through a whole revolution) must split finer, or
	// the search refines all of space against the full motion smear
	// before time ever bisects. The floors keep the pre-leverage behavior
	// as the minimum (leverage below 1 never coarsens a parameter) and a
	// tiny-range slider splittable at all.
	leverage := e.estimateParamLeverage(region, ranges, evalBoxes, result)
	paramTol := make([]float64, nParams)
	paramFloor := make([]float64, nParams)
	for i, r := range ranges {
		width := math.Max(r[1]-r[0], 1e-12)
		tol := tolerance / math.Max(leverage[i], 1)
		paramTol[i] = math.Max(tol, 1e-4*width)
		paramFloor[i] = paramTol[i] * floor / tolerance
	}
	paramsOf := func(p [][2]float64) map[string][2]float64 {
		m := make(map[string][2]float64, nParams)
		for i, n := range names {
			m[n] = p[i]
		}
		return m
	}

	// The worst cell is chosen from what is left UNDECIDED — see
	// certifySearch.
	worstLower := math.Inf(1)
	var worstCell paramCell
	undecided := func(c paramCell, lo float64) {
		if lo < worstLower {
			worstLower, worstCell = lo, c
		}
	}
	straddled := false
	firstEval := true

	pending := []pendingParamCell{{c: paramCell{s: region, p: ranges}, parentLo: math.Inf(-1)}}
	pending[0].parent = pending[0].c
	for len(pending) > 0 {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if len(pending) > maxActive || result.Evals >= evalBudget {
			straddled = true
			// A cut-short search overrides a tolerance reason: the verdict
			// was never allowed to resolve, whatever the cells looked like.
			result.Reason = ReasonPool
			if result.Evals >= evalBudget {
				result.Reason = ReasonBudget
			}
			for _, q := range pending {
				undecided(q.parent, q.parentLo)
			}
			break
		}
		n := len(pending)
		if n > certChunk {
			n = certChunk
		}
		batch := pending[:n]
		rest := pending[n:]

		boxes := make([][6]float32, len(batch))
		bounds := make([][][2]float64, len(batch))
		for i, q := range batch {
			c := q.c
			boxes[i] = [6]float32{
				float32(c.s[0]), float32(c.s[1]), float32(c.s[2]),
				float32(c.s[3]), float32(c.s[4]), float32(c.s[5]),
			}
			bounds[i] = c.p
		}
		row, err := evalBoxes(boxes, bounds)
		if err != nil {
			return nil, err
		}
		result.Evals += len(batch)

		// Straddlers refine nearest-to-disproof first: sorting by UPPER bound
		// and PREPENDING children makes the search drill toward a decisive
		// counterexample (a cell whose upper bound falls below the threshold)
		// instead of exhausting the frontier breadth-first. The lower bound
		// is the wrong key: inside overlapping material it saturates at the
		// deepest field value (every cell ties), while the upper bound keeps
		// tightening as cells shrink, so it discriminates at every depth.
		type straddler struct {
			c      paramCell
			lo, hi float64
		}
		var straddlers []straddler
		for i, q := range batch {
			c := q.c
			lo := float64(row[i].Lower)
			hi := float64(row[i].Upper)
			if firstEval {
				result.RegionLower, result.RegionUpper = lo, hi
				firstEval = false
			}
			if lo >= threshold {
				continue
			}
			if hi < threshold {
				result.Disproven = true
				result.CounterBox = c.s
				result.CounterUpper = hi
				result.CounterParams = paramsOf(c.p)
				return result, nil
			}
			straddlers = append(straddlers, straddler{c: c, lo: lo, hi: hi})
		}
		sort.Slice(straddlers, func(a, b int) bool { return straddlers[a].hi < straddlers[b].hi })
		var children []pendingParamCell
		for _, st := range straddlers {
			c := st.c
			// Bisect the widest dimension by normalized width, spatial axes
			// normalized to the tolerance and parameters to their
			// leverage-scaled paramTol, so both kinds of dimension exhaust
			// together.
			bestDim, bestW := -1, 1.0
			for k := 0; k < 3; k++ {
				if w := (c.s[k+3] - c.s[k]) / math.Max(tolerance, 1e-12); w > bestW {
					bestDim, bestW = k, w
				}
			}
			for j := 0; j < nParams; j++ {
				if w := (c.p[j][1] - c.p[j][0]) / paramTol[j]; w > bestW {
					bestDim, bestW = 3+j, w
				}
			}
			if bestDim < 0 && floor < tolerance {
				// Every dimension is at the tolerance, and no tolerance was
				// named: go on splitting the widest by the same measure,
				// down to the floor. Only the budget ends this search.
				bestW = 0
				for k := 0; k < 3; k++ {
					if width := c.s[k+3] - c.s[k]; width > floor {
						if w := width / math.Max(tolerance, 1e-12); w > bestW {
							bestDim, bestW = k, w
						}
					}
				}
				for j := 0; j < nParams; j++ {
					if width := c.p[j][1] - c.p[j][0]; width > paramFloor[j] {
						if w := width / paramTol[j]; w > bestW {
							bestDim, bestW = 3+j, w
						}
					}
				}
			}
			if bestDim < 0 {
				straddled = true // every dimension at the floor; undecided
				if result.Reason == "" {
					result.Reason = ReasonTolerance
				}
				undecided(c, st.lo)
				continue
			}
			a, b := c, c
			if bestDim < 3 {
				mid := (c.s[bestDim] + c.s[bestDim+3]) / 2
				a.s[bestDim+3] = mid
				b.s[bestDim] = mid
			} else {
				j := bestDim - 3
				a.p = append([][2]float64(nil), c.p...)
				b.p = append([][2]float64(nil), c.p...)
				mid := (c.p[j][0] + c.p[j][1]) / 2
				a.p[j] = [2]float64{c.p[j][0], mid}
				b.p[j] = [2]float64{mid, c.p[j][1]}
			}
			children = append(children,
				pendingParamCell{c: a, parent: c, parentLo: st.lo},
				pendingParamCell{c: b, parent: c, parentLo: st.lo})
		}
		pending = append(children, rest...)
	}

	switch {
	case !straddled && len(pending) == 0:
		result.Proven = true
	default:
		result.Indeterminate = true
		result.WorstBox = worstCell.s
		result.WorstLower = worstLower
		if worstCell.p != nil {
			result.WorstParams = paramsOf(worstCell.p)
		}
	}
	return result, nil
}

// SliderRanges parses the document and returns every declared slider's full
// [min, max] range — the parameter set CertifyOverSliders proves over. A
// document with no sliders is an error: "over every configuration" of a
// document with none is a point proof the caller should ask for directly.
func SliderRanges(code string) (map[string][2]float64, error) {
	res, err := parser.ParsePackage([]string{code})
	if err != nil {
		return nil, fmt.Errorf("parse document for sliders: %w", err)
	}
	sliders := slider.Extract(code, res.Package)
	if len(sliders) == 0 {
		return nil, fmt.Errorf("document declares no sliders (variables with .min and .max) to certify over")
	}
	params := make(map[string][2]float64, len(sliders))
	for _, s := range sliders {
		params[s.VariableName] = [2]float64{s.Min, s.Max}
	}
	return params, nil
}

// CertifyOverSliders is CertifyParams over every slider the document declares,
// each ranging its full declared [min, max]: "proven for every configuration
// of the document's parameters". It returns the ranges it used.
func (e *Evaluator) CertifyOverSliders(ctx context.Context, manifest map[string]string, code, defs, expr string, region [6]float64, threshold, tolerance float64) (*CertifyResult, error) {
	params, err := SliderRanges(code)
	if err != nil {
		return nil, err
	}
	return e.CertifyParams(ctx, manifest, code, defs, expr, region, threshold, tolerance, params)
}

// certParent is a straddling box the search split. Its children are kept as
// references to it, and if the budget runs out before a child is evaluated,
// the parent and its bound are all the search knows about that child.
type certParent struct {
	b  [6]float64
	lo float64
}

// certRef is a box awaiting evaluation: octant oct of parent, in
// subdivide8's order, or the parent's whole box when oct is wholeBox.
type certRef struct {
	parent int32
	oct    uint8
}

// wholeBox is the octant of the root, which has no parent to be split from.
const wholeBox = 8

func (r certRef) box(parents []certParent) [6]float64 {
	b := parents[r.parent].b
	if r.oct == wholeBox {
		return b
	}
	return octant(b, r.oct)
}

// octant is child o of box b in subdivide8's order — x halves outermost, z
// innermost, the upper half when the axis's bit is set.
func octant(b [6]float64, o uint8) [6]float64 {
	mx, my, mz := (b[0]+b[3])/2, (b[1]+b[4])/2, (b[2]+b[5])/2
	out := b
	if o&4 == 0 {
		out[3] = mx
	} else {
		out[0] = mx
	}
	if o&2 == 0 {
		out[4] = my
	} else {
		out[1] = my
	}
	if o&1 == 0 {
		out[5] = mz
	} else {
		out[2] = mz
	}
	return out
}

// compactParents drops the parents no pending box refers to any more and
// renumbers the references. A parent's children sit together in pending —
// appended as eight, and a chunk boundary only moves the front of the run —
// so one pass with the last index seen does it.
func compactParents(parents []certParent, pending []certRef) []certParent {
	fresh := make([]certParent, 0, len(pending)/8+2)
	last := int32(-1)
	for i := range pending {
		if pending[i].parent != last {
			last = pending[i].parent
			fresh = append(fresh, parents[last])
		}
		pending[i].parent = int32(len(fresh) - 1)
	}
	return fresh
}

// certifySearch is the branch-and-bound core shared by the point and
// parametric proofs: evalRow supplies each chunk's sound brackets.
//
// THE WORST BOX IS TAKEN FROM WHAT IS LEFT UNDECIDED, not from every box the
// search ever straddled. Interval bounds only tighten as boxes shrink, so the
// least lower bound ever seen is the ROOT's, every time: the region itself
// came back as the worst box, which reads as a search that never subdivided.
// What is undecided when the search stops is the cells that reached the
// tolerance still straddling and, if the budget ran out, the straddlers whose
// children were never evaluated.
func (e *Evaluator) certifySearch(ctx context.Context, expr string, region [6]float64, threshold, tolerance float64, evalRow func([][6]float32) ([]fidgetapi.IntervalSample, error)) (*CertifyResult, error) {
	result := &CertifyResult{Expr: expr, Threshold: threshold, Tolerance: tolerance}
	evalBudget, maxActive := e.budgetAndPool()
	// THE POOL MUST NOT END A SEARCH THE BUDGET WOULD CONTINUE. This search
	// is breadth-first, and a contact — a seated bit, a press fit — is a
	// surface of straddling cells whose count grows as area over cell size
	// squared, so a pool the size of the budget filled first: the seated PH2
	// stopped at 96 thousand evaluations of a 200 thousand budget. Every
	// pending box is the child of an evaluated one, so eight times the
	// budget is a pool the budget always reaches first, and at 8 bytes a box
	// it costs less than the old pool of whole boxes did.
	if pool := 8 * evalBudget; pool > maxActive {
		maxActive = min(pool, 8*certMaxActiveCap)
	}
	worstLower := math.Inf(1)
	var worstBox [6]float64
	undecided := func(b [6]float64, lo float64) {
		if lo < worstLower {
			worstLower, worstBox = lo, b
		}
	}
	straddled := false
	firstEval := true

	parents := []certParent{{b: region, lo: math.Inf(-1)}}
	pending := []certRef{{parent: 0, oct: wholeBox}}
	compactAt := 4096
	for len(pending) > 0 {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if len(pending) > maxActive || result.Evals >= evalBudget {
			straddled = true
			result.Reason = ReasonPool
			if result.Evals >= evalBudget {
				result.Reason = ReasonBudget
			}
			for _, r := range pending {
				undecided(parents[r.parent].b, parents[r.parent].lo)
			}
			break
		}
		// Evaluate a chunk from the front; subdivided straddlers go to the back.
		n := len(pending)
		if n > certChunk {
			n = certChunk
		}
		batch := pending[:n]
		rest := pending[n:]

		batchBoxes := make([][6]float64, len(batch))
		boxes := make([][6]float32, len(batch))
		for i, r := range batch {
			b := r.box(parents)
			batchBoxes[i] = b
			boxes[i] = [6]float32{
				float32(b[0]), float32(b[1]), float32(b[2]),
				float32(b[3]), float32(b[4]), float32(b[5]),
			}
		}
		row, err := evalRow(boxes)
		if err != nil {
			return nil, err
		}
		result.Evals += len(batch)

		var children []certRef
		for i, b := range batchBoxes {
			lo := float64(row[i].Lower)
			hi := float64(row[i].Upper)
			if firstEval {
				result.RegionLower, result.RegionUpper = lo, hi
				firstEval = false
			}
			if lo >= threshold {
				continue // proven expr >= threshold over this whole box
			}
			if hi < threshold {
				// Proven counterexample: expr < threshold everywhere in this box.
				result.Disproven = true
				result.CounterBox = b
				result.CounterUpper = hi
				return result, nil
			}
			// Straddles the threshold.
			if boxRadius(b) <= tolerance {
				straddled = true // too small to decide, does not provably violate
				if result.Reason == "" {
					result.Reason = ReasonTolerance
				}
				undecided(b, lo)
				continue
			}
			parents = append(parents, certParent{b: b, lo: lo})
			idx := int32(len(parents) - 1)
			for oct := uint8(0); oct < 8; oct++ {
				children = append(children, certRef{parent: idx, oct: oct})
			}
		}
		pending = append(rest, children...)
		if len(parents) > compactAt {
			parents = compactParents(parents, pending)
			compactAt = 2*len(parents) + 4096
		}
	}

	switch {
	case !straddled && len(pending) == 0:
		result.Proven = true
	default:
		result.Indeterminate = true
		result.WorstBox = worstBox
		result.WorstLower = worstLower
	}
	return result, nil
}

// precisionFloor is how small a box may get when the caller names no
// tolerance. Boxes reach the interval evaluator as float32, whose spacing is
// about 1.2e-7 of a coordinate's magnitude, so a box narrower than a few of
// those collapses to a point; 1e-6 of the region's scale — its diagonal or
// its farthest coordinate from the origin, whichever is larger — keeps every
// box a box. The budget stops the search long before this in practice.
func precisionFloor(region [6]float64) float64 {
	scale := math.Sqrt(
		(region[3]-region[0])*(region[3]-region[0]) +
			(region[4]-region[1])*(region[4]-region[1]) +
			(region[5]-region[2])*(region[5]-region[2]))
	for _, v := range region {
		scale = math.Max(scale, math.Abs(v))
	}
	return 1e-6 * scale
}

// boxRadius is half the diagonal of an axis-aligned box [xMin,yMin,zMin,xMax,yMax,zMax].
func boxRadius(b [6]float64) float64 {
	dx, dy, dz := b[3]-b[0], b[4]-b[1], b[5]-b[2]
	return 0.5 * math.Sqrt(dx*dx+dy*dy+dz*dz)
}

// subdivide8 splits a box into its eight octree children at the midpoints.
func subdivide8(b [6]float64) [][6]float64 {
	mx, my, mz := (b[0]+b[3])/2, (b[1]+b[4])/2, (b[2]+b[5])/2
	xs := [2][2]float64{{b[0], mx}, {mx, b[3]}}
	ys := [2][2]float64{{b[1], my}, {my, b[4]}}
	zs := [2][2]float64{{b[2], mz}, {mz, b[5]}}
	out := make([][6]float64, 0, 8)
	for _, x := range xs {
		for _, y := range ys {
			for _, z := range zs {
				out = append(out, [6]float64{x[0], y[0], z[0], x[1], y[1], z[1]})
			}
		}
	}
	return out
}

// checkParamNames rejects a requested parameter the document does not declare
// as a slider. The reserved clock name always resolves, so it is exempt.
func (e *Evaluator) checkParamNames(code string, params map[string][2]float64) error {
	var wanted []string
	for name := range params {
		if name != program.TimeVarName {
			wanted = append(wanted, name)
		}
	}
	if len(wanted) == 0 {
		return nil
	}
	sort.Strings(wanted)
	declared, err := SliderRanges(code)
	if err != nil {
		// A document with no sliders at all: report the first requested name
		// rather than the extraction error, which would read as unrelated.
		return fmt.Errorf("parameter %q is not a slider-style literal declaration in this document", wanted[0])
	}
	for _, name := range wanted {
		if _, ok := declared[name]; !ok {
			return fmt.Errorf("parameter %q is not a slider-style literal declaration in this document", name)
		}
	}
	return nil
}

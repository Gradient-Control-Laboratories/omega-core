//go:build cgo

package eval

import (
	"context"
	"fmt"
	"math"
	"sort"

	"gradientcontrol.com/omega/core/program"
)

const (
	// bbInitialGrid seeds the search with bbInitialGrid^3 cells so a
	// multimodal field cannot hide a basin from the first incumbent.
	bbInitialGrid = 16
	// bbSplitPerRound bounds how many best-bound cells subdivide each round
	// (8 children each), which caps a round's evaluation batch.
	bbSplitPerRound = 512
	// bbEvalBudget bounds one pass's total point evaluations. The search is
	// geometric once the active set localizes, so a healthy field converges
	// far below this; the budget only bites plateaus and bad Lipschitz
	// constants, which then report Converged false.
	bbEvalBudget = 150000
	// bbMaxActive caps the live cell pool. Overflow drops the worst-bound
	// cells but folds their bounds into the certificate, so the reported
	// interval stays sound (it just may fail to converge).
	bbMaxActive = 100000
	// bbRelTolerance accepts the certificate (in default-tolerance mode)
	// once the interval is within this fraction of the extremum's
	// magnitude. A minimizer plateau (a clearance sum is constant along the
	// closest-approach segment) forces the certificate to refine every
	// plateau cell; chasing an absolute tolerance there is pure waste when
	// the value itself converged long ago. The absolute floor still governs
	// near zero, and an explicit caller tolerance disables this entirely.
	bbRelTolerance = 0.01
)

// Extremum finds the global minimum and/or maximum (mode "min", "max",
// "both") of a scalar expression over the region by Lipschitz
// branch-and-bound: cells whose center value minus lipschitz*halfDiagonal
// cannot beat the incumbent are pruned, the rest subdivide. lipschitz <= 0
// defaults to 1 — correct for any single SDF and for min/max (CSG)
// combinations of SDFs; a sum of two SDFs needs 2. tolerance > 0 is honored
// exactly; tolerance <= 0 selects the adaptive default (see ExtremumResult).
func (e *Evaluator) Extremum(ctx context.Context, manifest map[string]string, code, defs, expr string, region [6]float64, mode string, tolerance, lipschitz float64) (*ExtremumResult, error) {
	if err := startQuery(ctx); err != nil {
		return nil, err
	}
	defer endQuery()
	switch mode {
	case "":
		mode = "min"
	case "min", "max", "both":
	default:
		return nil, fmt.Errorf("mode must be \"min\", \"max\", or \"both\", got %q", mode)
	}
	if region == ([6]float64{}) {
		region = [6]float64{
			-defaultBoundsRegion, -defaultBoundsRegion, -defaultBoundsRegion,
			defaultBoundsRegion, defaultBoundsRegion, defaultBoundsRegion,
		}
	}
	for k := 0; k < 3; k++ {
		if region[k+3] <= region[k] {
			return nil, fmt.Errorf("extremum region max must exceed min on every axis")
		}
	}
	if lipschitz <= 0 {
		lipschitz = 1
	}
	diag := math.Sqrt(
		(region[3]-region[0])*(region[3]-region[0]) +
			(region[4]-region[1])*(region[4]-region[1]) +
			(region[5]-region[2])*(region[5]-region[2]))
	// tolerance > 0 is the caller's exact contract; otherwise bbMinimize
	// applies the adaptive default with this absolute floor.
	absFloor := 1e-4 * diag
	reportedTolerance := tolerance
	if reportedTolerance <= 0 {
		reportedTolerance = absFloor
	}

	prog, err := e.prepareQuery(ctx, manifest, code, defs, []string{expr})
	if err != nil {
		return nil, err
	}

	result := &ExtremumResult{Mode: mode, Lipschitz: lipschitz, Tolerance: reportedTolerance}
	if mode == "min" || mode == "both" {
		pass, err := bbMinimize(ctx, prog, region, false, lipschitz, tolerance, absFloor)
		if err != nil {
			return nil, err
		}
		result.Min = pass.value
		result.MinAt = pass.at
		result.MinLower = pass.lower
		result.MinConverged = pass.converged
		result.Evals += pass.evals
	}
	if mode == "max" || mode == "both" {
		// The supremum of f is -(infimum of -f); the negation happens on the
		// sampled values, so one compiled query serves both passes.
		pass, err := bbMinimize(ctx, prog, region, true, lipschitz, tolerance, absFloor)
		if err != nil {
			return nil, err
		}
		result.Max = -pass.value
		result.MaxAt = pass.at
		result.MaxUpper = -pass.lower
		result.MaxConverged = pass.converged
		result.Evals += pass.evals
	}
	return result, nil
}

// bbCell is one branch-and-bound cell: an axis-aligned box with the field
// sampled at its center. bound = value - lipschitz*radius is the least value
// any point in the cell could take; a cell whose bound cannot beat the
// incumbent is pruned.
type bbCell struct {
	center [3]float64
	half   [3]float64
	value  float64
	bound  float64
}

type bbPass struct {
	value     float64
	at        [3]float64
	lower     float64
	evals     int
	converged bool
}

// bbMinimize runs Lipschitz branch-and-bound for the minimum of the
// package's single scalar return over the region (of -f when negate is
// set). Each round evaluates the children of the best-bound active cells in
// one batch, updates the incumbent, and re-prunes. tolerance > 0 is the
// caller's exact convergence contract; tolerance <= 0 selects the adaptive
// default max(absFloor, bbRelTolerance*|incumbent|). The certificate tracks
// every cell ever discarded unpruned (pool overflow), so the reported lower
// bound stays sound even when the search is cut short.
func bbMinimize(ctx context.Context, prog *program.Program, region [6]float64, negate bool, lipschitz, tolerance, absFloor float64) (*bbPass, error) {
	pass := &bbPass{value: math.Inf(1), lower: math.Inf(-1)}
	// droppedFloor is the least bound among cells discarded unexplored; the
	// final lower bound may never rise above it.
	droppedFloor := math.Inf(1)

	evalCells := func(cells []bbCell) error {
		xs := make([]float32, len(cells))
		ys := make([]float32, len(cells))
		zs := make([]float32, len(cells))
		for i, c := range cells {
			xs[i] = float32(c.center[0])
			ys[i] = float32(c.center[1])
			zs[i] = float32(c.center[2])
		}
		results, err := EvalGrad(prog, xs, ys, zs)
		if err != nil {
			return err
		}
		if len(results) != 1 {
			return fmt.Errorf("extremum requires a single scalar expression, got %d values", len(results))
		}
		for i := range cells {
			c := &cells[i]
			v := float64(results[0][i].Value)
			if negate {
				v = -v
			}
			radius := math.Sqrt(c.half[0]*c.half[0] + c.half[1]*c.half[1] + c.half[2]*c.half[2])
			if math.IsNaN(v) {
				// An unevaluable center can't prune its cell: keep exploring
				// it, but never let it become the incumbent.
				c.value = math.Inf(1)
				c.bound = math.Inf(-1)
				continue
			}
			c.value = v
			c.bound = v - lipschitz*radius
			if v < pass.value {
				pass.value = v
				pass.at = c.center
			}
		}
		pass.evals += len(cells)
		return nil
	}

	// Seed with a uniform grid.
	var cells []bbCell
	var half [3]float64
	for k := 0; k < 3; k++ {
		half[k] = (region[k+3] - region[k]) / (2 * bbInitialGrid)
	}
	for iz := 0; iz < bbInitialGrid; iz++ {
		for iy := 0; iy < bbInitialGrid; iy++ {
			for ix := 0; ix < bbInitialGrid; ix++ {
				cells = append(cells, bbCell{
					center: [3]float64{
						region[0] + (2*float64(ix)+1)*half[0],
						region[1] + (2*float64(iy)+1)*half[1],
						region[2] + (2*float64(iz)+1)*half[2],
					},
					half: half,
				})
			}
		}
	}
	if err := evalCells(cells); err != nil {
		return nil, err
	}

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		// Prune, and take stock of the certificate.
		active := cells[:0]
		activeLower := math.Inf(1)
		for _, c := range cells {
			if c.bound < pass.value {
				active = append(active, c)
				if c.bound < activeLower {
					activeLower = c.bound
				}
			}
		}
		cells = active
		pass.lower = math.Min(math.Min(activeLower, droppedFloor), pass.value)
		effTolerance := tolerance
		if effTolerance <= 0 {
			effTolerance = absFloor
			if !math.IsInf(pass.value, 0) {
				if rel := bbRelTolerance * math.Abs(pass.value); rel > effTolerance {
					effTolerance = rel
				}
			}
		}
		if pass.value-pass.lower <= effTolerance {
			pass.converged = true
			return pass, nil
		}
		if len(cells) == 0 || pass.evals+8 > bbEvalBudget {
			return pass, nil
		}

		sort.Slice(cells, func(i, j int) bool { return cells[i].bound < cells[j].bound })
		splitCount := bbSplitPerRound
		if remaining := (bbEvalBudget - pass.evals) / 8; splitCount > remaining {
			splitCount = remaining
		}
		if splitCount > len(cells) {
			splitCount = len(cells)
		}
		splitting := cells[:splitCount]
		rest := cells[splitCount:]
		if len(rest) > bbMaxActive {
			for _, c := range rest[bbMaxActive:] {
				if c.bound < droppedFloor {
					droppedFloor = c.bound
				}
			}
			rest = rest[:bbMaxActive]
		}

		children := make([]bbCell, 0, len(splitting)*8)
		for _, c := range splitting {
			h := [3]float64{c.half[0] / 2, c.half[1] / 2, c.half[2] / 2}
			for oz := -1; oz <= 1; oz += 2 {
				for oy := -1; oy <= 1; oy += 2 {
					for ox := -1; ox <= 1; ox += 2 {
						children = append(children, bbCell{
							center: [3]float64{
								c.center[0] + float64(ox)*h[0],
								c.center[1] + float64(oy)*h[1],
								c.center[2] + float64(oz)*h[2],
							},
							half: h,
						})
					}
				}
			}
		}
		if err := evalCells(children); err != nil {
			return nil, err
		}
		cells = append(rest, children...)
	}
}

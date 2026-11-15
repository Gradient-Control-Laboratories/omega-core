//go:build cgo

package eval

import (
	"context"
	"fmt"
	"math"
)

// defaultVolumeTolerance is the relative tolerance Volume takes when the caller names none.
const defaultVolumeTolerance = 0.05

// volumeBudget bounds the interval evaluations of one volume: where it ends the search, the bracket is sound but
// wider than the tolerance asked, and the result says so.
var volumeBudget = 64_000_000

// ReasonDepth is why a VolumeBounds did not converge when the search reached its deepest subdivision first.
const ReasonDepth = "depth"

// Volumeis VolumeOf for the document as written: no manifest and no defs.
func (e *Evaluator) Volume(ctx context.Context, code, expr string, region [6]float64, relativeTolerance float64) (VolumeBounds, error) {
	return e.VolumeOf(ctx, nil, code, "", expr, region, relativeTolerance)
}

// VolumeOf classifies cells using field intervals and charges every unresolved cell to Upper.
// defs are applied to the document as the other queries apply them, so the bracket is of the part
// a slider was moved to. relativeTolerance <= 0 selects defaultVolumeTolerance.
func (e *Evaluator) VolumeOf(ctx context.Context, manifest map[string]string, code, defs, expr string, region [6]float64, relativeTolerance float64) (VolumeBounds, error) {
	var result VolumeBounds
	if err := startQuery(ctx); err != nil {
		return result, err
	}
	defer endQuery()
	var root [6]float32
	for axis := 0; axis < 3; axis++ {
		if math.IsNaN(region[axis]) || math.IsNaN(region[axis+3]) || math.IsInf(region[axis], 0) || math.IsInf(region[axis+3], 0) || region[axis] >= region[axis+3] {
			return result, fmt.Errorf("volume requires a finite containing region")
		}
		root[axis] = math.Nextafter32(float32(region[axis]), float32(math.Inf(-1)))
		root[axis+3] = math.Nextafter32(float32(region[axis+3]), float32(math.Inf(1)))
	}
	if relativeTolerance <= 0 {
		relativeTolerance = defaultVolumeTolerance
	}
	if relativeTolerance >= 1 || math.IsNaN(relativeTolerance) {
		return result, fmt.Errorf("volume tolerance must be between zero and one")
	}
	prog, err := e.prepareQuery(ctx, manifest, code, defs, []string{expr})
	if err != nil {
		return result, err
	}
	volume := func(box [6]float32) float64 {
		return (float64(box[3]) - float64(box[0])) * (float64(box[4]) - float64(box[1])) * (float64(box[5]) - float64(box[2]))
	}
	result.Upper = volume(root)
	type cell struct {
		box   [6]float32
		depth int
	}
	for depth := 18; depth <= 36 && result.Evals < volumeBudget; depth += 2 {
		stack := []cell{{box: root}}
		lower, upper := 0.0, volume(root)
		for len(stack) > 0 && result.Evals < volumeBudget {
			if err := ctx.Err(); err != nil {
				return result, err
			}
			count := min(len(stack), 4096, volumeBudget-result.Evals)
			batch := append([]cell(nil), stack[len(stack)-count:]...)
			stack = stack[:len(stack)-count]
			boxes := make([][6]float32, count)
			for i, node := range batch {
				boxes[i] = node.box
			}
			intervals, err := EvalInterval(prog, boxes)
			if err != nil {
				return result, err
			}
			if len(intervals) != 1 || len(intervals[0]) != len(batch) {
				return result, fmt.Errorf("volume expression must be scalar")
			}
			result.Evals += len(batch)
			for i, node := range batch {
				box := node.box
				sample := intervals[0][i]
				if sample.Upper < 0 {
					lower += volume(box)
				} else if sample.Lower >= 0 {
					upper -= volume(box)
				} else if node.depth < depth {
					axis := 0
					for j := 1; j < 3; j++ {
						if box[j+3]-box[j] > box[axis+3]-box[axis] {
							axis = j
						}
					}
					mid := box[axis] + (box[axis+3]-box[axis])/2
					left, right := box, box
					left[axis+3], right[axis] = mid, mid
					stack = append(stack, cell{left, node.depth + 1}, cell{right, node.depth + 1})
				}
			}
		}
		result.Lower, result.Upper = max(result.Lower, lower), min(result.Upper, upper)
		if result.Upper-result.Lower <= relativeTolerance*result.Upper {
			result.Converged = true
			break
		}
	}
	if !result.Converged {
		result.Reason = ReasonDepth
		if result.Evals >= volumeBudget {
			result.Reason = ReasonBudget
		}
	}
	result.Upper *= 1 + 1e-12
	result.Lower *= 1 - 1e-12
	return result, nil
}

//go:build cgo

package eval

import (
	"context"
	"fmt"
	"math"
)

// Volume classifies cells using field intervals and charges every unresolved cell to Upper.
func (e *Evaluator) Volume(ctx context.Context, code, expr string, region [6]float64, relativeTolerance float64) (VolumeBounds, error) {
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
	if relativeTolerance <= 0 || relativeTolerance >= 1 || math.IsNaN(relativeTolerance) {
		return result, fmt.Errorf("volume tolerance must be between zero and one")
	}
	prog, err := e.prepareQuery(ctx, nil, code, "", []string{expr})
	if err != nil {
		return result, err
	}
	volume := func(box [6]float32) float64 {
		return (float64(box[3]) - float64(box[0])) * (float64(box[4]) - float64(box[1])) * (float64(box[5]) - float64(box[2]))
	}
	result.Upper = volume(root)
	const budget = 64_000_000
	type cell struct {
		box   [6]float32
		depth int
	}
	for depth := 18; depth <= 36 && result.Evals < budget; depth += 2 {
		stack := []cell{{box: root}}
		lower, upper := 0.0, volume(root)
		for len(stack) > 0 && result.Evals < budget {
			if err := ctx.Err(); err != nil {
				return result, err
			}
			count := min(len(stack), 4096, budget-result.Evals)
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
			break
		}
	}
	result.Upper *= 1 + 1e-12
	result.Lower *= 1 - 1e-12
	return result, nil
}

package mesh

import (
	"fmt"
	"sort"

	"gradientcontrol.com/omega/core/backend/internal/fidget/internal/convert"
	"gradientcontrol.com/omega/core/backend/internal/fidget/internal/fidget"
	"gradientcontrol.com/omega/core/program"
)

// EvalGrad evaluates each scalar return of the program at the given points,
// returning the field value and spatial gradient per point. Vector-valued
// returns flatten into consecutive scalar results, matching how the meshing
// pipeline treats returns. xs, ys, zs must be the same length.
func EvalGrad(
	prog *program.Program,
	xs, ys, zs []float32,
) ([][]fidget.GradResult, error) {
	var results [][]fidget.GradResult
	err := convert.ToTrees(prog, func(treeIndices []fidget.TreeIndex, _ []string) error {
		for _, treeIndex := range treeIndices {
			canvasIndex := fidget.BuildCanvas(treeIndex)
			grads, err := fidget.EvalGrad(canvasIndex, xs, ys, zs)
			if err != nil {
				return err
			}
			results = append(results, grads)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

// EvalInterval interval-bounds each scalar return of the program over every
// box, returning a sound [lower, upper] bracket per box (oracle-aware). It
// mirrors EvalGrad's canvas-build path; a query program has one scalar return,
// so results[0] holds the brackets. Each box is [xMin,yMin,zMin,xMax,yMax,zMax].
func EvalInterval(
	prog *program.Program,
	boxes [][6]float32,
) ([][]fidget.IntervalResult, error) {
	var results [][]fidget.IntervalResult
	err := convert.ToTrees(prog, func(treeIndices []fidget.TreeIndex, _ []string) error {
		for _, treeIndex := range treeIndices {
			canvasIndex := fidget.BuildCanvas(treeIndex)
			ivals, err := fidget.EvalInterval(canvasIndex, boxes)
			if err != nil {
				return err
			}
			results = append(results, ivals)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

// GradGroup is one return value's evaluation: its GSL type name and the
// per-component samples ([component][point]). A scalar has one component.
type GradGroup struct {
	TypeName string
	Samples  [][]fidget.GradResult
}

// EvalGradGroups evaluates every return of the program at the given points,
// accepting vector- and matrix-valued returns: each return becomes one
// GradGroup whose components are evaluated in order.
func EvalGradGroups(
	prog *program.Program,
	xs, ys, zs []float32,
) ([]GradGroup, error) {
	var results []GradGroup
	err := convert.ToTreeGroups(prog, func(treeIndices []fidget.TreeIndex, groups []convert.ReturnGroup) error {
		next := 0
		for _, group := range groups {
			g := GradGroup{TypeName: group.TypeName}
			for c := 0; c < group.Count; c++ {
				canvasIndex := fidget.BuildCanvas(treeIndices[next])
				next++
				grads, err := fidget.EvalGrad(canvasIndex, xs, ys, zs)
				if err != nil {
					return err
				}
				g.Samples = append(g.Samples, grads)
			}
			results = append(results, g)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

// EvalIntervalParamBoxes interval-bounds the program's single scalar return
// over each spatial box with each box's OWN parameter ranges: names orders
// the parameters and bounds[i] carries box i's [lo, hi] per name. Each
// param's slider declaration builds as a tape var, so every bracket is sound
// over its box and its parameter ranges simultaneously — the substrate for a
// search that subdivides parameters alongside space.
func EvalIntervalParamBoxes(
	prog *program.Program,
	names []string,
	boxes [][6]float32,
	bounds [][][2]float64,
) ([]fidget.IntervalResult, error) {
	var results []fidget.IntervalResult
	err := convert.ToTreesWithParams(prog, names, func(treeIndices []fidget.TreeIndex, paramTrees map[string]fidget.TreeIndex) error {
		if len(treeIndices) != 1 {
			return fmt.Errorf("parametric interval evaluation requires a single scalar return, got %d", len(treeIndices))
		}
		varTrees := make([]fidget.TreeIndex, 0, len(names))
		for _, n := range names {
			varTrees = append(varTrees, paramTrees[n])
		}
		rows := make([][][2]float32, len(bounds))
		for i, row := range bounds {
			rows[i] = make([][2]float32, len(row))
			for v, b := range row {
				rows[i][v] = [2]float32{float32(b[0]), float32(b[1])}
			}
		}
		canvasIndex := fidget.BuildCanvas(treeIndices[0])
		ivals, err := fidget.EvalIntervalVars(canvasIndex, boxes, varTrees, rows)
		if err != nil {
			return err
		}
		results = ivals
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

// EvalIntervalParams is EvalIntervalParamBoxes with one shared parameter
// range applied to every box.
func EvalIntervalParams(
	prog *program.Program,
	boxes [][6]float32,
	params map[string][2]float64,
) ([]fidget.IntervalResult, error) {
	names := make([]string, 0, len(params))
	for n := range params {
		names = append(names, n)
	}
	sort.Strings(names)
	row := make([][2]float64, len(names))
	for i, n := range names {
		row[i] = params[n]
	}
	bounds := make([][][2]float64, len(boxes))
	for i := range bounds {
		bounds[i] = row
	}
	return EvalIntervalParamBoxes(prog, names, boxes, bounds)
}

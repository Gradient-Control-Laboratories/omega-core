//go:build cgo

package eval

import (
	"fmt"
	"sort"

	"github.com/Gradient-Control-Laboratories/omega-core/internal/fidget"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/fidget/convert"
	"github.com/Gradient-Control-Laboratories/omega-core/program"
)

// EvalGrad evaluates each scalar return of the program at the given points,
// returning the field value and spatial gradient per point. Vector-valued
// returns flatten into consecutive scalar results, matching how the meshing
// pipeline treats returns. xs, ys, zs must be the same length.
func EvalGrad(
	prog *program.Program,
	xs, ys, zs []float32,
) ([][]GradSample, error) {
	return evalGrad(prog, convert.Surface{}, xs, ys, zs)
}

// evalGrad is EvalGrad with the surfaces at s (Evaluator.SetCondition).
func evalGrad(
	prog *program.Program,
	s convert.Surface,
	xs, ys, zs []float32,
) ([][]GradSample, error) {
	var results [][]GradSample
	err := scalarTrees(prog, s, func(treeIndices []fidget.TreeIndex) error {
		for _, treeIndex := range treeIndices {
			canvasIndex := fidget.BuildCanvas(treeIndex)
			grads, err := fidget.EvalGrad(canvasIndex, xs, ys, zs)
			if err != nil {
				return err
			}
			results = append(results, toGradSamples(grads))
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
) ([][]IntervalSample, error) {
	return evalInterval(prog, convert.Surface{}, boxes)
}

// evalInterval is EvalInterval with the surfaces at s (Evaluator.SetCondition).
func evalInterval(
	prog *program.Program,
	s convert.Surface,
	boxes [][6]float32,
) ([][]IntervalSample, error) {
	var results [][]IntervalSample
	err := scalarTrees(prog, s, func(treeIndices []fidget.TreeIndex) error {
		for _, treeIndex := range treeIndices {
			canvasIndex := fidget.BuildCanvas(treeIndex)
			ivals, err := fidget.EvalInterval(canvasIndex, boxes)
			if err != nil {
				return err
			}
			results = append(results, toIntervalSamples(ivals))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

// EvalGradGroups evaluates every return of the program at the given points,
// accepting vector- and matrix-valued returns: each return becomes one
// GradGroup whose components are evaluated in order.
func EvalGradGroups(
	prog *program.Program,
	xs, ys, zs []float32,
) ([]GradGroup, error) {
	return evalGradGroups(prog, convert.Surface{}, xs, ys, zs)
}

// evalGradGroups is EvalGradGroups with the surfaces at s (Evaluator.SetCondition).
func evalGradGroups(
	prog *program.Program,
	s convert.Surface,
	xs, ys, zs []float32,
) ([]GradGroup, error) {
	var results []GradGroup
	err := groupTrees(prog, s, func(treeIndices []fidget.TreeIndex, groups []convert.ReturnGroup) error {
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
				g.Samples = append(g.Samples, toGradSamples(grads))
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
) ([]IntervalSample, error) {
	return evalIntervalParamBoxes(prog, convert.Surface{}, names, boxes, bounds)
}

// evalIntervalParamBoxes is EvalIntervalParamBoxes with the surfaces at s
// (Evaluator.SetCondition).
func evalIntervalParamBoxes(
	prog *program.Program,
	s convert.Surface,
	names []string,
	boxes [][6]float32,
	bounds [][][2]float64,
) ([]IntervalSample, error) {
	var results []IntervalSample
	err := paramTrees(prog, s, names, func(treeIndices []fidget.TreeIndex, paramTrees map[string]fidget.TreeIndex) error {
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
		results = toIntervalSamples(ivals)
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
) ([]IntervalSample, error) {
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

// scalarTrees builds prog's scalar returns with the surfaces at s: ToTrees's
// trees for the field as built, else ToSurfaceTrees's, refusing a return that
// is not a number as ToTrees refuses it.
func scalarTrees(prog *program.Program, s convert.Surface, f func([]fidget.TreeIndex) error) error {
	if !s.Moved() {
		return convert.ToTrees(prog, func(treeIndices []fidget.TreeIndex, _ []string) error {
			return f(treeIndices)
		})
	}
	return convert.ToSurfaceTrees(prog, s, nil, func(treeIndices []fidget.TreeIndex, groups []convert.ReturnGroup, _ map[string]fidget.TreeIndex) error {
		if err := scalarGroups(groups); err != nil {
			return err
		}
		return f(treeIndices)
	})
}

// groupTrees is scalarTrees for returns of any numeric type, as ToTreeGroups
// builds them.
func groupTrees(prog *program.Program, s convert.Surface, f func([]fidget.TreeIndex, []convert.ReturnGroup) error) error {
	if !s.Moved() {
		return convert.ToTreeGroups(prog, f)
	}
	return convert.ToSurfaceTrees(prog, s, nil, func(treeIndices []fidget.TreeIndex, groups []convert.ReturnGroup, _ map[string]fidget.TreeIndex) error {
		return f(treeIndices, groups)
	})
}

// paramTrees is scalarTrees with the named declarations built as input vars,
// as ToTreesWithParams builds them.
func paramTrees(prog *program.Program, s convert.Surface, names []string, f func([]fidget.TreeIndex, map[string]fidget.TreeIndex) error) error {
	if !s.Moved() {
		return convert.ToTreesWithParams(prog, names, f)
	}
	return convert.ToSurfaceTrees(prog, s, names, func(treeIndices []fidget.TreeIndex, groups []convert.ReturnGroup, vars map[string]fidget.TreeIndex) error {
		if err := scalarGroups(groups); err != nil {
			return err
		}
		return f(treeIndices, vars)
	})
}

// scalarGroups refuses a return that is not one number.
func scalarGroups(groups []convert.ReturnGroup) error {
	for _, g := range groups {
		if g.Count != 1 {
			return fmt.Errorf("expected numeric return values, got %s", g.TypeName)
		}
	}
	return nil
}

func toGradSamples(samples []fidget.GradResult) []GradSample {
	out := make([]GradSample, len(samples))
	for i, s := range samples {
		out[i] = GradSample{Value: s.Value, Dx: s.Dx, Dy: s.Dy, Dz: s.Dz}
	}
	return out
}

func toIntervalSamples(samples []fidget.IntervalResult) []IntervalSample {
	out := make([]IntervalSample, len(samples))
	for i, s := range samples {
		out[i] = IntervalSample{Lower: s.Lower, Upper: s.Upper}
	}
	return out
}

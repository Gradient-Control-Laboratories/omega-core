package closure

import (
	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/program"
)

// operandContainsCall reports whether fc is op or one of its operands or
// arguments, at any depth.
func operandContainsCall(op *program.Operand, fc *program.FnCall) bool {
	if op == nil {
		return false
	}
	switch v := op.T.(type) {
	case *program.Operand_FnCall:
		if v.FnCall == fc {
			return true
		}
		if v.FnCall != nil {
			for _, a := range v.FnCall.Arguments {
				if operandContainsCall(a, fc) {
					return true
				}
			}
		}
	case *program.Operand_Expression:
		if v.Expression != nil {
			for _, o := range v.Expression.Operands {
				if operandContainsCall(o, fc) {
					return true
				}
			}
		}
	}
	return false
}

// capturesReadBy is what a lambda passed to the higher-order call fc captures
// from enclosingFn, the function it is written in: the locals assigned before
// the call that target, the lambda, reads, directly or through another such
// local. Capturing every local instead took in the very assignment holding
// fc, and any after it, and specializing the lambda specialized fc again,
// without end.
//
// A parameter of enclosingFn read the same way is captured too, ahead of the
// locals, as a reference to itself; readsParam reports that one was. Such a
// capture names what only enclosingFn has, so it cannot be baked into a
// function of its own: the caller lifts the lambda instead.
func capturesReadBy(enclosingFn, target *program.Function, fc *program.FnCall) (captures []captureBinding, readsParam bool) {
	own := map[string]bool{}
	for _, p := range target.Parameters {
		own[p.Name] = true
	}
	for _, a := range target.Assignments {
		for _, v := range a.Variables {
			own[v.Name] = true
		}
	}
	need := map[string]bool{}
	add := func(op *program.Operand) {
		walkBareReferences(op, func(_ *program.Operand, r *program.Reference) {
			if !own[r.Name] {
				need[r.Name] = true
			}
		})
	}
	for _, a := range target.Assignments {
		add(a.Expression)
	}
	for _, r := range target.Returns {
		add(r)
	}
	end := len(enclosingFn.Assignments)
	for i, a := range enclosingFn.Assignments {
		if operandContainsCall(a.Expression, fc) {
			end = i
			break
		}
	}
	keep := make([]bool, end)
	for i := end - 1; i >= 0; i-- {
		a := enclosingFn.Assignments[i]
		for _, v := range a.Variables {
			if need[v.Name] {
				keep[i] = true
			}
		}
		if keep[i] {
			add(a.Expression)
		}
	}
	rebound := map[string]bool{}
	for i := 0; i < end; i++ {
		for _, v := range enclosingFn.Assignments[i].Variables {
			rebound[v.Name] = true
		}
	}
	var out []captureBinding
	for _, p := range enclosingFn.Parameters {
		if need[p.Name] && !rebound[p.Name] {
			out = append(out, captureBinding{
				name:  p.Name,
				value: &program.Operand{T: &program.Operand_Reference{Reference: &program.Reference{Name: p.Name}}},
			})
			readsParam = true
		}
	}
	for i := 0; i < end; i++ {
		if !keep[i] {
			continue
		}
		a := enclosingFn.Assignments[i]
		for _, v := range a.Variables {
			out = append(out, captureBinding{name: v.Name, value: proto.Clone(a.Expression).(*program.Operand)})
		}
	}
	return out, readsParam
}

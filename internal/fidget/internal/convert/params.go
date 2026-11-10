package convert

import (
	"errors"
	"fmt"

	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/internal/fidget/internal/fidget"
	"gradientcontrol.com/omega/core/program"
)

// SetParamVars asks the delegate to build the named top-level declarations as
// fidget input VARIABLES instead of baked constants. A requested name whose
// declaration is a single numeric literal (the slider form) becomes a tape
// var carrying its literal as the current value; ParamTrees reports the tree
// index built for each, which interval evaluation can widen to a range. A
// requested name that never builds as a literal is simply absent from
// ParamTrees — the caller decides whether that is an error. Requesting
// program.TimeVarName builds the implicit clock as the var instead
// (BuildPackage handles it), so a certificate can range over time.
func (d *Delegate) SetParamVars(names []string) {
	if len(names) == 0 {
		return
	}
	d.paramRequest = make(map[string]bool, len(names))
	for _, n := range names {
		d.paramRequest[n] = true
	}
	d.paramTrees = make(map[string]fidget.TreeIndex, len(names))
}

// ParamTrees returns the input-var tree built for each requested param name.
func (d *Delegate) ParamTrees() map[string]fidget.TreeIndex {
	return d.paramTrees
}

// paramAssignment intercepts a requested slider declaration: a single
// constant tree is rebuilt as a named input var at the same value. Returns
// nil when the assignment is not a requested param (the normal path applies).
func (d *Delegate) paramAssignment(name string, expr build.Expression) build.Expression {
	if d.paramRequest == nil || !d.paramRequest[name] {
		return nil
	}
	if _, done := d.paramTrees[name]; done {
		// Only the first (top-level) declaration becomes the var; a shadowing
		// local inside a function keeps its constant meaning.
		return nil
	}
	node, ok := expr.(*TreeNode)
	if !ok || len(node.trees) != 1 || len(node.constValues) != 1 {
		// Seen, but not a slider-style literal. Record it: that is a real
		// caller error, and must not be mistaken for a parameter the built
		// expression simply never referenced.
		if d.paramRejected == nil {
			d.paramRejected = make(map[string]bool)
		}
		d.paramRejected[name] = true
		return nil
	}
	varTree := fidget.NewInput(name, node.constValues[0])
	pv, err := NewTreeNode(node.Operand(), node.Start(), varTree)
	if err != nil {
		return nil
	}
	d.paramTrees[name] = varTree
	return pv
}

// ToTreesWithParams is ToTrees with the named declarations built as input
// vars: f additionally receives the tree index of each param's var, valid
// only inside f. Params that did not resolve to a slider-style literal
// declaration are reported as an error before f runs — a certificate over a
// range the tape never saw would be unsound. The implicit clock
// (program.TimeVarName) always resolves, as a var in place of its constant.
func ToTreesWithParams(
	prog *program.Program,
	params []string,
	f func(treeIndices []fidget.TreeIndex, paramTrees map[string]fidget.TreeIndex) error,
) error {
	if prog == nil || prog.Primary == nil {
		return fmt.Errorf("program or primary package is nil")
	}
	var gslSource string
	if len(prog.Primary.Files) > 0 {
		gslSource = prog.Primary.Files[0].Source
	}
	// Mark and build the same copy: markParamLiterals keys operands by pointer.
	prog = buildCopy(prog)
	d := NewDelegate(gslSource)
	d.SetParamVars(params)
	d.markParamLiterals(prog)

	fidget.Lock()
	defer fidget.Unlock()
	defer fidget.Clear()

	if err := build.Build(d, prog, true); err != nil {
		return build.ErrWithPkgSrc(prog, err)
	}
	if len(d.outputs) == 0 {
		return errors.New("no outputs")
	}
	for _, name := range params {
		if _, ok := d.paramTrees[name]; ok {
			continue
		}
		if d.paramRejected[name] {
			return fmt.Errorf("parameter %q is not a slider-style literal declaration in this document", name)
		}
		// Never seen at all: nothing in the built expression referenced it, so
		// the dead-code pass dropped its declaration. That is not an error and
		// not the folding hazard IsParameter guards against — a parameter
		// reference is never folded to a constant, so an absent parameter is
		// one this expression genuinely does not depend on, and the claim
		// holds for every value of it. Bind a free var the expression does not
		// use, so the caller's name-indexed lookup still resolves. Callers
		// that must reject a name the DOCUMENT never declared check that
		// against the document text, which survives dead-code elimination —
		// see Evaluator.CertifyParams.
		d.paramTrees[name] = fidget.NewInput(name, 0)
	}
	treeIndices := make([]fidget.TreeIndex, 0, len(d.outputs))
	for _, node := range d.outputs {
		if len(node.Trees()) != 1 {
			return build.ErrLoc(d.GetScope(), node.Start(),
				fmt.Errorf("expected numeric return values, got %s", node.Type().String()))
		}
		treeIndices = append(treeIndices, node.Trees()[0])
	}
	return f(treeIndices, d.paramTrees)
}

// IsParameter implements eval.ParameterChecker. A requested parameter's
// declaration is a numeric literal, so the linker records ConstValues on every
// operand that reads it — and the builder short-circuits any operand carrying
// ConstValues straight to a constant without visiting it. That bypasses the
// input var paramAssignment binds, leaving interval evaluation to return the
// parameter's nominal value for every range it is handed: a claim certified at
// a single point and reported as proven over the whole interval.
//
// Reporting a parameter's own references here keeps them dynamic, so they build
// through BuildVarRef and reach the var.
//
// The fold does not ask about references, though: it asks about literals, and
// evaluates a reference by evaluating the assignment behind it. So a top-level
// declaration that reads a parameter but not p, c = k * 2.0, folded to its
// value at k's nominal, the assembler's stamp stood, and every use of c was
// the constant. A certificate then held at the nominal and was reported over
// the whole range. Reporting the declaration's literal as well (paramLiterals)
// makes the fold fail on it, as the viewer's checker does for a slider, so
// everything derived from the parameter stays dynamic and the assembler's
// stamps are cleared (eval's setConst).
func (d *Delegate) IsParameter(op *program.Operand) bool {
	if d.paramRequest == nil || op == nil {
		return false
	}
	if d.paramLiterals[op] {
		return true
	}
	ref := op.GetReference()
	if ref == nil || ref.GetPrefix() != "" {
		// Only unqualified names: a parameter is a top-level declaration in
		// the document under test, never an imported package member.
		return false
	}
	return d.paramRequest[ref.GetName()]
}

// markParamLiterals records the literal each requested parameter's top-level
// declaration assigns: the value itself, or the value under a unary minus, the
// slider forms paramAssignment accepts. Only the first declaration of a name
// counts, as there.
func (d *Delegate) markParamLiterals(prog *program.Program) {
	if d.paramRequest == nil || prog == nil || prog.Primary == nil {
		return
	}
	seen := make(map[string]bool)
	for _, f := range prog.Primary.Files {
		for _, a := range f.Assignments {
			if len(a.Variables) != 1 {
				continue
			}
			name := a.Variables[0].Name
			if !d.paramRequest[name] || seen[name] {
				continue
			}
			seen[name] = true
			op := a.Expression
			if e := op.GetExpression(); e != nil && e.Operation == program.Operation_NEG && len(e.Operands) == 1 {
				op = e.Operands[0]
			}
			if _, ok := op.GetT().(*program.Operand_Value); !ok {
				continue
			}
			if d.paramLiterals == nil {
				d.paramLiterals = make(map[*program.Operand]bool)
			}
			d.paramLiterals[op] = true
		}
	}
}

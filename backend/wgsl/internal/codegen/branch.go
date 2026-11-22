package codegen

import (
	"fmt"

	"gradientcontrol.com/omega/core/backend/wgsl/internal/ir"
	"gradientcontrol.com/omega/core/backend/wgsl/internal/wgsl"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// A CHOICE THE SAME EVERYWHERE. A GSL ternary is WGSL's select(falseValue,
// trueValue, condition), which evaluates both values. builtin's
// choose(condition, onTrue, onFalse) is the same ternary, but its author says
// every point takes the same side (a constant, a slider, a descriptor handed
// to a function), so the codegen writes it as a branch instead of a call:
//
//	let gsl_branch = condition;
//	var gsl_onTrue : T;
//	if (gsl_branch) { gsl_onTrue = onTrue; }
//	var gsl_onFalse : T;
//	if (!gsl_branch) { gsl_onFalse = onFalse; }
//	… select(gsl_onFalse, gsl_onTrue, gsl_branch) …
//
// ir.SinkStatements moves into each guard the lets only it reads, so the side
// not taken costs nothing. Each guard has a variable of its own, since the
// sink keeps a guard's variable just before it. The author is trusted, not
// checked: a parameter of a function can carry p from its caller, which the
// codegen cannot see. Where the condition does vary, both sides run, about as
// a select would; no implicit derivative (textureSample, dpdx), whose use WGSL
// refuses in control flow that is not uniform, is emitted. A condition that
// reads nothing is a constant the GPU compiler folds out of a select, and a
// vector condition selects per component, so both stay selects. The CPU and
// interval paths read choose's GSL body, the ternary.

// chooseWGSLName is the name codegen gives builtin's choose.
const chooseWGSLName = "builtin_choose"

// chooses reports whether a function is builtin's choose(condition, onTrue,
// onFalse).
func chooses(wgslName string, argCount int) bool {
	return wgslName == chooseWGSLName && argCount == 3
}

// ChooseExpr is a call to choose, written as a branch.
type ChooseExpr struct {
	wgsl.Expression
	d       *Delegate
	cond    wgsl.Expression
	onTrue  wgsl.Expression
	onFalse wgsl.Expression
	// written are what the value written reads, once Codegen has run: the
	// branch's names when it made one, so a statement holding it does not
	// keep the lets the guards should take in.
	written []string
}

// chooseBranch is a call to choose: its branch's statements, written into the
// function being generated, and the chosen value.
func (d *Delegate) chooseBranch(op *program.Operand, args []wgsl.Expression, returnType wgsl.TypeInfo) *ChooseExpr {
	c := &ChooseExpr{d: d, cond: args[0].ToBool().Unwrap(), onTrue: args[1], onFalse: args[2]}
	c.Expression = wgsl.NewExpr(op, returnType, build.PrecedenceHighest)
	c.Expression.SetOuter(c)
	return c
}

func (c *ChooseExpr) Reads() []string {
	if c.written != nil {
		return append([]string(nil), c.written...)
	}
	return dedupReadsOf(c.cond, c.onTrue, c.onFalse)
}

func (c *ChooseExpr) Codegen(ctx wgsl.CodegenContext) string {
	// In the order a select writes them, false, true, condition: a value can
	// add statements as it is written.
	falseSrc := c.onFalse.Codegen(ctx)
	trueSrc := c.onTrue.Codegen(ctx)
	condSrc := c.cond.Codegen(ctx)
	if c.cond.WGSLType().Name != wgsl.TypeBool.Name || len(c.cond.Reads()) == 0 {
		return fmt.Sprintf("select(%s, %s, %s)", falseSrc, trueSrc, condSrc)
	}
	d := c.d
	block := d.currentScope().Block()
	branch := d.uniqueLocalName("gsl_branch")
	whenTrue := d.uniqueLocalName("gsl_onTrue")
	whenFalse := d.uniqueLocalName("gsl_onFalse")
	typeName := c.WGSLType().Name
	block.Add(ir.NewLetStmt(branch, "", condSrc, c.cond.Reads()))
	block.Add(ir.NewVarStmt(whenTrue, typeName, "", nil))
	onTrue := ir.NewBlock()
	onTrue.Add(ir.NewAssignStmt(whenTrue, trueSrc, c.onTrue.Reads()))
	block.Add(ir.NewGuardStmt(branch, onTrue, []string{branch}))
	block.Add(ir.NewVarStmt(whenFalse, typeName, "", nil))
	onFalse := ir.NewBlock()
	onFalse.Add(ir.NewAssignStmt(whenFalse, falseSrc, c.onFalse.Reads()))
	block.Add(ir.NewGuardStmt("!"+branch, onFalse, []string{branch}))
	c.written = []string{branch, whenTrue, whenFalse}
	return fmt.Sprintf("select(%s, %s, %s)", whenFalse, whenTrue, branch)
}

// dedupReadsOf is what the expressions read, each name once.
func dedupReadsOf(es ...wgsl.Expression) []string {
	seen := map[string]bool{}
	var out []string
	for _, e := range es {
		for _, r := range e.Reads() {
			if !seen[r] {
				seen[r] = true
				out = append(out, r)
			}
		}
	}
	return out
}

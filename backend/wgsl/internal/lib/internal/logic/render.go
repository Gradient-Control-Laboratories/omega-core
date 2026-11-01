package logic

import (
	"fmt"

	api "gradientcontrol.com/omega/core/backend/wgsl/internal/lib"
	"gradientcontrol.com/omega/core/backend/wgsl/internal/lib/internal/codegen"
	"gradientcontrol.com/omega/core/internal/edit"
	"gradientcontrol.com/omega/core/program"
)

// collectibleFunction groups a function name with its AST declaration.
// It is used as a helper to easily traverse package-level function declarations
// during location span collection.
type collectibleFunction struct {
	name string
	fn   *program.Function
}

// locationCollector traverses the AST to collect all source locations that contribute to a shape
type locationCollector struct {
	module  *program.Package
	spans   []edit.Span
	visited map[string]bool // track visited declarations to avoid cycles
	// seen holds the spans already collected. An operand reached along many
	// paths — a library call inlined into every part that uses it — would
	// otherwise be listed once per path: the 30 parts of the Fastener Matrix
	// came to 282,783 spans, 10,773 of them distinct, about 20 MB of JSON in
	// one message to the browser.
	seen map[edit.Span]bool
}

func newLocationCollector(module *program.Package) *locationCollector {
	return &locationCollector{
		module:  module,
		spans:   []edit.Span{},
		visited: make(map[string]bool),
		seen:    make(map[edit.Span]bool),
	}
}

// addSpan adds a source span if it has valid location info
func (c *locationCollector) addSpan(start, end *program.Location) {
	if start == nil || end == nil {
		return
	}
	span := edit.Span{
		StartLine:   int(start.Line),
		StartColumn: int(start.Column),
		EndLine:     int(end.Line),
		EndColumn:   int(end.Column),
	}
	if c.seen[span] {
		return
	}
	c.seen[span] = true
	c.spans = append(c.spans, span)
}

func packageAssignments(pkg *program.Package) []*program.Assignment {
	var list []*program.Assignment
	for _, file := range pkg.Files {
		if file == nil {
			continue
		}
		list = append(list, file.Assignments...)
	}
	return list
}

func packageFunctions(pkg *program.Package) []collectibleFunction {
	var list []collectibleFunction
	for _, file := range pkg.Files {
		if file == nil {
			continue
		}
		for _, ft := range file.FnTypes {
			if ft == nil {
				continue
			}
			for _, g := range ft.Groups {
				if g == nil {
					continue
				}
				for _, fn := range g.Functions {
					list = append(list, collectibleFunction{name: ft.Name, fn: fn})
				}
			}
		}
	}
	return list
}

func functionStartEnd(fn *program.Function) (*program.Location, *program.Location) {
	if fn == nil {
		return nil, nil
	}
	var start, end *program.Location
	if len(fn.Assignments) > 0 {
		start = fn.Assignments[0].Start
		end = fn.Assignments[len(fn.Assignments)-1].End
	}
	if len(fn.Returns) > 0 {
		if start == nil {
			start = fn.Returns[0].Start
		}
		end = fn.Returns[len(fn.Returns)-1].End
	}
	return start, end
}

// collectFromOperand recursively collects locations from an operand
func (c *locationCollector) collectFromOperand(op *program.Operand, assignments []*program.Assignment, functions []collectibleFunction) {
	if op == nil {
		return
	}

	// Add the operand's own location
	c.addSpan(op.Start, op.End)

	switch t := op.T.(type) {
	case *program.Operand_Reference:
		c.collectFromReference(t.Reference.Name, t.Reference.Prefix, assignments, functions)

	case *program.Operand_FnCall:
		c.collectFromFnCall(t.FnCall, assignments, functions)

	case *program.Operand_Expression:
		// Collect from all operands in the expression
		for _, operand := range t.Expression.Operands {
			c.collectFromOperand(operand, assignments, functions)
		}
	case *program.Operand_PropRef:
		refOp := &program.Operand{T: &program.Operand_Reference{Reference: t.PropRef.Reference}}
		c.collectFromOperand(refOp, assignments, functions)
	}
}

// collectFromReference finds a variable declaration and collects its locations
func (c *locationCollector) collectFromReference(name, prefix string, assignments []*program.Assignment, functions []collectibleFunction) {
	if prefix != "" {
		return
	}

	// Search local assignments first (in reverse order for shadowing)
	for i := len(assignments) - 1; i >= 0; i-- {
		assign := assignments[i]
		if assign == nil {
			continue
		}
		nameIndex := -1
		for idx, varInfo := range assign.Variables {
			if varInfo != nil && varInfo.Name == name {
				nameIndex = idx
				break
			}
		}
		if nameIndex == -1 {
			continue
		}

		vk := fmt.Sprintf("assign:%p:%d", assign, nameIndex)
		if c.visited[vk] {
			return
		}
		c.visited[vk] = true

		c.addSpan(assign.Start, assign.End)
		c.collectFromOperand(assign.Expression, assignments[:i], functions)
		return
	}

	// Search local functions next
	for _, fnInfo := range functions {
		if fnInfo.name == name {
			vk := fmt.Sprintf("fn:%p", fnInfo.fn)
			if c.visited[vk] {
				return
			}
			c.visited[vk] = true

			fStart, fEnd := functionStartEnd(fnInfo.fn)
			c.addSpan(fStart, fEnd)
			c.collectFromFunctionAllReturns(fnInfo.fn)
			return
		}
	}

	// Search package-level assignments across the whole package
	pkgAssignments := packageAssignments(c.module)
	for i := len(pkgAssignments) - 1; i >= 0; i-- {
		assign := pkgAssignments[i]
		if assign == nil {
			continue
		}
		nameIndex := -1
		for idx, varInfo := range assign.Variables {
			if varInfo != nil && varInfo.Name == name {
				nameIndex = idx
				break
			}
		}
		if nameIndex == -1 {
			continue
		}

		vk := fmt.Sprintf("assign:%p:%d", assign, nameIndex)
		if c.visited[vk] {
			return
		}
		c.visited[vk] = true

		c.addSpan(assign.Start, assign.End)
		c.collectFromOperand(assign.Expression, pkgAssignments[:i], packageFunctions(c.module))
		return
	}

	// Search package-level functions across the whole package
	for _, fnInfo := range packageFunctions(c.module) {
		if fnInfo.name == name {
			vk := fmt.Sprintf("fn:%p", fnInfo.fn)
			if c.visited[vk] {
				return
			}
			c.visited[vk] = true

			fStart, fEnd := functionStartEnd(fnInfo.fn)
			c.addSpan(fStart, fEnd)
			c.collectFromFunctionAllReturns(fnInfo.fn)
			return
		}
	}
}

// collectFromFnCall finds a function and collects from a specific return index.
func (c *locationCollector) collectFromFnCall(fc *program.FnCall, assignments []*program.Assignment, functions []collectibleFunction) {
	if fc.Reference.Prefix != "" {
		for _, arg := range fc.Arguments {
			c.collectFromOperand(arg, assignments, functions)
		}
		return
	}

	// Local or package-level function call
	var foundFn *program.Function
	for _, fnInfo := range functions {
		if fnInfo.name == fc.Reference.Name && len(fnInfo.fn.Parameters) == len(fc.Arguments) {
			foundFn = fnInfo.fn
			break
		}
	}
	if foundFn == nil {
		for _, fnInfo := range packageFunctions(c.module) {
			if fnInfo.name == fc.Reference.Name && len(fnInfo.fn.Parameters) == len(fc.Arguments) {
				foundFn = fnInfo.fn
				break
			}
		}
	}

	if foundFn != nil {
		vk := fmt.Sprintf("fncall:%p", foundFn)
		if !c.visited[vk] {
			c.visited[vk] = true
			fStart, fEnd := functionStartEnd(foundFn)
			c.addSpan(fStart, fEnd)
			c.collectFromFunctionAllReturns(foundFn)
		}
	}

	for _, arg := range fc.Arguments {
		c.collectFromOperand(arg, assignments, functions)
	}
}

// collectFromFunctionAllReturns collects locations from all returns in a function
func (c *locationCollector) collectFromFunctionAllReturns(fn *program.Function) {
	var innerFns []collectibleFunction
	for _, ft := range fn.FnTypes {
		if ft == nil {
			continue
		}
		for _, g := range ft.Groups {
			if g == nil {
				continue
			}
			for _, sub := range g.Functions {
				innerFns = append(innerFns, collectibleFunction{name: ft.Name, fn: sub})
			}
		}
	}
	for _, ret := range fn.Returns {
		c.collectFromOperand(ret, fn.Assignments, innerFns)
	}
}

func Render(
	source string,
	prog *program.Program,
) (*api.RenderResult, error) {
	if prog == nil || prog.Primary == nil {
		return nil, fmt.Errorf("nil program or primary package")
	}
	pkg := prog.Primary

	SupportedMajor := int32(0)
	SupportedMinor := int32(0)
	if pkg.VersionMajor != SupportedMajor || pkg.VersionMinor != SupportedMinor {
		return nil, fmt.Errorf("wgsl render version mismatch: %d.%d != %d.%d",
			pkg.VersionMajor, pkg.VersionMinor, SupportedMajor, SupportedMinor)
	}

	var returns []*program.Operand
	var assignments []*program.Assignment
	if len(pkg.Files) > 0 {
		lastFile := pkg.Files[len(pkg.Files)-1]
		returns = lastFile.Returns
		assignments = lastFile.Assignments
	}

	result := &api.RenderResult{}

	// Extract all source locations and names for each shape by traversing the AST
	shapeLocations := make([]api.ShapeLocations, len(returns))
	shapeNames := make([]string, len(returns))
	for i, ret := range returns {
		collector := newLocationCollector(pkg)
		collector.collectFromOperand(ret, assignments, packageFunctions(pkg))
		shapeLocations[i] = api.ShapeLocations{Spans: collector.spans}
		shapeNames[i] = program.OperandName(ret, source)
	}

	result.ShapeCount = len(returns)
	result.ShapeNames = shapeNames
	result.ShapeLocations = shapeLocations

	wgslCode, meshSdfSlots, parts, spinodoidSlots, err := codegen.WGSLWithParts(prog)
	if err != nil {
		return nil, fmt.Errorf("build WGSL failed:\n%w", err)
	}
	result.WGSL = wgslCode
	result.MeshSdfSlots = meshSdfSlots
	result.SpinodoidSlots = spinodoidSlots
	result.Parts = parts

	return result, nil
}

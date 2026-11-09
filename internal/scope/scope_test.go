package scope_test

import (
	"fmt"
	"testing"

	"gradientcontrol.com/omega/core/internal/scope"
	"gradientcontrol.com/omega/core/program"
)

func TestScope_CoreNavigation(t *testing.T) {
	// Prepare a simple mock Program structure
	prog := &program.Program{
		Primary: &program.Package{
			Path: "",
			Files: []*program.PackageFile{
				{
					Path: "main.gsl",
					Assignments: []*program.Assignment{
						{
							Variables: []*program.Variable{
								{Name: "globalVar"},
							},
						},
					},
					FnTypes: []*program.FnType{
						{
							Name: "testFn",
							Groups: []*program.FnGroup{
								{
									ParamCount: 1,
									Functions: []*program.Function{
										{
											Parameters: []*program.Parameter{
												{Name: "param1"},
											},
											Assignments: []*program.Assignment{
												{
													Variables: []*program.Variable{
														{Name: "localVar"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// 1. Initialize ProgScope
	progScope, err := scope.Init(prog)
	if err != nil {
		t.Fatalf("failed to init program scope: %v", err)
	}

	// 2. Navigation assertions
	if progScope.ProgScope() != progScope {
		t.Errorf("ProgScope.ProgScope() expected %v, got %v", progScope, progScope.ProgScope())
	}

	primaryPkg := progScope.Primary
	if primaryPkg == nil {
		t.Fatal("expected primary package scope to be present")
	}
	if primaryPkg.ContainingScope() != nil {
		t.Errorf("PkgScope.ContainingScope() should be nil, got %v", primaryPkg.ContainingScope())
	}

	fileScope := primaryPkg.Files[0]
	if fileScope.ContainingScope() != primaryPkg {
		t.Errorf("file scope containing scope expected %v, got %v", primaryPkg, fileScope.ContainingScope())
	}

	// 3. Symbol resolution assertions
	// Resolve globalVar from file scope
	elem := fileScope.NamedElem("globalVar")
	if elem == nil {
		t.Fatal("expected to find 'globalVar' in file scope")
	}
	if elem.Name() != "globalVar" {
		t.Errorf("expected symbol name 'globalVar', got %q", elem.Name())
	}
	if elem.ContainingScope() != fileScope {
		t.Errorf("expected globalVar containing scope to be fileScope, got %v", elem.ContainingScope())
	}

	// Traverse to function scope
	fnType := fileScope.Functions["testFn"]
	if fnType == nil {
		t.Fatal("expected function testFn to exist")
	}
	fnGroup := fnType.Groups[1]
	if fnGroup == nil {
		t.Fatal("expected fn group with arity 1 to exist")
	}
	fnScope := fnGroup.Functions[0]

	// Resolve local variable inside function
	localElem := fnScope.NamedElem("localVar")
	if localElem == nil {
		t.Fatal("expected to find 'localVar' inside fnScope")
	}
	if localElem.ContainingScope() != fnScope {
		t.Errorf("expected localVar containing scope to be fnScope, got %v", localElem.ContainingScope())
	}

	// Resolve parameter inside function
	paramElem := fnScope.NamedElem("param1")
	if paramElem == nil {
		t.Fatal("expected to find parameter 'param1'")
	}

	// Resolve outer variable from function scope (delegation)
	outerElem := fnScope.NamedElem("globalVar")
	if outerElem == nil {
		t.Fatal("expected to resolve 'globalVar' from fnScope")
	}
	if outerElem.ContainingScope() != fileScope {
		t.Errorf("expected resolved 'globalVar' containing scope to be fileScope, got %v", outerElem.ContainingScope())
	}
}

func TestScope_BuilderInterface(t *testing.T) {
	// Construct an empty package file AST
	pkgFile := &program.PackageFile{}
	pkgScope := scope.NewPkgScope(nil, &program.Package{Path: "test_pkg"})
	fileScope := pkgScope.NewPkgFileScope(pkgFile)

	// Ensure fileScope implements BuilderScope
	var bScope scope.BuilderScope = fileScope

	// Test AddAssignment
	loc := &program.Location{Line: 1, Column: 1}
	err := bScope.AddAssignment([]string{"x"}, &program.Operand{Start: loc}, loc, loc, nil)
	if err != nil {
		t.Fatalf("failed to add assignment: %v", err)
	}

	// Verify symbol exists
	xElem := bScope.NamedElem("x")
	if xElem == nil {
		t.Fatal("expected variable 'x' to be registered")
	}
}

func TestScope_Advanced(t *testing.T) {
	prog := &program.Program{
		Primary: &program.Package{Path: "main"},
	}
	progScope, err := scope.Init(prog)
	if err != nil {
		t.Fatalf("failed to init: %v", err)
	}

	// 1. Test AddDependency
	depPkg := &program.Package{
		Path: "dep",
		Files: []*program.PackageFile{
			{
				Path: "dep.gsl",
				Imports: []*program.Import{
					{Path: "dep"},
				},
			},
		},
	}
	depScope, err := progScope.AddDependency(depPkg)
	if err != nil {
		t.Fatalf("failed to add dependency: %v", err)
	}
	if progScope.Dependencies["dep"] != depScope {
		t.Error("expected dependency scope to be registered")
	}

	// Test adding existing dependency (idempotence)
	depScope2, err := progScope.AddDependency(depPkg)
	if err != nil || depScope2 != depScope {
		t.Errorf("expected idempotence on AddDependency, got %v, error: %v", depScope2, err)
	}

	// 2. Test ErrLoc & Scope.Err
	fileScope := progScope.Primary.NewPkgFileScope(&program.PackageFile{Path: "main.gsl"})
	loc := &program.Location{Line: 42, Column: 7}
	testErr := fmt.Errorf("some error")
	err1 := fileScope.Err(loc, testErr)
	if err1 == nil || err1.Error() == "" {
		t.Errorf("expected error, got %v", err1)
	}

	err2 := progScope.Err(loc, testErr)
	if err2 == nil {
		t.Error("expected progScope.Err to return non-nil error")
	}

	// 3. Test ResolvePaths
	// Create mock scopes/elements to test ResolvePaths
	v := &program.Variable{Name: "v"}
	assignScope := &scope.AssignScope{Parent: fileScope}
	elemV := &scope.Var{Parent: assignScope, Var: v}

	// Reference in same file scope
	pkg1, ref1 := scope.ResolvePaths(fileScope, elemV)
	if pkg1 != "" || ref1 != "v" {
		t.Errorf("ResolvePaths same scope failed: pkg=%q, ref=%q", pkg1, ref1)
	}

	// Reference in different package
	depFileScope := depScope.NewPkgFileScope(&program.PackageFile{Path: "dep.gsl"})
	pkg2, ref2 := scope.ResolvePaths(depFileScope, elemV)
	if pkg2 != "main" || ref2 != "v" {
		t.Errorf("ResolvePaths cross package failed: pkg=%q, ref=%q", pkg2, ref2)
	}

	// 4. Test AddPropertyAssignment validations
	fileScope.AddAssignment([]string{"y"}, &program.Operand{Start: loc}, loc, loc, nil)
	err = fileScope.AddPropertyAssignment("y", "min", &program.Operand{}, loc, loc)
	if err != nil {
		t.Errorf("expected property min to be assigned successfully, got %v", err)
	}
	// Test duplicate property
	err = fileScope.AddPropertyAssignment("y", "min", &program.Operand{}, loc, loc)
	if err == nil {
		t.Error("expected error when assigning duplicate property")
	}
	// Test invalid property name
	err = fileScope.AddPropertyAssignment("y", "invalidProp", &program.Operand{}, loc, loc)
	if err == nil {
		t.Error("expected error when assigning invalid property name")
	}
	// Test unknown variable property assignment
	err = fileScope.AddPropertyAssignment("z", "min", &program.Operand{}, loc, loc)
	if err == nil {
		t.Error("expected error when assigning property to unknown variable")
	}

	// 5. Test expression depth limits
	for i := 0; i < 127; i++ {
		if err := fileScope.IncExprDepth(); err != nil {
			t.Fatalf("unexpected error at depth %d: %v", i, err)
		}
	}
	if err := fileScope.IncExprDepth(); err == nil {
		t.Error("expected error when exceeding max expression depth")
	}
	fileScope.DecExprDepth()

	// 6. Test function creation validations
	// Duplicate parameter names
	_, err = fileScope.NewFnScope("myFn", []string{"a", "a"})
	if err == nil {
		t.Error("expected error with duplicate parameter names")
	}

	// Reassigning built-in names
	_, err = fileScope.NewFnScope("p", []string{"a"})
	if err == nil {
		t.Error("expected error when reassigning 'p'")
	}

	// Adding overload with same parameter count
	fnScope, err := fileScope.NewFnScope("myFn", []string{"a"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err = fileScope.NewFnScope("myFn", []string{"b"})
	if err == nil {
		t.Error("expected error when adding function with same arity")
	}

	// 7. Test invalid FnScope builder actions
	err = fnScope.AddImport(&program.Import{Path: "foo"})
	if err == nil {
		t.Error("expected error when adding import to FnScope")
	}
	err = fnScope.AddMeshImport(&program.MeshImport{Path: "foo"})
	if err == nil {
		t.Error("expected error when adding mesh import to FnScope")
	}
	err = fnScope.AddStepImport(&program.StepImport{Path: "foo"})
	if err == nil {
		t.Error("expected error when adding step import to FnScope")
	}
	err = fnScope.AddWarpImport(&program.WarpImport{Path: "foo"})
	if err == nil {
		t.Error("expected error when adding warp import to FnScope")
	}

	// Verify FnScope Inc/Dec expression depth delegation
	if err := fnScope.IncExprDepth(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	fnScope.DecExprDepth()

	// 8. Test imports and package references
	err = fileScope.AddImport(&program.Import{Path: "lib"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Duplicate import
	err = fileScope.AddImport(&program.Import{Path: "lib"})
	if err == nil {
		t.Error("expected error on duplicate import")
	}
	// Import naming collision with variable
	err = fileScope.AddImport(&program.Import{Path: "y"})
	if err == nil {
		t.Error("expected error on import name collision with variable")
	}

	// Add package reference
	fileScope.AddImportRef("myRef", depScope)
	refElem := fileScope.NamedElem("myRef")
	if refElem == nil {
		t.Fatal("expected 'myRef' import reference to exist")
	}
	pkgRef, ok := refElem.(*scope.PkgRef)
	if !ok {
		t.Fatalf("expected PkgRef type, got %T", refElem)
	}
	if pkgRef.PkgScopePtr() != depScope {
		t.Error("expected package scope pointer to match")
	}
	pkgRef.SetPkgScopePtr(depScope)

	// Navigate containing scopes from PkgRef
	if pkgRef.ContainingScope() != fileScope {
		t.Error("PkgRef containing scope must be fileScope")
	}
	if pkgRef.Assignments() != nil {
		t.Error("PkgRef assignments must be nil")
	}
	if pkgRef.ProgScope() != progScope {
		t.Error("PkgRef prog scope mismatch")
	}
	if pkgRef.PkgScope() != depScope {
		t.Error("PkgRef pkg scope mismatch")
	}

	// 9. Test other AST levels containing scopes
	fnTypeScope := fileScope.NamedElem("myFn").(*scope.FnTypeScope)
	if fnTypeScope.ContainingScope() != fileScope {
		t.Error("FnTypeScope containing scope mismatch")
	}
	if fnTypeScope.Assignments() != nil {
		t.Error("FnTypeScope assignments must be nil")
	}
	if fnTypeScope.ProgScope() != progScope {
		t.Error("FnTypeScope prog scope mismatch")
	}
	if fnTypeScope.PkgScope() != progScope.Primary {
		t.Error("FnTypeScope pkg scope mismatch")
	}
	if fnTypeScope.PkgFileScope() != fileScope {
		t.Error("FnTypeScope pkg file scope mismatch")
	}
	if fnTypeScope.Err(loc, testErr) == nil {
		t.Error("FnTypeScope error generation failed")
	}

	fnGroupScope := fnTypeScope.Groups[1]
	if fnGroupScope.ContainingScope() != fileScope {
		t.Error("FnGroupScope containing scope mismatch")
	}
	if fnGroupScope.Assignments() != nil {
		t.Error("FnGroupScope assignments must be nil")
	}
	if fnGroupScope.ProgScope() != progScope {
		t.Error("FnGroupScope prog scope mismatch")
	}
	if fnGroupScope.PkgScope() != progScope.Primary {
		t.Error("FnGroupScope pkg scope mismatch")
	}
	if fnGroupScope.PkgFileScope() != fileScope {
		t.Error("FnGroupScope pkg file scope mismatch")
	}
	if fnGroupScope.Err(loc, testErr) == nil {
		t.Error("FnGroupScope error generation failed")
	}

	// 10. Test Slot Import validations
	err = fileScope.AddMeshImport(&program.MeshImport{Path: "m1", Index: 1})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Duplicate mesh path
	err = fileScope.AddMeshImport(&program.MeshImport{Path: "m1", Index: 2})
	if err == nil {
		t.Error("expected error with duplicate mesh path")
	}
	// Duplicate mesh slot index
	err = fileScope.AddMeshImport(&program.MeshImport{Path: "m2", Index: 1})
	if err == nil {
		t.Error("expected error with duplicate mesh slot index")
	}

	err = fileScope.AddStepImport(&program.StepImport{Path: "s1", Index: 3})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Duplicate step path
	err = fileScope.AddStepImport(&program.StepImport{Path: "s1", Index: 4})
	if err == nil {
		t.Error("expected error with duplicate step path")
	}
	// Duplicate step slot index
	err = fileScope.AddStepImport(&program.StepImport{Path: "s2", Index: 3})
	if err == nil {
		t.Error("expected error with duplicate step slot index")
	}

	err = fileScope.AddWarpImport(&program.WarpImport{Path: "w1", Index: 5})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	// Duplicate warp path
	err = fileScope.AddWarpImport(&program.WarpImport{Path: "w1", Index: 6})
	if err == nil {
		t.Error("expected error with duplicate warp path")
	}
	// Duplicate warp slot index
	err = fileScope.AddWarpImport(&program.WarpImport{Path: "w2", Index: 5})
	if err == nil {
		t.Error("expected error with duplicate warp slot index")
	}

	// 11. Test FnScope building a nested function
	nestedFnScope, err := fnScope.NewFnScope("nested", []string{"p1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if nestedFnScope.ContainingScope() != fnScope {
		t.Errorf("expected nested function containing scope to be fnScope, got %v", nestedFnScope.ContainingScope())
	}

	// Test Call helper
	callObj := fnScope.(*scope.FnScope).CallFrom(&program.FnCall{})
	if callObj == nil || callObj.Name() != "myFn" {
		t.Errorf("expected call name 'myFn', got %v", callObj)
	}

	// 12. Test FnScope builder assignments and property assignments
	fnConcrete := fnScope.(*scope.FnScope)
	err = fnScope.AddAssignment([]string{"fnLocal"}, &program.Operand{Start: loc}, loc, loc, nil)
	if err != nil {
		t.Errorf("unexpected error on FnScope.AddAssignment: %v", err)
	}
	if len(fnConcrete.Assignments()) != 1 {
		t.Errorf("expected 1 assignment in function, got %d", len(fnConcrete.Assignments()))
	}
	err = fnScope.AddPropertyAssignment("fnLocal", "max", &program.Operand{}, loc, loc)
	if err != nil {
		t.Errorf("unexpected error on FnScope.AddPropertyAssignment: %v", err)
	}
	// Verify parameters and error helper
	if fnConcrete.Parameters["a"] == nil {
		t.Error("expected parameter 'a' to exist")
	}
	paramObj := fnConcrete.Parameters["a"]
	if paramObj.Name() != "a" || paramObj.ContainingScope() != fnScope {
		t.Error("invalid parameter helper methods")
	}
	if fnScope.Err(loc, testErr) == nil {
		t.Error("expected fnScope.Err to return non-nil error")
	}
	if fnConcrete.PkgScope() != progScope.Primary {
		t.Error("FnScope.PkgScope() mismatch")
	}
	if fnConcrete.ProgScope() != progScope {
		t.Error("FnScope.ProgScope() mismatch")
	}

	// 13. Test fallback read-only methods on scopes
	if fileScope.Assignments() == nil {
		t.Error("PkgFileScope.Assignments() should not be nil")
	}
	if progScope.PkgFileScope() != nil || progScope.PkgScope() != nil || progScope.Assignments() != nil || progScope.NamedElem("x") != nil {
		t.Error("invalid ProgScope fallback query results")
	}
	if progScope.Primary.ContainingScope() != nil || progScope.Primary.PkgFileScope() != nil || progScope.Primary.Assignments() != nil || progScope.Primary.NamedElem("nonexistent") != nil || progScope.Primary.Err(loc, testErr) == nil {
		t.Error("invalid PkgScope fallback query results")
	}
	if progScope.Primary.PkgFileScope() != nil || progScope.Primary.PkgScope() != progScope.Primary || progScope.Primary.ProgScope() != progScope {
		t.Error("invalid PkgScope helper methods")
	}

	// 14. Test PkgRef queries and containing scopes
	if pkgRef.PkgFileScope() != fileScope || pkgRef.Name() != "myRef" || pkgRef.NamedElem("nonexistent") != nil || pkgRef.Err(loc, testErr) == nil {
		t.Error("invalid PkgRef helper methods")
	}
	// Check ContainingScope of Call
	if callObj.ContainingScope() != fileScope {
		t.Error("Call containing scope mismatch")
	}

	// 15. SetReturnOperands
	fnScope.SetReturnOperands([]*program.Operand{})
	fileScope.SetReturnOperands([]*program.Operand{})

	// 16. Test AgentScope getters/setters on all scopes
	var dummyAgent scope.AgentScope
	progScope.SetAgentScope(dummyAgent)
	if progScope.GetAgentScope() != dummyAgent {
		t.Error("ProgScope GetAgentScope mismatch")
	}
	progScope.Primary.SetAgentScope(dummyAgent)
	if progScope.Primary.GetAgentScope() != dummyAgent {
		t.Error("PkgScope GetAgentScope mismatch")
	}
	fileScope.SetAgentScope(dummyAgent)
	if fileScope.GetAgentScope() != dummyAgent {
		t.Error("PkgFileScope GetAgentScope mismatch")
	}
	fnScope.SetAgentScope(dummyAgent)
	if fnScope.GetAgentScope() != dummyAgent {
		t.Error("FnScope GetAgentScope mismatch")
	}
	if fnTypeScope.GetAgentScope() != nil {
		t.Error("FnTypeScope GetAgentScope must be nil")
	}
	fnTypeScope.SetAgentScope(dummyAgent)
	if fnGroupScope.GetAgentScope() != nil {
		t.Error("FnGroupScope GetAgentScope must be nil")
	}
	fnGroupScope.SetAgentScope(dummyAgent)
	if pkgRef.GetAgentScope() != nil {
		t.Error("PkgRef GetAgentScope must be nil")
	}
	pkgRef.SetAgentScope(dummyAgent)
}

func TestScope_PropertySynthesis(t *testing.T) {
	pkg := &program.Package{
		Files: []*program.PackageFile{{}},
	}
	prog := &program.Program{Primary: pkg}
	progScope, err := scope.Init(prog)
	if err != nil {
		t.Fatalf("scope.Init failed: %v", err)
	}
	fileScope := progScope.Primary.Files[0]

	loc := &program.Location{Line: 5, Column: 10}
	exprOp := program.NewValue(5.0, loc, loc)

	// Variable "r" with min, max, period
	err = fileScope.AddAssignment([]string{"r"}, exprOp, loc, loc, nil)
	if err != nil {
		t.Fatalf("AddAssignment failed: %v", err)
	}

	minOp := program.NewValue(1.0, loc, loc)
	maxOp := program.NewValue(10.0, loc, loc)
	periodOp := program.NewValue(2.0, loc, loc)

	// Add min
	err = fileScope.AddPropertyAssignment("r", "min", minOp, loc, loc)
	if err != nil {
		t.Fatalf("AddPropertyAssignment min failed: %v", err)
	}
	assign := fileScope.File.Assignments[0]
	if assign.Range != nil {
		t.Errorf("expected Range to be nil when only min is present")
	}

	// Add max -> synthesis should trigger
	err = fileScope.AddPropertyAssignment("r", "max", maxOp, loc, loc)
	if err != nil {
		t.Fatalf("AddPropertyAssignment max failed: %v", err)
	}
	if assign.Range == nil {
		t.Fatalf("expected Range to be synthesized when both min and max are present")
	}
	if !assign.Range.IsSynthetic() {
		t.Errorf("expected synthesized range to report IsSynthetic() == true")
	}
	if minVal, _ := assign.Range.MinConst(); minVal != 1.0 {
		t.Errorf("expected MinConst 1.0, got %v", minVal)
	}
	if maxVal, _ := assign.Range.MaxConst(); maxVal != 10.0 {
		t.Errorf("expected MaxConst 10.0, got %v", maxVal)
	}
	if _, ok := assign.Range.PeriodConst(); ok {
		t.Errorf("expected PeriodConst to be absent before period property is added")
	}

	// Add period -> updates synthesized Range
	err = fileScope.AddPropertyAssignment("r", "period", periodOp, loc, loc)
	if err != nil {
		t.Fatalf("AddPropertyAssignment period failed: %v", err)
	}
	if pVal, ok := assign.Range.PeriodConst(); !ok || pVal != 2.0 {
		t.Errorf("expected PeriodConst 2.0, got (%v, %v)", pVal, ok)
	}

	// Explicit Range should NOT be overwritten by property synthesis
	explicitRange := &program.Range{
		Start: &program.Location{Line: 1, Column: 1},
		End:   &program.Location{Line: 1, Column: 20},
		Min:   program.NewValue(0.1, loc, loc),
		Max:   program.NewValue(0.9, loc, loc),
	}
	err = fileScope.AddAssignment([]string{"explicitVar"}, exprOp, loc, loc, explicitRange)
	if err != nil {
		t.Fatalf("AddAssignment with range failed: %v", err)
	}
	explicitAssign := fileScope.File.Assignments[1]
	if explicitAssign.Range.IsSynthetic() {
		t.Errorf("expected explicit range to NOT be synthetic")
	}

	_ = fileScope.AddPropertyAssignment("explicitVar", "min", minOp, loc, loc)
	_ = fileScope.AddPropertyAssignment("explicitVar", "max", maxOp, loc, loc)
	if explicitAssign.Range != explicitRange {
		t.Errorf("expected explicit RangeAnnotation to not be overwritten by property assignments")
	}
}

func TestScope_SetParam(t *testing.T) {
	pkg := &program.Package{
		Files: []*program.PackageFile{{}},
	}
	prog := &program.Program{Primary: pkg}
	progScope, err := scope.Init(prog)
	if err != nil {
		t.Fatalf("scope.Init failed: %v", err)
	}
	fileScope := progScope.Primary.Files[0]

	fnScope, err := fileScope.NewFnScope("testFn", []string{"a", "b"})
	if err != nil {
		t.Fatalf("NewFnScope failed: %v", err)
	}

	locA := &program.Location{Line: 10, Column: 5}
	locB := &program.Location{Line: 10, Column: 15}
	r := &program.Range{
		Start: locB,
		End:   &program.Location{Line: 10, Column: 30},
		Min:   program.NewValue(0.5, locB, locB),
		Max:   program.NewValue(5.0, locB, locB),
	}

	fnScope.SetParam(0, locA, nil)
	fnScope.SetParam(1, locB, r)

	fns := fnScope.(*scope.FnScope)
	if fns.Function.Parameters[0].Start != locA {
		t.Errorf("expected parameter a Start to be set to locA")
	}
	if fns.Function.Parameters[0].Range != nil {
		t.Errorf("expected parameter a Range to be nil")
	}
	if fns.Function.Parameters[1].Start != locB {
		t.Errorf("expected parameter b Start to be set to locB")
	}
	if fns.Function.Parameters[1].Range != r {
		t.Errorf("expected parameter b Range to be set to r")
	}
}

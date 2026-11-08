// Package variation constructs self-contained sub-programs and prunes
// primary programs over a shared assembled *program.Program.
package variation

import (
	"gradientcontrol.com/omega/core/internal/assembler/deadcode"
	"gradientcontrol.com/omega/core/program"
	"gradientcontrol.com/omega/core/program/textures"
)

// Program builds the self-contained program that evaluates the given returns
// over prog: prog's primary package with its mesh-typed and unreachable
// declarations pruned and the returns in place of its own, with reachable
// dependencies, source text, and texture slots preserved.
func Program(prog *program.Program, returns ...*program.Operand) *program.Program {
	var extraRoots []*program.Operand
	if len(returns) == 0 {
		extraRoots = ControlRoots(prog.Primary)
	}
	return ProgramWithExtraRoots(prog, returns, extraRoots)
}

// ProgramWithExtraRoots builds a self-contained variation program for returns
// while treating extraRoots as additional liveness roots during dead-code pruning.
func ProgramWithExtraRoots(prog *program.Program, returns, extraRoots []*program.Operand) *program.Program {
	module := prog.Primary
	deadcodeReturns := returns
	if len(extraRoots) > 0 {
		deadcodeReturns = make([]*program.Operand, 0, len(returns)+len(extraRoots))
		deadcodeReturns = append(deadcodeReturns, returns...)
		deadcodeReturns = append(deadcodeReturns, extraRoots...)
	}
	var files []*program.PackageFile
	for i, f := range module.Files {
		var fReturns []*program.Operand
		if i == len(module.Files)-1 {
			fReturns = deadcodeReturns
		}
		files = append(files, &program.PackageFile{
			Path:        f.Path,
			Source:      f.Source,
			Imports:     f.Imports,
			Assignments: pruneMeshTypedAssignments(f.Assignments),
			FnTypes:     f.FnTypes,
			Returns:     fReturns,
			MeshImports: f.MeshImports,
			StepImports: f.StepImports,
			WarpImports: f.WarpImports,
		})
	}

	out := &program.Program{
		VersionMajor: prog.VersionMajor,
		VersionMinor: prog.VersionMinor,
		Primary: &program.Package{
			VersionMajor: module.VersionMajor,
			VersionMinor: module.VersionMinor,
			Path:         module.Path,
			Files:        files,
			Scene:        program.CloneScene(module.Scene),
		},
	}
	if len(prog.Dependencies) > 0 {
		out.Dependencies = append([]*program.Package(nil), prog.Dependencies...)
	}
	// Invariant: module.Files is guaranteed non-empty by parser/assembler invariants.
	if pruned, err := deadcode.Remove(out); err == nil {
		out = pruned
	}
	out.SetReturns(returns)
	out.MeshSdfSlots, out.GlobalTextures = textures.Extract(out)
	return out
}

// PrunePrimary prunes declarations in prog that are unreachable from prog's
// primary returns, ColorReturns, MaterialReturns, and UI control roots, while
// preserving ColorReturns, MaterialReturns, and PackageFile.Source, and
// refreshes MeshSdfSlots and GlobalTextures.
func PrunePrimary(prog *program.Program) *program.Program {
	module := prog.Primary
	returns := prog.Returns()
	colorReturns := prog.ColorReturns()
	materialReturns := prog.MaterialReturns()
	controlRoots := ControlRoots(module)

	deadcodeReturns := make([]*program.Operand, 0, len(returns)+len(colorReturns)+len(materialReturns)+len(controlRoots))
	deadcodeReturns = append(deadcodeReturns, returns...)
	for _, c := range colorReturns {
		if c != nil {
			deadcodeReturns = append(deadcodeReturns, c)
		}
	}
	for _, m := range materialReturns {
		if m != nil {
			deadcodeReturns = append(deadcodeReturns, m)
		}
	}
	deadcodeReturns = append(deadcodeReturns, controlRoots...)

	var files []*program.PackageFile
	for i, f := range module.Files {
		var fReturns []*program.Operand
		if i == len(module.Files)-1 {
			fReturns = deadcodeReturns
		}
		files = append(files, &program.PackageFile{
			Path:            f.Path,
			Source:          f.Source,
			Imports:         f.Imports,
			Assignments:     pruneMeshTypedAssignments(f.Assignments),
			FnTypes:         f.FnTypes,
			Returns:         fReturns,
			ColorReturns:    f.ColorReturns,
			MaterialReturns: f.MaterialReturns,
			MeshImports:     f.MeshImports,
			StepImports:     f.StepImports,
			WarpImports:     f.WarpImports,
		})
	}

	out := &program.Program{
		VersionMajor: prog.VersionMajor,
		VersionMinor: prog.VersionMinor,
		Primary: &program.Package{
			VersionMajor: module.VersionMajor,
			VersionMinor: module.VersionMinor,
			Path:         module.Path,
			Files:        files,
			Scene:        program.CloneScene(module.Scene),
		},
	}
	if len(prog.Dependencies) > 0 {
		out.Dependencies = append([]*program.Package(nil), prog.Dependencies...)
	}
	// Invariant: module.Files is guaranteed non-empty by parser/assembler invariants.
	if pruned, err := deadcode.Remove(out); err == nil {
		out = pruned
	}
	out.SetReturns(returns)
	lastIdx := len(out.Primary.Files) - 1
	out.Primary.Files[lastIdx].ColorReturns = colorReturns
	out.Primary.Files[lastIdx].MaterialReturns = materialReturns
	out.MeshSdfSlots, out.GlobalTextures = textures.Extract(out)
	return out
}

// ControlRoots returns references to top-level range sliders and UI control
// assignments so dead-code pruning keeps their declarations aligned with the
// viewer's uniform parameter list.
func ControlRoots(pkg *program.Package) []*program.Operand {
	// Invariant: pkg, its Files, Assignments, and Variables are guaranteed non-nil by parser/assembler invariants.
	var roots []*program.Operand
	for _, f := range pkg.Files {
		for _, a := range f.Assignments {
			keep := a.Range != nil || containsUIControlCall(a.Expression)
			if !keep {
				continue
			}
			for _, v := range a.Variables {
				if v.Name != "" {
					roots = append(roots, &program.Operand{
						T: &program.Operand_Reference{
							Reference: &program.Reference{Name: v.Name},
						},
					})
				}
			}
			if a.Range != nil {
				if a.Range.Min != nil {
					roots = append(roots, a.Range.Min)
				}
				if a.Range.Max != nil {
					roots = append(roots, a.Range.Max)
				}
				if a.Range.Period != nil {
					roots = append(roots, a.Range.Period)
				}
			}
		}
	}
	return roots
}

func containsUIControlCall(op *program.Operand) bool {
	// Invariant: op and its oneof payload (FnCall, Expression) are guaranteed non-nil in valid ASTs.
	switch v := op.T.(type) {
	case *program.Operand_FnCall:
		if ref := v.FnCall.Reference; ref != nil && ref.Prefix == "" {
			switch ref.Name {
			case "uiPoint", "uiTransform", "uiColor", "uiDirection", "uiSpline":
				return true
			}
		}
		for _, arg := range v.FnCall.Arguments {
			if containsUIControlCall(arg) {
				return true
			}
		}
	case *program.Operand_Expression:
		for _, child := range v.Expression.Operands {
			if containsUIControlCall(child) {
				return true
			}
		}
	}
	return false
}

func pruneMeshTypedAssignments(assignments []*program.Assignment) []*program.Assignment {
	// Invariant: Assignment and Variable entries are guaranteed non-nil by parser/assembler invariants.
	meshTyped := make(map[string]bool)
	for changed := true; changed; {
		changed = false
		for _, a := range assignments {
			if len(a.Variables) != 1 {
				continue
			}
			name := a.Variables[0].Name
			if meshTyped[name] {
				continue
			}
			if operandReferencesMesh(a.Expression, meshTyped) {
				meshTyped[name] = true
				changed = true
			}
		}
	}
	if len(meshTyped) == 0 {
		return assignments
	}
	kept := make([]*program.Assignment, 0, len(assignments))
	for _, a := range assignments {
		if len(a.Variables) == 1 && meshTyped[a.Variables[0].Name] {
			continue
		}
		kept = append(kept, a)
	}
	return kept
}

func operandReferencesMesh(op *program.Operand, meshTyped map[string]bool) bool {
	// Invariant: op and its oneof payload (Reference, FnCall, Expression) are guaranteed non-nil in valid ASTs.
	switch v := op.T.(type) {
	case *program.Operand_Reference:
		return meshTyped[v.Reference.Name]
	case *program.Operand_FnCall:
		fc := v.FnCall
		if fc.Reference.Prefix == "" {
			switch fc.Reference.Name {
			case "meshUnion", "meshIntersect", "meshDifference", "mesh":
				return true
			}
		}
		for _, arg := range fc.Arguments {
			if operandReferencesMesh(arg, meshTyped) {
				return true
			}
		}
	case *program.Operand_Expression:
		for _, child := range v.Expression.Operands {
			if operandReferencesMesh(child, meshTyped) {
				return true
			}
		}
	}
	return false
}

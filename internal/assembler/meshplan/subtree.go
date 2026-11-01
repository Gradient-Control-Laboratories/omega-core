package meshplan

import (
	"gradientcontrol.com/omega/core/internal/assembler/textures"
	"gradientcontrol.com/omega/core/program"
)

// SdfSubtree is an analytical leaf of the plan: the SDF expression the slicer
// carved out of the program's returns, over the program's own assignments.
type SdfSubtree struct {
	// ID is the unique correlation key matching MeshPlanNode.SubtreeID in the export plan.
	// Asynchronous meshing responses from Fidget use this key to route resulting STL URLs
	// back to the matching leaf node in the plan tree.
	ID string

	// Return is the leaf's expression. It references the sliced program's
	// assignments by name, so it is bytes where the program is megabytes, and
	// a response naming N leaves carries the program once rather than N
	// times. SubtreeProgram rebuilds the program that evaluates it.
	Return *program.Operand
}

// SubtreeProgram builds the self-contained program that evaluates the given
// returns over prog: prog's primary package with its mesh-typed assignments
// pruned and the returns in place of its own, sharing prog's dependencies,
// with the texture slots the WGSL backend reads extracted. One return makes a
// plan leaf's program for meshing; several make the live preview's.
func SubtreeProgram(prog *program.Program, returns ...*program.Operand) *program.Program {
	module := prog.Primary
	var files []*program.PackageFile
	for i, f := range module.Files {
		var fReturns []*program.Operand
		if i == 0 {
			fReturns = returns
		}
		files = append(files, &program.PackageFile{
			Path:        f.Path,
			Imports:     f.Imports,
			Assignments: pruneMeshTypedAssignments(f.Assignments),
			FnTypes:     f.FnTypes,
			Returns:     fReturns,
		})
	}

	out := &program.Program{
		Primary: &program.Package{
			VersionMajor: module.VersionMajor,
			VersionMinor: module.VersionMinor,
			Files:        files,
			Scene:        program.CloneScene(module.Scene),
		},
	}
	if len(prog.Dependencies) > 0 {
		out.Dependencies = append([]*program.Package(nil), prog.Dependencies...)
	}
	out.MeshSdfSlots, out.GlobalTextures = textures.Extract(out)
	return out
}

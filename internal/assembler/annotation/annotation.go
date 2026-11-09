// Package annotation reads the declaration-level annotations the parser records
// on a package file, and holds each one to the contract its kind carries.
//
// @material marks a browsable surface, and a surface is a material record and
// an albedo together: a record carries no colour of its own, so neither half is
// usable alone. Every @material is therefore a FUNCTION returning the pair —
// `metal.brass()` gives the record and the colour in one return — and Validate
// rejects an annotated constant, which has no way to carry both. ValidateReturns
// then checks the shape of the pair for the forms it can type.
//
// An earlier design paired them by name instead: every `X` had an `XColor`, and
// physical added an `XDensity`. That made the contract depend on spelling, and
// nothing could catch a consumer reaching for the wrong `XColor`. Returning the
// pair makes the mismatch unrepresentable rather than merely checked.
//
// @finish marks the other side of that pairing: the material library's own
// presets, which carry no albedo on purpose because they exist to be combined
// with a colour from elsewhere. Those are constants, and keeping them a separate
// kind is what lets the rule above stay absolute — it also tells the browser
// which swatches need a stand-in colour to be drawn at all.
//
// Unlike @point and @transform, @material is not lowered into the expression
// tree (see buildRecordedAnnotation in the parser), so nothing here runs during
// codegen and an annotation cannot change what its symbol compiles to.
package annotation

import (
	"fmt"

	"gradientcontrol.com/omega/core/program"
)

const (
	// Material marks a function the material browser can show: one returning a
	// material record and the albedo that goes with it.
	Material = "material"

	// Finish marks a surface finish that deliberately carries no albedo — the
	// material library's own presets, which exist to be paired with a colour
	// from somewhere else ("boltMat, boltColor = metal.brass();
	// bolt.material = material.brushedMetal; bolt.color = boltColor;"). The
	// browser has to draw one against a stand-in colour, where a @material
	// supplies its own.
	Finish = "finish"

	// Surface marks an operator preset: an overload of a function that takes a
	// surface first and returns the modified one, taking only what a document
	// supplies (the substrate, its colour, the body's frame) and baking the
	// operator's other arguments. The Bodies panel wraps a body's surface with
	// one (doc/material_composition.md, "Design: operators as presets").
	Surface = "surface"
)

// Validate checks every annotation in prog. A kind this package does not know
// is left alone, so a program written against a newer parser still assembles.
func Validate(prog *program.Program) error {
	if prog == nil || prog.Primary == nil {
		return nil
	}
	pkgs := append([]*program.Package{prog.Primary}, prog.Dependencies...)
	for _, pkg := range pkgs {
		if err := validatePackage(pkg); err != nil {
			return err
		}
	}
	return nil
}

func validatePackage(pkg *program.Package) error {
	if pkg == nil {
		return nil
	}
	for _, file := range pkg.Files {
		if file == nil {
			continue
		}
		for _, ann := range file.Annotations {
			if ann == nil {
				continue
			}
			var err error
			switch ann.Kind {
			case Material:
				err = validateMaterial(pkg, file, ann)
			case Finish:
				err = validateFinish(pkg, file, ann)
			case Surface:
				err = validateSurface(pkg, file, ann)
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// validateMaterial requires an @material to be a function. A record and an
// albedo are one surface, and a constant can hold only one of them, so a
// constant is not a material however it is named — the colour that belongs to
// it would have to live somewhere else, which is the design this one replaced.
func validateMaterial(pkg *program.Package, file *program.PackageFile, ann *program.Annotation) error {
	if declaresFunction(pkg, ann.Symbol) {
		return nil
	}
	if declaresConstant(pkg, ann.Symbol) {
		return fileErr(file, ann.Start, fmt.Sprintf(
			"@%s %q is a constant: a material is a record and its colour together, so an "+
				"@%s is a function returning both (\"%s() { return material.make(...), "+
				"[r, g, b]; }\"). A record with no albedo is a @%s",
			ann.Kind, ann.Symbol, Material, ann.Symbol, Finish))
	}
	return fileErr(file, ann.Start, fmt.Sprintf(
		"@%s marks %q, which package %q does not declare",
		ann.Kind, ann.Symbol, pkg.Path))
}

// validateFinish keeps the two kinds from drifting into synonyms. A finish is
// the colourless half of the pairing: the kinds are what tell the browser
// whether a swatch needs a stand-in colour or brings its own.
func validateFinish(pkg *program.Package, file *program.PackageFile, ann *program.Annotation) error {
	if declaresFunction(pkg, ann.Symbol) || declaresConstant(pkg, ann.Symbol) {
		return nil
	}
	return fileErr(file, ann.Start, fmt.Sprintf(
		"@%s marks %q, which package %q does not declare",
		ann.Kind, ann.Symbol, pkg.Path))
}

// validateSurface requires an @surface to be a function: an operator takes a
// surface and returns the modified one. The parser already refuses the
// annotation on a constant, so what is left to check is that the name is
// declared at all.
func validateSurface(pkg *program.Package, file *program.PackageFile, ann *program.Annotation) error {
	if declaresFunction(pkg, ann.Symbol) {
		return nil
	}
	return fileErr(file, ann.Start, fmt.Sprintf(
		"@%s marks %q, which package %q does not declare as a function",
		ann.Kind, ann.Symbol, pkg.Path))
}

// declaresConstant reports whether the package assigns name at the top level of
// any of its files.
func declaresConstant(pkg *program.Package, name string) bool {
	for _, file := range pkg.Files {
		if file == nil {
			continue
		}
		for _, assign := range file.Assignments {
			if assign == nil {
				continue
			}
			for _, variable := range assign.Variables {
				if variable != nil && variable.Name == name {
					return true
				}
			}
		}
	}
	return false
}

// declaresFunction reports whether the package declares a function named name
// at the top level of any of its files, at any arity.
func declaresFunction(pkg *program.Package, name string) bool {
	for _, file := range pkg.Files {
		if file == nil {
			continue
		}
		for _, fnType := range file.FnTypes {
			if fnType != nil && fnType.Name == name {
				return true
			}
		}
	}
	return false
}

func fileErr(file *program.PackageFile, loc *program.Location, msg string) error {
	path := ""
	if file != nil {
		path = file.Path
	}
	line := 0
	col := 0
	if loc != nil {
		line = int(loc.Line)
		col = int(loc.Column)
	}
	return &program.Error{
		FilePath: path,
		LineNum:  line,
		Column:   col,
		Message:  msg,
	}
}

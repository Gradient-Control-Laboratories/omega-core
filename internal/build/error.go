package build

import (
	"errors"
	"fmt"
	"strings"

	"gradientcontrol.com/omega/core/program"
)

// Convert an error to a build error with source code context.
func ErrWithPkgSrc(
	prog *program.Program,
	err error,
) error {
	if err == nil {
		return nil
	}

	var primarySrc string
	if prog.Primary != nil && len(prog.Primary.Files) > 0 {
		primarySrc = prog.Primary.Files[0].Source
	}

	src := primarySrc
	if errWithLoc, hasPkgPath := err.(*ErrorWithLoc); hasPkgPath {
		if errWithLoc.PkgPath != "" && (prog.Primary == nil || errWithLoc.PkgPath != prog.Primary.Path) {
			var pkg *program.Package
			for _, dep := range prog.Dependencies {
				if dep.Path == errWithLoc.PkgPath {
					pkg = dep
					break
				}
			}
			if pkg != nil {
				idx := errWithLoc.FileIdx
				if idx < 0 || idx >= len(pkg.Files) {
					idx = 0
				}
				if idx < len(pkg.Files) {
					file := pkg.Files[idx]
					if file.Source != "" {
						src = file.Source
					}
				}
			}
		} else if prog.Primary != nil {
			idx := errWithLoc.FileIdx
			if idx < 0 || idx >= len(prog.Primary.Files) {
				idx = 0
			}
			if idx < len(prog.Primary.Files) {
				file := prog.Primary.Files[idx]
				if file.Source != "" {
					src = file.Source
				}
			}
		}
	}
	return ErrWithSrc(src, err)
}

// Convert an error to a build error with source code context.
func ErrWithSrc(source string, err error) error {
	if err == nil {
		return nil
	}

	if buildErr, ok := err.(*program.Error); ok {
		if buildErr.Line == "" {
			buildErr.Line = LineFromSource(source, buildErr.LineNum)
		}
		return buildErr
	}

	lineNum := 1
	column := 1
	var line string

	errWithLoc, ok := err.(*ErrorWithLoc)
	if ok {
		lineNum = int(errWithLoc.Loc.Line)
		column = int(errWithLoc.Loc.Column)
		line = LineFromSource(source, lineNum)
	}

	wErr, ok := err.(interface {
		Unwrap() []error
	})
	if ok {
		wrappedErrs := wErr.Unwrap()
		newErrs := make([]error, len(wrappedErrs))
		for i, err := range wrappedErrs {
			newErrs[i] = ErrWithSrc(source, err)
		}
		return errors.Join(newErrs...)
	}

	return &program.Error{
		Line:    line,
		LineNum: lineNum,
		Column:  column,
		Length:  1,
		Message: err.Error(),
	}
}

func LineFromSource(source string, lineNum int) string {
	lines := strings.Split(source, "\n")
	if 1 <= lineNum && lineNum <= len(lines) {
		return lines[lineNum-1]
	}
	return ""
}

// ----------------------------------------------------------------------------

type ErrorWithLoc struct {
	PkgPath string
	FileIdx int
	Loc     *program.Location
	err     error
}

func ErrLoc(scope *Scope, loc *program.Location, err error) error {
	if loc == nil {
		return err
	}
	locErr, ok := err.(*ErrorWithLoc)
	if ok {
		return locErr
	}

	var fileIdx int
	for !scope.IsTopLevelScope() {
		if scope.IsFileScope() {
			fileIdx = scope.FileIdx()
		}
		scope = scope.Parent()
	}

	return &ErrorWithLoc{
		PkgPath: scope.Name(),
		FileIdx: fileIdx,
		Loc:     loc,
		err:     err,
	}
}

func PkgErrLoc(pkgPath string, buildErr *program.Error) error {
	return &ErrorWithLoc{
		PkgPath: pkgPath,
		Loc: &program.Location{
			Line:   int32(buildErr.LineNum),
			Column: int32(buildErr.Column),
		},
		err: errors.New(buildErr.Message),
	}
}

func (e *ErrorWithLoc) Error() string {
	prefix := fmt.Sprintf("%d:%d", e.Loc.Line, e.Loc.Column)
	if e.PkgPath != "" {
		prefix = fmt.Sprintf("package:%s:%s", e.PkgPath, prefix)
	}
	return fmt.Sprintf("%s: %s", prefix, e.err.Error())
}

func (e *ErrorWithLoc) Unwrap() error {
	return e.err
}

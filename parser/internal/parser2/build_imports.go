package parser2

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2"
	"gradientcontrol.com/omega/core/internal/scope"
	"gradientcontrol.com/omega/core/parser/internal/identifier2"
	"gradientcontrol.com/omega/core/program"
)

func buildImports(pkgFileScope scope.BuilderScope, imports []*Import) error {
	for _, imp := range imports {
		path, err := verifyImportPath(imp.Path)
		if err != nil {
			return &participle.ParseError{
				Pos: imp.Pos,
				Msg: err.Error(),
			}
		}

		if path == "builtin" && imp.Prefix == "." {
			continue
		}

		var prefix string
		if path == "omega" && imp.Prefix == "." {
			// Legacy omega dot imports still expose vector constants.
			path = "vector"
			prefix = "."
		} else if path == "omega" {
			path = "builtin"
			if imp.Prefix == "" {
				prefix = "omega"
			} else {
				prefix, err = verifyImportPrefix(imp.Prefix, path)
			}
		} else {
			prefix, err = verifyImportPrefix(imp.Prefix, path)
		}
		if err != nil {
			return &participle.ParseError{
				Pos: imp.Pos,
				Msg: err.Error(),
			}
		}

		start := posToLocation(imp.Pos)
		end := posToLocation(imp.EndPos)
		newImp := program.NewImport(prefix, path, start, end)

		err = pkgFileScope.AddImport(newImp)
		if err != nil {
			return &participle.ParseError{
				Pos: imp.Pos,
				Msg: err.Error(),
			}
		}
	}

	return nil
}

func verifyImportPath(path string) (string, error) {
	if path == "" {
		return "", errors.New("import path is empty")
	}

	unquoted, err := strconv.Unquote(path)
	if err != nil {
		return "", errors.New("missing double quotes around import path")
	}
	path = unquoted

	if strings.Contains(path, "//") {
		return "", errors.New("import path cannot contain '//'")
	}

	if strings.HasPrefix(path, "/") || strings.HasSuffix(path, "/") {
		return "", errors.New("import path cannot start or end with '/'")
	}

	for _, part := range strings.Split(path, "/") {
		if part == "." || part == ".." {
			return "", fmt.Errorf("import path contains invalid element %q", part)
		}
	}

	for _, c := range path {
		if c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' {
			continue
		}
		if c == '-' || c == '.' || c == '_' || c == '~' || c == '+' || c == '/' {
			continue
		}
		return "", fmt.Errorf("import path contains invalid character %q", c)
	}

	if strings.HasSuffix(path, ".gsl") {
		return "", errors.New("remove the '.gsl' suffix from the import path")
	}

	return path, nil
}

func verifyImportPrefix(prefix string, path string) (string, error) {
	if prefix != "" {
		if prefix != "." {
			err := identifier2.ValidatePackageName(prefix)
			if err != nil {
				return "", err
			}
		}
		return prefix, nil
	}

	// if no prefix is provided, use the last path segment as the prefix

	parts := strings.Split(path, "/")
	simpleName := parts[len(parts)-1]
	err := identifier2.ValidatePackageName(simpleName)
	return "", err
}

func buildMeshImports(pkgFileScope scope.BuilderScope, imports []*MeshImport) error {
	for _, imp := range imports {
		path, err := verifyAssetImportPath(imp.Path, "mesh", ".stl")
		if err != nil {
			return &participle.ParseError{Pos: imp.Pos, Msg: err.Error()}
		}

		start := posToLocation(imp.Pos)
		end := posToLocation(imp.EndPos)
		meshImport := program.NewMeshImport(imp.Index, path, start, end)
		err = pkgFileScope.AddMeshImport(meshImport)
		if err != nil {
			return &participle.ParseError{Pos: imp.Pos, Msg: err.Error()}
		}
	}
	return nil
}

func buildStepImports(pkgFileScope scope.BuilderScope, imports []*StepImport) error {
	for _, imp := range imports {
		path, err := verifyAssetImportPath(imp.Path, "step", ".step", ".stp", ".hodge")
		if err != nil {
			return &participle.ParseError{Pos: imp.Pos, Msg: err.Error()}
		}

		start := posToLocation(imp.Pos)
		end := posToLocation(imp.EndPos)
		stepImport := program.NewStepImport(imp.Index, path, start, end)
		err = pkgFileScope.AddStepImport(stepImport)
		if err != nil {
			return &participle.ParseError{Pos: imp.Pos, Msg: err.Error()}
		}
	}
	return nil
}

func buildWarpImports(pkgFileScope scope.BuilderScope, imports []*WarpImport) error {
	for _, imp := range imports {
		path, err := verifyAssetImportPath(imp.Path, "warp", ".castsim")
		if err != nil {
			return &participle.ParseError{Pos: imp.Pos, Msg: err.Error()}
		}

		start := posToLocation(imp.Pos)
		end := posToLocation(imp.EndPos)
		warpImport := program.NewWarpImport(imp.Index, path, start, end)
		err = pkgFileScope.AddWarpImport(warpImport)
		if err != nil {
			return &participle.ParseError{Pos: imp.Pos, Msg: err.Error()}
		}
	}
	return nil
}

func verifyAssetImportPath(path, kind string, suffixes ...string) (string, error) {
	unquoted, err := strconv.Unquote(path)
	if err != nil {
		return "", fmt.Errorf("missing double quotes around %s import path", kind)
	}
	path = unquoted
	if path == "" {
		return "", fmt.Errorf("%s import path is empty", kind)
	}
	if strings.ContainsAny(path, `/\`) || path == "." || path == ".." {
		return "", fmt.Errorf("%s import path must be a filename beside the GSL file", kind)
	}
	if strings.HasPrefix(path, ".") || !hasAssetSuffix(path, suffixes) {
		return "", fmt.Errorf("%s import path must name a %s file", kind, strings.Join(suffixes, " or "))
	}
	// Allow any printable character (uploaded CAD files routinely have spaces,
	// parentheses, etc.); only path separators (above) and control characters
	// are rejected, matching the document store's asset-filename rule.
	for _, c := range path {
		if c < 0x20 || c == 0x7f {
			return "", fmt.Errorf("%s import path contains invalid character %q", kind, c)
		}
	}
	return path, nil
}

func hasAssetSuffix(path string, suffixes []string) bool {
	path = strings.ToLower(path)
	for _, suffix := range suffixes {
		if strings.HasSuffix(path, suffix) {
			return true
		}
	}
	return false
}

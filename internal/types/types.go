package types

import (
	"fmt"

	"gradientcontrol.com/omega/core/program"
)

type PackageAnalyzer struct {
	prog        *program.Program
	pkgTypes    map[string]*TypeEnv
	currentPkg  *program.Package
	env         *TypeEnv
	visiting    map[string]bool
	fileImports map[*program.PackageFile]*FileImports
	pkgVarDefs  map[string]pkgVarDef
	fnFiles     map[*program.Function]*program.PackageFile
	isPkgFn     map[*program.Function]bool
	filePkgs    map[*program.PackageFile]*program.Package
	inUserFn    int
}

// Analyze performs type analysis on the GSL program and populates type information on AST nodes.
func Analyze(prog *program.Program) error {
	if prog == nil {
		return nil
	}

	pkgTypes := make(map[string]*TypeEnv)
	fnFiles := make(map[*program.Function]*program.PackageFile)
	fileImports := make(map[*program.PackageFile]*FileImports)
	filePkgs := make(map[*program.PackageFile]*program.Package)

	// Analyze dependencies first, then primary package
	for _, dep := range prog.Dependencies {
		env, err := analyzePackage(dep, dep.Path, prog, pkgTypes, fnFiles, fileImports, filePkgs)
		if err != nil {
			return err
		}
		pkgTypes[dep.Path] = env
	}

	if prog.Primary != nil {
		_, err := analyzePackage(prog.Primary, prog.Primary.Path, prog, pkgTypes, fnFiles, fileImports, filePkgs)
		if err != nil {
			return err
		}
	}

	return nil
}

func analyzePackage(pkg *program.Package, path string, prog *program.Program, pkgTypes map[string]*TypeEnv, fnFiles map[*program.Function]*program.PackageFile, fileImports map[*program.PackageFile]*FileImports, filePkgs map[*program.PackageFile]*program.Package) (*TypeEnv, error) {
	if pkg == nil {
		return nil, nil
	}
	pkg.Path = path
	program.UpdateImports(pkg, pkg.Path)
	for _, f := range pkg.Files {
		filePkgs[f] = pkg
	}

	pa := &PackageAnalyzer{
		prog:        prog,
		pkgTypes:    pkgTypes,
		currentPkg:  pkg,
		env:         newEnv(nil),
		visiting:    make(map[string]bool),
		fileImports: fileImports,
		pkgVarDefs:  make(map[string]pkgVarDef),
		fnFiles:     fnFiles,
		isPkgFn:     make(map[*program.Function]bool),
		filePkgs:    filePkgs,
	}
	pkgTypes[pkg.Path] = pa.env

	// 1. Resolve imports for each file in the package
	for _, file := range pkg.Files {
		fi := &FileImports{
			named: make(map[string]*TypeEnv),
		}
		for _, imp := range file.Imports {
			importedEnv, exists := pkgTypes[imp.Path]
			if !exists {
				importedPkg, err := prog.GetImport(imp.Path)
				if err != nil {
					return nil, fmt.Errorf("failed to get import %q: %v", imp.Path, err)
				}
				if importedPkg != nil {
					importedEnv, err = analyzePackage(importedPkg, imp.Path, prog, pkgTypes, pa.fnFiles, pa.fileImports, pa.filePkgs)
					if err != nil {
						return nil, fmt.Errorf("failed to analyze package %q: %v", imp.Path, err)
					}
					pkgTypes[imp.Path] = importedEnv
				}
			}
			if importedEnv != nil {
				prefix := imp.PrefixOrSimpleName()
				if prefix == "." {
					fi.dots = append(fi.dots, importedEnv)
					simpleName := imp.SimpleName()
					if simpleName != "" && fi.named[simpleName] == nil {
						fi.named[simpleName] = importedEnv
					}
				} else {
					fi.named[prefix] = importedEnv
				}
			} else {
				return nil, fmt.Errorf("unknown package prefix: %s", imp.PrefixOrSimpleName())
			}
		}
		pa.fileImports[file] = fi
	}

	// 2. Map all package-level variables to their assignments
	for _, file := range pkg.Files {
		for _, assign := range file.Assignments {
			for i, v := range assign.Variables {
				pa.pkgVarDefs[v.Name] = pkgVarDef{
					assignment: assign,
					file:       file,
					index:      i,
				}
			}
		}
	}

	// 3. Register all functions in the package environment and map them to their declaring file
	for _, file := range pkg.Files {
		for _, fnType := range file.FnTypes {
			pa.env.fns[fnType.Name] = fnType
			for _, group := range fnType.Groups {
				for _, fn := range group.Functions {
					pa.isPkgFn[fn] = true
					pa.registerFn(fn, file)
				}
			}
		}
	}

	// 4. Resolve types of all package-level variables
	for _, file := range pkg.Files {
		for _, assign := range file.Assignments {
			for _, v := range assign.Variables {
				_, err := pa.resolvePackageVar(v.Name)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	// 4a. Infer the returns of ANNOTATED ZERO-ARGUMENT functions. Every other
	// function body is inferred from a call site, because its parameter types
	// come from the arguments — so an uncalled function is never typed at all.
	// A zero-parameter function is the exception: there are no argument types
	// to supply, so its body can be inferred standalone. That is what lets the
	// annotation package hold a @material pair to the material and colour it
	// promises even in a document that never draws with it, which is the whole
	// point of an annotation surviving deadcode.
	for _, file := range pkg.Files {
		for _, ann := range file.Annotations {
			if ann == nil {
				continue
			}
			fnType := pa.env.fns[ann.Symbol]
			if fnType == nil {
				continue
			}
			for _, group := range fnType.Groups {
				if group.ParamCount != 0 {
					continue
				}
				for _, fn := range group.Functions {
					retTypes, err := pa.inferUserCall(fn, nil, pa.env, file)
					if err != nil {
						return nil, fmt.Errorf("@%s %s: %w", ann.Kind, ann.Symbol, err)
					}
					// Stamp the returns. inferOperandType deliberately does not
					// cache inside a function body, because the same body takes
					// its operand types from each call site's arguments. With no
					// parameters there is exactly one instantiation, so the types
					// are unambiguous and a later pass can read them.
					if len(retTypes) == len(fn.Returns) {
						for i, retOp := range fn.Returns {
							if retOp != nil {
								retOp.Type = retTypes[i]
							}
						}
					}
				}
			}
		}
	}

	// 4b. Type the property expressions (name.color = …, name.material = …).
	// They are expressions like any other, but hang off a variable rather than
	// forming an assignment, so the steps above never reached them. Left
	// untyped, a list inside one had no shape: the constant fold stamped it as
	// flat values, and a backend rebuilding a constant from its type rebuilt a
	// list of rows as one long vector, so indexing it gave a scalar. That is
	// how material.withRoughness failed on the fidget path while the same call
	// in an ordinary assignment worked.
	for _, file := range pkg.Files {
		for _, assign := range file.Assignments {
			for _, v := range assign.Variables {
				for _, prop := range v.Properties {
					if _, err := inferOperandType(prop.Expression, pa.env, file, pa); err != nil {
						return nil, err
					}
				}
			}
		}
	}

	// 5. Type check package returns
	for _, file := range pkg.Files {
		for _, ret := range file.Returns {
			_, err := inferOperandType(ret, pa.env, file, pa)
			if err != nil {
				return nil, err
			}
		}
		for _, ret := range file.ColorReturns {
			_, err := inferOperandType(ret, pa.env, file, pa)
			if err != nil {
				return nil, err
			}
		}
		for _, ret := range file.MaterialReturns {
			_, err := inferOperandType(ret, pa.env, file, pa)
			if err != nil {
				return nil, err
			}
		}
	}

	return pa.env, nil
}

func (pa *PackageAnalyzer) registerFn(fn *program.Function, file *program.PackageFile) {
	if oldFile, exists := pa.fnFiles[fn]; exists {
		oldPkg := pa.filePkgs[oldFile]
		newPkg := pa.filePkgs[file]
		if oldPkg != nil && oldPkg.Path == "" && newPkg != nil && newPkg.Path != "" {
			pa.fnFiles[fn] = file
		} else {
			return
		}
	} else {
		pa.fnFiles[fn] = file
	}
	for _, localFnType := range fn.FnTypes {
		for _, group := range localFnType.Groups {
			for _, localFn := range group.Functions {
				pa.registerFn(localFn, file)
			}
		}
	}
}

func (pa *PackageAnalyzer) resolvePackageVar(name string) (*program.Type, error) {
	if t := pa.env.vars[name]; t != nil {
		return t, nil
	}
	def, exists := pa.pkgVarDefs[name]
	if !exists {
		return nil, nil // not a package var
	}
	if pa.visiting[name] {
		return nil, fmt.Errorf("circular reference detected for variable: %s", name)
	}
	pa.visiting[name] = true
	defer func() { pa.visiting[name] = false }()

	rhsTypes, err := inferOperandType(def.assignment.Expression, pa.env, def.file, pa)
	if err != nil {
		return nil, err
	}
	if len(rhsTypes) != len(def.assignment.Variables) {
		return nil, fmt.Errorf("assignment count mismatch: expected %d values, got %d; multi-value functions (like *Normal) must assign both values", len(def.assignment.Variables), len(rhsTypes))
	}
	for i, v := range def.assignment.Variables {
		v.Type = rhsTypes[i]
		pa.env.vars[v.Name] = rhsTypes[i]
	}
	return pa.env.vars[name], nil
}

func commonType(a, b *program.Type) (*program.Type, error) {
	if a == nil || b == nil {
		return nil, fmt.Errorf("nil type")
	}
	if !a.IsNumeric() || !b.IsNumeric() {
		return nil, fmt.Errorf("incompatible types: %s and %s", a.TypeString(), b.TypeString())
	}
	if a.Equal(b) {
		return a, nil
	}
	if a.Kind == program.TypeKind_SCALAR && (b.Kind == program.TypeKind_LIST || b.Kind == program.TypeKind_MATRIX) {
		return b, nil
	}
	if (a.Kind == program.TypeKind_LIST || a.Kind == program.TypeKind_MATRIX) && b.Kind == program.TypeKind_SCALAR {
		return a, nil
	}
	if a.Kind == b.Kind && a.ElemType != nil && b.ElemType != nil && a.Size == b.Size {
		commonElem, err := commonType(a.ElemType, b.ElemType)
		if err == nil {
			return &program.Type{Kind: a.Kind, Size: a.Size, ElemType: commonElem}, nil
		}
	}

	return nil, fmt.Errorf("incompatible types: %s and %s", a.TypeString(), b.TypeString())
}

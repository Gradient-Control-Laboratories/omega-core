package closure

import (
	"fmt"

	"gradientcontrol.com/omega/core/program"
)

// Transform rewrites higher-order functions, closures, and partial applications
// into specialized first-order functions and direct calls across the program AST.
func Transform(prog *program.Program) error {
	if prog == nil {
		return nil
	}

	t := newTransformer(prog)
	return t.run()
}

type transformer struct {
	prog          *program.Program
	nameCounters  map[string]int
	pkgFunctions  map[string]map[string]*program.FnType // pkgPath -> name -> FnType
	fileFunctions map[*program.PackageFile]map[string]*program.FnType
	pkgByPath     map[string]*program.Package
	filePkgs      map[*program.PackageFile]*program.Package
	processedFns  map[*program.Function]bool
	// liftedFns memoizes liftFunction: a lifted function's shape depends only
	// on its target and the ordered names of its captures, never on their
	// values, so every instantiation of a factory shares one.
	liftedFns map[liftKey]string
	// listBindings records the list of functions a call site supplied for a
	// parameter of a specialized clone, so a fold over that parameter inside
	// the clone can be unrolled. It is kept beside the AST rather than in it:
	// the elements name factory templates that are about to be pruned, so an
	// assignment holding them would not survive to type analysis.
	listBindings map[*program.Function]map[string][]*program.Operand
	// listParamCache memoizes paramIsFunctionList, which is asked about every
	// argument of every call and otherwise walks the callee's whole body.
	listParamCache map[listParamKey]bool
	// fnStack is the chain of functions the descent is currently inside,
	// outermost first. A nested function reads the names its enclosing
	// functions bind, so deciding whether a bare reference is a value or a
	// function takes the whole chain and not just the innermost link — see
	// bindsValueName.
	fnStack []*program.Function
}

// listParamKey identifies one parameter of one function.
type listParamKey struct {
	fn    *program.Function
	param string
}

// liftKey identifies a lifted function: the closure target it clones and the
// capture names that lead its parameter list, in order.
type liftKey struct {
	target   *program.Function
	captures string
}

func newTransformer(prog *program.Program) *transformer {
	return &transformer{
		prog:           prog,
		nameCounters:   make(map[string]int),
		pkgFunctions:   make(map[string]map[string]*program.FnType),
		fileFunctions:  make(map[*program.PackageFile]map[string]*program.FnType),
		pkgByPath:      make(map[string]*program.Package),
		filePkgs:       make(map[*program.PackageFile]*program.Package),
		processedFns:   make(map[*program.Function]bool),
		liftedFns:      make(map[liftKey]string),
		listBindings:   make(map[*program.Function]map[string][]*program.Operand),
		listParamCache: make(map[listParamKey]bool),
	}
}

type closureVal struct {
	targetFnName   string
	targetFnPrefix string
	targetFile     *program.PackageFile
	targetFn       *program.Function
	enclosingName  string
	captures       []captureBinding
	isNestedReturn bool
	// lifted marks a closure whose target lives in another package than the
	// file that created it. Its function was synthesized in the target's own
	// file with the captures turned into leading parameters, so the body never
	// leaves the package whose names it uses; a call through the closure
	// supplies the capture values as leading arguments instead. See
	// liftFunction.
	lifted bool
	// callerPrefix is the prefix the creating file used to reach the factory,
	// which is the prefix a lifted call has to carry to reach the function
	// synthesized in the factory's package.
	callerPrefix string
}

type captureBinding struct {
	name  string
	value *program.Operand
}

func (t *transformer) run() error {
	t.indexFunctions()

	// 1. Validate returns across all packages and functions
	if err := t.validateReturns(); err != nil {
		return err
	}

	// 2. Transform starting from primary package top-level expressions and functions
	if t.prog.Primary != nil {
		for _, file := range t.prog.Primary.Files {
			if err := t.transformFile(file); err != nil {
				return err
			}
		}
	}

	// 2b. A library function nothing reached, whose body calls a higher-order
	// template — fastener.shcsHole folding drill.hole, say — would otherwise
	// keep that call while step 3 deletes what it calls, and the next link
	// fails on every document that imports the library. Transform just those;
	// every other unreached function is left exactly as it was.
	for _, dep := range t.prog.Dependencies {
		for _, file := range dep.Files {
			if err := t.transformUnreachedTemplateCallers(file); err != nil {
				return err
			}
		}
	}

	// 3. Prune unspecialized higher-order templates and closure factories
	var allFiles []*program.PackageFile
	if t.prog.Primary != nil {
		allFiles = append(allFiles, t.prog.Primary.Files...)
	}
	for _, dep := range t.prog.Dependencies {
		allFiles = append(allFiles, dep.Files...)
	}
	t.pruneHigherOrderTemplates(allFiles)

	return nil
}

func (t *transformer) indexFunctions() {
	if t.prog.Primary != nil {
		pkgMap := make(map[string]*program.FnType)
		for _, file := range t.prog.Primary.Files {
			t.filePkgs[file] = t.prog.Primary
			fileMap := make(map[string]*program.FnType)
			for _, ft := range file.FnTypes {
				pkgMap[ft.Name] = ft
				fileMap[ft.Name] = ft
			}
			t.fileFunctions[file] = fileMap
		}
		t.pkgFunctions[""] = pkgMap
		t.pkgFunctions[t.prog.Primary.Path] = pkgMap
		t.pkgByPath[""] = t.prog.Primary
		if t.prog.Primary.Path != "" {
			t.pkgByPath[t.prog.Primary.Path] = t.prog.Primary
		}
	}

	for _, dep := range t.prog.Dependencies {
		pkgMap := make(map[string]*program.FnType)
		for _, file := range dep.Files {
			t.filePkgs[file] = dep
			fileMap := make(map[string]*program.FnType)
			for _, ft := range file.FnTypes {
				pkgMap[ft.Name] = ft
				fileMap[ft.Name] = ft
			}
			t.fileFunctions[file] = fileMap
		}
		t.pkgFunctions[dep.Path] = pkgMap
		t.pkgByPath[dep.Path] = dep
	}
}

func (t *transformer) nextClosureName(enclosingName string) string {
	base := fmt.Sprintf("gcl_closure_%s", enclosingName)
	count := t.nameCounters[base]
	t.nameCounters[base] = count + 1
	if count == 0 {
		return base
	}
	return fmt.Sprintf("%s_%d", base, count+1)
}

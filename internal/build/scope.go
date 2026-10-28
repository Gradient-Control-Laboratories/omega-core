package build

import (
	"fmt"

	"gradientcontrol.com/omega/core/internal/scope"
	"gradientcontrol.com/omega/core/program"
)

type Scope struct {
	builder *Builder
	uScope  scope.Scope

	// the package path or function name associated with this scope
	// or empty for the primary scope and "root" for the root scope
	name string

	// True if this scope is a file scope.
	// Only file scopes have imports, and their imports are not promoted to the parent scope.
	// File scopes do not contain localVar, paramVars, or fnTypes, which are all defined at the package scope.
	isFileScope bool
	fileIdx     int

	parent       *Scope
	dotImports   []*Scope
	namedImports map[string]*Scope

	// The source location of the node that created this scope
	start *program.Location

	// a map of local variable name to the variable's info in this scope
	localVars map[string]VarInfo

	// a map of param name to the parameter's info in this scope
	paramVars map[string]VarInfo

	// a map of function name to the function's declaration in this scope
	fnDecls map[string]*program.FnType

	// a map of function name to the file scope where the function was declared
	fnDeclScopes map[string]*Scope

	// a map of function name to the function's variable info in this scope
	fnVars map[string]VarInfo

	// a map of function to compiled function specializations (FnInfo)
	fnInfos map[*program.Function][]FnInfo

	// inlined is true for the ephemeral scope created when a function call is
	// inlined at its call site. Locals declared while inlining live in this
	// scope, but their generated WGSL (e.g. a `let`) belongs to whatever
	// function body the inlining is happening inside — see IsInlinedScope.
	inlined bool

	agentScope AgentScope
}

// Returns the function declaration with the given name if it exists, otherwise returns nil.
func (s *Scope) ExistingFnDecl(name string) *program.FnType {
	curr := s
	for curr != nil {
		if fnDecl, ok := curr.fnDecls[name]; ok {
			return fnDecl
		}
		curr = curr.parent
	}
	curr = s
	for curr != nil {
		var builtinDecl *program.FnType
		for _, dotScope := range curr.dotImports {
			if dotScope.name == "builtin" {
				if fnDecl, ok := dotScope.fnDecls[name]; ok {
					builtinDecl = fnDecl
				}
				continue
			}
			if fnDecl, ok := dotScope.fnDecls[name]; ok {
				return fnDecl
			}
		}
		if builtinDecl != nil {
			return builtinDecl
		}
		curr = curr.parent
	}
	return nil
}

// Returns the scope containing the function declaration with the given name if it exists, otherwise returns nil.
func (s *Scope) FindFnScope(name string) *Scope {
	curr := s
	for curr != nil {
		if fileScope, ok := curr.fnDeclScopes[name]; ok && fileScope != nil {
			return fileScope
		}
		if _, ok := curr.fnDecls[name]; ok {
			return curr
		}
		curr = curr.parent
	}
	curr = s
	for curr != nil {
		var builtinScope *Scope
		for _, dotScope := range curr.dotImports {
			if dotScope.name == "builtin" {
				if fileScope, ok := dotScope.fnDeclScopes[name]; ok && fileScope != nil {
					builtinScope = fileScope
				} else if _, ok := dotScope.fnDecls[name]; ok {
					builtinScope = dotScope
				}
				continue
			}
			if fileScope, ok := dotScope.fnDeclScopes[name]; ok && fileScope != nil {
				return fileScope
			}
			if _, ok := dotScope.fnDecls[name]; ok {
				return dotScope
			}
		}
		if builtinScope != nil {
			return builtinScope
		}
		curr = curr.parent
	}
	return nil
}

// FindFunction finds the function declaration, matching function overload, and enclosing scope.
func (s *Scope) FindFunction(name string, paramCount int) (*program.FnType, *program.Function, *Scope) {
	decl := s.ExistingFnDecl(name)
	if decl == nil {
		return nil, nil, nil
	}
	enclosingScope := s.FindFnScope(name)
	for _, group := range decl.Groups {
		if int(group.ParamCount) == paramCount {
			if len(group.Functions) > 0 {
				return decl, group.Functions[0], enclosingScope
			}
		}
	}
	return decl, nil, enclosingScope
}

// RegisterFnDecl registers a function declaration in this scope (or its parent if file scope).
func (s *Scope) RegisterFnDecl(
	name string, decl *program.FnType,
	newFnVar func(string, *program.FnType, *Scope) VarInfo,
	fileScope *Scope,
) (*program.FnType, error) {
	if s.parent != nil && s.IsFileScope() {
		return s.parent.RegisterFnDecl(name, decl, newFnVar, fileScope)
	}

	fnDecl := s.fnDecls[name]
	if fnDecl == nil {
		// Invariant: Upstream parser (verifyUndefinedLocally) and linker (checkLocalDeclarationCollisions)
		// guarantee that local and dot-imported declarations do not collide with existing symbols.
		if s.fnDecls == nil {
			s.fnDecls = make(map[string]*program.FnType)
		}
		s.fnDecls[name] = decl
		if s.fnDeclScopes == nil {
			s.fnDeclScopes = make(map[string]*Scope)
		}
		s.fnDeclScopes[name] = fileScope
		if s.fnVars == nil {
			s.fnVars = make(map[string]VarInfo)
		}
		if newFnVar != nil {
			s.fnVars[name] = newFnVar(name, decl, fileScope)
		}
		fnDecl = decl
	} else if fnDecl != decl {
		for _, group := range decl.Groups {
			var existingGroup *program.FnGroup
			for _, eg := range fnDecl.Groups {
				if eg.ParamCount == group.ParamCount {
					existingGroup = eg
					break
				}
			}
			if existingGroup != nil {
				existingGroup.Functions = append(existingGroup.Functions, group.Functions...)
			} else {
				fnDecl.Groups = append(fnDecl.Groups, group)
			}
		}
	}
	return fnDecl, nil
}

// InfoForArgs returns cached FnInfo for the given function and argument types, or nil.
func (s *Scope) InfoForArgs(fn *program.Function, args []Expression) FnInfo {
	curr := s
	for curr != nil {
		if infos, ok := curr.fnInfos[fn]; ok {
			for _, info := range infos {
				if MatchArgTypes(args, info.ParamTypes()) {
					return info
				}
			}
		}
		curr = curr.parent
	}
	return nil
}

// AddFnInfo adds compiled function info to this scope's cache.
func (s *Scope) AddFnInfo(fn *program.Function, fnInfo FnInfo) error {
	if s.fnInfos == nil {
		s.fnInfos = make(map[*program.Function][]FnInfo)
	}
	s.fnInfos[fn] = append(s.fnInfos[fn], fnInfo)
	return nil
}

func NewScope(b *Builder, name string, start *program.Location, parent *Scope, uScope scope.Scope) *Scope {
	return &Scope{builder: b, name: name, start: start, parent: parent, uScope: uScope}
}

// Return the package path or function name associated with this scope
// or empty for the primary scope and "root" for the root scope
func (s *Scope) Name() string {
	return s.name
}

func (s *Scope) Parent() *Scope {
	return s.parent
}

func (s *Scope) UScope() scope.Scope {
	return s.uScope
}

// Return true if this scope is the primary package
func (s *Scope) IsPrimaryScope() bool {
	return s.name == ""
}

// Return true if this scope or any of its ancestors is the primary package
func (s *Scope) InPrimaryScope() bool {
	curr := s
	for curr != nil {
		if curr.IsPrimaryScope() {
			return true
		}
		curr = curr.Parent()
	}
	return false
}

// Return true if this scope is a file scope
func (s *Scope) IsFileScope() bool {
	return s.isFileScope
}

func (s *Scope) FileIdx() int {
	return s.fileIdx
}

// Return true if this scope is either the primary package or an imported package
func (s *Scope) IsTopLevelScope() bool {
	return s.parent == s.builder.rootScope
}

// Return true if this scope is an imported package
func (s *Scope) IsImportScope() bool {
	return s.IsTopLevelScope() && !s.IsPrimaryScope()
}

// IsInlinedScope reports whether this is the ephemeral scope created to hold an
// inlined call's locals. Such locals have no function body of their own; their
// generated definitions must be emitted into the enclosing function body that
// references them (the current scope at codegen time), not gated to a scope
// that no longer exists by then.
func (s *Scope) IsInlinedScope() bool {
	return s.inlined
}

func (s *Scope) String() string {
	return fmt.Sprintf("Scope(%s)", s.name)
}

type AgentScope interface {
	IsAgentScope()
}

func (s *Scope) GetAgentScope() AgentScope {
	if s.uScope != nil && (s.IsTopLevelScope() || s.IsFileScope()) {
		return s.uScope.GetAgentScope()
	}
	return s.agentScope
}

func (s *Scope) SetAgentScope(agentScope AgentScope) {
	if s.uScope != nil && (s.IsTopLevelScope() || s.IsFileScope()) {
		s.uScope.SetAgentScope(agentScope)
	} else {
		s.agentScope = agentScope
	}
}

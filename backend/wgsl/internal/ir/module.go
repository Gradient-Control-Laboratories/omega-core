package ir

import (
	"bytes"
	"io"
	"strings"
)

// Module represents the complete top-level WGSL shader document.
type Module struct {
	Uniforms        []*UniformDecl
	Predefined      []string
	Imports         []*FuncDecl
	Functions       []*FuncDecl
	RuntimeSnippets []string
	Structs         []*StructDecl

	predefinedSeen map[string]bool
	snippetsSeen   map[string]bool
}

// NewModule constructs an empty Module container.
func NewModule() *Module {
	return &Module{
		predefinedSeen: make(map[string]bool),
		snippetsSeen:   make(map[string]bool),
	}
}

// AddUniform adds a bind-group uniform declaration.
func (m *Module) AddUniform(u *UniformDecl) {
	m.Uniforms = append(m.Uniforms, u)
}

// AddPredefined adds a static predefined function, avoiding duplicate definitions.
func (m *Module) AddPredefined(src string) {
	trimmed := strings.TrimSpace(src)
	if trimmed == "" || m.predefinedSeen[trimmed] {
		return
	}
	m.predefinedSeen[trimmed] = true
	m.Predefined = append(m.Predefined, trimmed)
}

// AddImport adds an imported function declaration.
func (m *Module) AddImport(fn *FuncDecl) {
	m.Imports = append(m.Imports, fn)
}

// AddFunction adds a user or synthesized function declaration in preserved registration order.
func (m *Module) AddFunction(fn *FuncDecl) {
	m.Functions = append(m.Functions, fn)
}

// AddRuntimeSnippet adds a static runtime snippet or constant template, avoiding duplicates.
func (m *Module) AddRuntimeSnippet(src string) {
	trimmed := strings.TrimSpace(src)
	if trimmed == "" || m.snippetsSeen[trimmed] {
		return
	}
	m.snippetsSeen[trimmed] = true
	m.RuntimeSnippets = append(m.RuntimeSnippets, trimmed)
}

// AddStruct adds a struct declaration.
func (m *Module) AddStruct(s *StructDecl) {
	m.Structs = append(m.Structs, s)
}

// Format renders the complete shader in canonical declaration order:
// Uniforms -> Predefined -> Imports -> Functions -> RuntimeSnippets -> Structs.
func (m *Module) Format(w io.Writer) error {
	first := true
	emitSep := func() error {
		if !first {
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
		}
		first = false
		return nil
	}

	for _, u := range m.Uniforms {
		if err := emitSep(); err != nil {
			return err
		}
		if err := u.Format(w); err != nil {
			return err
		}
	}

	for _, p := range m.Predefined {
		if err := emitSep(); err != nil {
			return err
		}
		if _, err := io.WriteString(w, p+"\n"); err != nil {
			return err
		}
	}

	for _, fn := range m.Imports {
		if err := emitSep(); err != nil {
			return err
		}
		if err := fn.Format(w); err != nil {
			return err
		}
	}

	for _, fn := range m.Functions {
		if err := emitSep(); err != nil {
			return err
		}
		if err := fn.Format(w); err != nil {
			return err
		}
	}

	for _, s := range m.RuntimeSnippets {
		if err := emitSep(); err != nil {
			return err
		}
		if _, err := io.WriteString(w, s+"\n"); err != nil {
			return err
		}
	}

	for _, st := range m.Structs {
		if err := emitSep(); err != nil {
			return err
		}
		if err := st.Format(w); err != nil {
			return err
		}
	}

	return nil
}

// WGSL returns the formatted shader string.
func (m *Module) WGSL() string {
	var buf bytes.Buffer
	_ = m.Format(&buf)
	return strings.TrimSpace(buf.String())
}

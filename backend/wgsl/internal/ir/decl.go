package ir

import (
	"fmt"
	"io"
)

// StructField represents a named, typed field within a WGSL struct declaration.
type StructField struct {
	// Read-only immutable fields set at construction
	Name string
	Type string
}

// StructDecl represents a top-level WGSL struct declaration.
type StructDecl struct {
	// Read-only immutable fields set at construction
	Name   string
	Fields []StructField
}

// NewStructDecl constructs a new struct declaration.
func NewStructDecl(name string, fields []StructField) *StructDecl {
	f := make([]StructField, len(fields))
	copy(f, fields)
	return &StructDecl{
		Name:   name,
		Fields: f,
	}
}

// Format renders the struct declaration.
func (s *StructDecl) Format(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "struct %s {\n", s.Name); err != nil {
		return err
	}
	for _, f := range s.Fields {
		if _, err := fmt.Fprintf(w, "  %s : %s,\n", f.Name, f.Type); err != nil {
			return err
		}
	}
	_, err := fmt.Fprint(w, "}")
	return err
}

// UniformDecl represents a top-level bind-group uniform declaration.
type UniformDecl struct {
	// Read-only immutable fields set at construction
	Group   int
	Binding int
	Name    string
	Type    string
}

// NewUniformDecl constructs a new uniform declaration.
func NewUniformDecl(group, binding int, name, typeName string) *UniformDecl {
	return &UniformDecl{
		Group:   group,
		Binding: binding,
		Name:    name,
		Type:    typeName,
	}
}

// Format renders the uniform declaration.
func (u *UniformDecl) Format(w io.Writer) error {
	_, err := fmt.Fprintf(w, "@group(%d) @binding(%d) var<uniform> %s: %s;\n", u.Group, u.Binding, u.Name, u.Type)
	return err
}

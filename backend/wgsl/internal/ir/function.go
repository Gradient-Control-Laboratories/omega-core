package ir

import (
	"fmt"
	"io"
	"strings"
)

// Param represents a named, typed parameter in a function declaration.
type Param struct {
	// Read-only immutable fields set at construction
	Name string
	Type string
}

// FuncDecl represents a top-level WGSL function declaration.
type FuncDecl struct {
	// Read-only immutable fields set at construction
	Name           string
	Params         []*Param
	RawParams      string
	ReturnType     string
	Body           *Block
	TrailingSource string
	RawSource      string
}

// NewFuncDecl constructs a new function declaration.
func NewFuncDecl(name string, params []*Param, returnType string, body *Block) *FuncDecl {
	p := make([]*Param, len(params))
	copy(p, params)
	if body == nil {
		body = NewBlock()
	}
	return &FuncDecl{
		Name:       name,
		Params:     p,
		ReturnType: returnType,
		Body:       body,
	}
}

// NewRawFuncDecl constructs a function declaration from pre-formatted WGSL source.
func NewRawFuncDecl(rawSource string) *FuncDecl {
	return &FuncDecl{
		RawSource: strings.TrimSpace(rawSource),
	}
}

// Format renders the function signature and indented body.
func (f *FuncDecl) Format(w io.Writer) error {
	if f.RawSource != "" {
		if _, err := io.WriteString(w, f.RawSource); err != nil {
			return err
		}
		if f.TrailingSource != "" {
			if _, err := fmt.Fprintf(w, "\n\n%s", f.TrailingSource); err != nil {
				return err
			}
		}
		return nil
	}

	paramsJoined := f.RawParams
	if paramsJoined == "" {
		var paramStrs []string
		for _, p := range f.Params {
			paramStrs = append(paramStrs, fmt.Sprintf("%s: %s", p.Name, p.Type))
		}
		paramsJoined = strings.Join(paramStrs, ", ")
	}

	if f.ReturnType != "" && f.ReturnType != "void" {
		if _, err := fmt.Fprintf(w, "fn %s(%s) -> %s {\n", f.Name, paramsJoined, f.ReturnType); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(w, "fn %s(%s) {\n", f.Name, paramsJoined); err != nil {
			return err
		}
	}

	if err := f.Body.Format(w, "  "); err != nil {
		return err
	}
	if _, err := fmt.Fprint(w, "}\n"); err != nil {
		return err
	}
	if f.TrailingSource != "" {
		if _, err := fmt.Fprintf(w, "\n%s\n", f.TrailingSource); err != nil {
			return err
		}
	}
	return nil
}

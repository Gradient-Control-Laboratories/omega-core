package propval

import (
	"fmt"
	"strings"

	"gradientcontrol.com/omega/core/program"
)

// CheckCycles validates that property assignments do not introduce direct or
// transitive circular dependencies.
func CheckCycles(prog *program.Program) error {
	if prog == nil || prog.Primary == nil {
		return nil
	}

	for _, pkg := range allPackages(prog) {
		if err := checkPackageCycles(pkg); err != nil {
			return err
		}
	}
	return nil
}

type depGraph struct {
	// Read-only immutable graph structures
	nodes map[string]bool
	edges map[string][]string
	locs  map[string]*program.Location
	file  *program.PackageFile
}

func checkPackageCycles(pkg *program.Package) error {
	if pkg == nil {
		return nil
	}

	for _, f := range pkg.Files {
		if f == nil {
			continue
		}

		g := &depGraph{
			nodes: make(map[string]bool),
			edges: make(map[string][]string),
			locs:  make(map[string]*program.Location),
			file:  f,
		}

		// 1. Collect standard variable assignments
		for _, a := range f.Assignments {
			if a == nil {
				continue
			}

			var refs []string
			collectRefs(a.Expression, &refs)

			for _, v := range a.Variables {
				if v == nil {
					continue
				}
				g.nodes[v.Name] = true
				if a.Start != nil {
					g.locs[v.Name] = a.Start
				}
				g.edges[v.Name] = append(g.edges[v.Name], refs...)

				// 2. Collect property assignment dependencies
				for _, p := range v.Properties {
					if p == nil {
						continue
					}
					propNode := fmt.Sprintf("%s.%s", v.Name, p.Name)
					g.nodes[propNode] = true
					if p.Start != nil {
						g.locs[propNode] = p.Start
					}

					var propRefs []string
					collectRefs(p.Expression, &propRefs)
					g.edges[propNode] = append(g.edges[propNode], propRefs...)

					// Only .box crops the variable definition itself (cropToBox).
					// Metadata properties (.color, .material, .min, .max, .period)
					// do not modify the variable's value or geometry.
					if p.Name == "box" {
						g.edges[v.Name] = append(g.edges[v.Name], propNode)
					}
				}
			}
		}

		// 3. Detect cycles using DFS
		visited := make(map[string]bool)
		recStack := make(map[string]bool)
		var path []string

		var dfs func(u string) ([]string, bool)
		dfs = func(u string) ([]string, bool) {
			visited[u] = true
			recStack[u] = true
			path = append(path, u)

			for _, v := range g.edges[u] {
				// Normalize reference if v is a variable and target has a property
				if recStack[v] {
					// Found cycle! Extract cycle slice
					cycle := []string{v}
					for i := len(path) - 1; i >= 0; i-- {
						cycle = append([]string{path[i]}, cycle...)
						if path[i] == v {
							break
						}
					}
					return cycle, true
				}

				if !visited[v] {
					if cycle, found := dfs(v); found {
						return cycle, true
					}
				}
			}

			recStack[u] = false
			path = path[:len(path)-1]
			return nil, false
		}

		for node := range g.nodes {
			if !visited[node] {
				if cycle, found := dfs(node); found {
					// Only report cycle if it involves a property node
					hasProperty := false
					for _, n := range cycle {
						if strings.Contains(n, ".") {
							hasProperty = true
							break
						}
					}

					if hasProperty {
						// Format cycle trace: f.box -> target -> z -> f.box
						trace := strings.Join(cycle, " -> ")
						loc := g.locs[cycle[0]]
						return fileErr(g.file, loc, fmt.Sprintf("cyclic dependency detected: %s", trace))
					}
				}
			}
		}
	}

	return nil
}

func collectRefs(op *program.Operand, refs *[]string) {
	if op == nil {
		return
	}

	switch t := op.T.(type) {
	case *program.Operand_Reference:
		if t.Reference != nil && t.Reference.Prefix == "" && t.Reference.Name != "" {
			name := t.Reference.Name
			if name != program.PVarName && name != program.TimeVarName && !program.IsBuiltin(name) && name != "auto" {
				*refs = append(*refs, name)
			}
		}
	case *program.Operand_PropRef:
		if t.PropRef != nil && t.PropRef.Reference != nil {
			base := t.PropRef.Reference.Name
			prop := t.PropRef.Property
			if base != "" && prop != "" {
				*refs = append(*refs, fmt.Sprintf("%s.%s", base, prop))
				*refs = append(*refs, base)
			}
		}
	case *program.Operand_Expression:
		if t.Expression != nil {
			// Check if Operation_BOUNDINGBOX
			if t.Expression.Operation == program.Operation_BOUNDINGBOX && len(t.Expression.Operands) > 0 {
				if ref, ok := t.Expression.Operands[0].T.(*program.Operand_Reference); ok && ref.Reference != nil {
					*refs = append(*refs, fmt.Sprintf("%s.box", ref.Reference.Name))
					*refs = append(*refs, ref.Reference.Name)
				}
			}
			// A frame read (s.u, s.v) depends on the frame declared, as a box
			// read does, so s.u = f(s.u) is caught as a cycle.
			if frame := frameAccessorName(t.Expression.Operation); frame != "" && len(t.Expression.Operands) > 0 {
				if ref, ok := t.Expression.Operands[0].T.(*program.Operand_Reference); ok && ref.Reference != nil {
					*refs = append(*refs, fmt.Sprintf("%s.%s", ref.Reference.Name, frame))
					*refs = append(*refs, ref.Reference.Name)
				}
			}
			for i, child := range t.Expression.Operands {
				if t.Expression.Operation == program.Operation_ACCESS && i > 0 {
					continue
				}
				collectRefs(child, refs)
			}
		}
	case *program.Operand_FnCall:
		if t.FnCall != nil {
			if t.FnCall.Reference != nil && t.FnCall.Reference.Prefix == "" && t.FnCall.Reference.Name != "" && !program.IsBuiltin(t.FnCall.Reference.Name) {
				*refs = append(*refs, t.FnCall.Reference.Name)
			}
			for _, arg := range t.FnCall.Arguments {
				collectRefs(arg, refs)
			}
		}
	}
}

// frameAccessorName is the property a frame read operation reads ("u" for
// SURFACEU, "v" for SURFACEV), or "" for any other operation.
func frameAccessorName(operation program.Operation) string {
	switch operation {
	case program.Operation_SURFACEU:
		return program.UAccessor
	case program.Operation_SURFACEV:
		return program.VAccessor
	}
	return ""
}

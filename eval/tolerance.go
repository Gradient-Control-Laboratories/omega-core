//go:build cgo

package eval

import (
	"context"
	"fmt"
	"strings"

	"github.com/Gradient-Control-Laboratories/omega-core/parser"
)

// tolUnitPrefix names the unit parameter certify ranges in place of a
// toleranced declaration: x becomes x's value + lo + (hi - lo) * u.
const tolUnitPrefix = "gslTolU_"

// TolerancedNames lists the document's top-level declarations that carry a
// working tolerance (x.tolerance = tolerance.limits(lo, hi)), in order.
func TolerancedNames(code string) ([]string, error) {
	res, err := parser.ParsePackage([]string{code})
	if err != nil {
		return nil, fmt.Errorf("parse document for tolerances: %w", err)
	}
	var names []string
	seen := map[string]bool{}
	for _, f := range res.Package.Files {
		for _, a := range f.Assignments {
			if len(a.Variables) != 1 || a.Variables[0] == nil {
				continue
			}
			v := a.Variables[0]
			for _, p := range v.Properties {
				if p != nil && p.Name == "tolerance" && !seen[v.Name] {
					seen[v.Name] = true
					names = append(names, v.Name)
				}
			}
		}
	}
	return names, nil
}

// CertifyOverTolerances is CertifyParams over every working tolerance the
// document declares: each toleranced declaration x is read as x's value + lo +
// (hi - lo) * u, its limits [lo, hi] = x.tolerance, with u ranging [0, 1], so
// the verdict holds for every value each takes between its limits, and, for a
// claim that is pointwise in p, for limits that vary over the part as well.
// Design sliders keep their values. A disproof names each tolerance's failing
// share of its limits in CounterParams, keyed by the declaration's name, from
// 0 at its low limit to 1 at its high one.
func (e *Evaluator) CertifyOverTolerances(ctx context.Context, manifest map[string]string, code, defs, expr string, region [6]float64, threshold, tolerance float64) (*CertifyResult, error) {
	names, err := TolerancedNames(code)
	if err != nil {
		return nil, err
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("document declares no tolerances (x.tolerance = tolerance.limits(lo, hi)) to certify over")
	}
	// Each unit runs over [0, w], w the tolerance's width at the region's
	// centre, so a unit moves the value by about one unit, as the value
	// itself would: certify resolves a parameter by its leverage, and one
	// whose leverage is a width of micrometres would otherwise be split a
	// thousand times finer than space.
	center := [][3]float64{{(region[0] + region[3]) / 2, (region[1] + region[4]) / 2, (region[2] + region[5]) / 2}}
	widthExprs := make([]string, len(names))
	for i, name := range names {
		widthExprs[i] = name + ".tolerance[1] - " + name + ".tolerance[0]"
	}
	groups, err := e.EvalGsl(ctx, manifest, code, defs, widthExprs, center)
	if err != nil {
		return nil, fmt.Errorf("reading the tolerances' widths: %w", err)
	}
	widths := make(map[string]float64, len(names))
	overrides := make(map[string]string, len(names))
	params := make(map[string][2]float64, len(names))
	for i, name := range names {
		w := float64(groups[i].Samples[0][0].Value)
		if !(w > 0) {
			return nil, fmt.Errorf("tolerance on %q is empty: its limits are %g wide at the region's centre", name, w)
		}
		value, ok := declarationValue(code, name)
		if !ok {
			return nil, fmt.Errorf("tolerance on %q: it is not a single top-level declaration", name)
		}
		u := tolUnitPrefix + name
		overrides[name] = fmt.Sprintf("@range(0, %g) %s = %g;\n%s = (%s) + %s.tolerance[0] + (%s.tolerance[1] - %s.tolerance[0]) * %s / %g;",
			w, u, w/2, name, value, name, name, name, u, w)
		params[u] = [2]float64{0, w}
		widths[name] = w
	}
	result, err := e.CertifyParams(ctx, manifest, substituteTopLevelDecls(code, overrides), defs, expr, region, threshold, tolerance, params)
	if err != nil || result == nil {
		return result, err
	}
	result.Params = shares(result.Params, widths)
	result.CounterParams = shares(result.CounterParams, widths)
	result.WorstParams = shares(result.WorstParams, widths)
	return result, nil
}

// shares keys a parameter map by the toleranced declarations' names, each
// range as shares of its tolerance, 0 at the low limit and 1 at the high.
func shares(m map[string][2]float64, widths map[string]float64) map[string][2]float64 {
	if m == nil {
		return nil
	}
	out := make(map[string][2]float64, len(m))
	for k, v := range m {
		name := strings.TrimPrefix(k, tolUnitPrefix)
		w := widths[name]
		out[name] = [2]float64{v[0] / w, v[1] / w}
	}
	return out
}

// declarationValue is the source of the value a single top-level declaration
// of name assigns, "0.0065" for zinc = 0.0065;, with any annotations before it
// left out; ok is false when there is none.
func declarationValue(code, name string) (string, bool) {
	for _, s := range topLevelStatements(code) {
		if s.name != name {
			continue
		}
		decl := strings.TrimSpace(s.text[s.keep:])
		// Annotations first (@range(...)), then the name.
		for strings.HasPrefix(decl, "@") {
			open := strings.IndexByte(decl, '(')
			space := strings.IndexAny(decl, " \t\n")
			if open >= 0 && (space < 0 || open < space) {
				depth, i := 0, open
				for ; i < len(decl); i++ {
					if decl[i] == '(' {
						depth++
					} else if decl[i] == ')' {
						depth--
						if depth == 0 {
							break
						}
					}
				}
				decl = strings.TrimSpace(decl[i+1:])
			} else if space >= 0 {
				decl = strings.TrimSpace(decl[space:])
			} else {
				return "", false
			}
		}
		if !strings.HasPrefix(decl, name) {
			continue
		}
		rest := strings.TrimSpace(decl[len(name):])
		if !strings.HasPrefix(rest, "=") || strings.HasPrefix(rest, "==") {
			continue // a property statement (name.tolerance = ...), or not a declaration
		}
		value := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(rest[1:]), ";"))
		return value, value != ""
	}
	return "", false
}

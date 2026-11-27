package fix

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Gradient-Control-Laboratories/omega-core/parser"
	"github.com/Gradient-Control-Laboratories/omega-core/program"
)

var removedSweepLoftFns = map[string]bool{
	"sweepLoft":                        true,
	"sweepLoftScale":                   true,
	"sweepLoftScaleSmooth":             true,
	"sweepLoftScaleSmoothSeparateCaps": true,
	"bezierTubeDir":                    true,
}

var roundExactFns = map[string]bool{
	"unionRoundExact":     true,
	"intersectRoundExact": true,
}

var stdlibFixer = Fixer{
	Name: "stdlib",
	Doc:  "migrate breaking stdlib calls and report manual migration hints for removed functions",
	Help: `The stdlib fixer updates GSL standard library calls across breaking changes:

1. Automatic migrations:
   - Renames members the standard library renamed, through the package's prefix
     (its last path segment, an explicit import name, or bare under import .):
       honeycomb.squareHoneycomb           -> honeycomb.rectangular
       honeycomb.squareDiagonalHoneycomb   -> honeycomb.rectangularDiagonal
       honeycomb.squareHoneycombReinforced -> honeycomb.rectangularReinforced
       honeycomb.hexagonalHoneycomb        -> honeycomb.hexagonal
       honeycomb.triangularHoneycomb       -> honeycomb.triangular
       honeycomb.octetHoneycomb            -> honeycomb.octet
   - Replaces removed 2-argument thunk calls unionRoundExact(a, b) and
     intersectRoundExact(a, b) with value-based calls with radius 0:
       unionRoundExact(() { return a; }, () { return b; })
         -> unionRoundExact(a, b, 0)
   - Replaces removed 4-argument thunk calls unionRoundExact(a, b, n1, n2) and
     intersectRoundExact(a, b, n1, n2) with 5-argument value calls with radius 0:
       unionRoundExact(() { return a; }, () { return b; }, n1, n2)
         -> unionRoundExact(a, b, 0, n1, n2)

2. Manual migration hints (reported to output, file left unchanged):
   - Calls to removed builtin 3D lofting functions (sweepLoft, sweepLoftScale,
     sweepLoftScaleSmooth, sweepLoftScaleSmoothSeparateCaps, bezierTubeDir)
     report a line-numbered suggestion to migrate planar sweeps to sweep,
     sweepOpen, or sweepOpenEdges with edgeBezier, or keep a local library copy.
`,
	Apply: func(src string) (string, bool, error) {
		out, changed, _, err := applyStdlibFix(src)
		return out, changed, err
	},
	ApplyWithNotes: applyStdlibFix,
}

type textEdit struct {
	start, end int
	text       string
}

func applyStdlibFix(src string) (string, bool, []Note, error) {
	if src == "" {
		return src, false, nil, nil
	}

	useCRLF := strings.Contains(src, "\r\n")
	normalized := src
	if useCRLF {
		normalized = strings.ReplaceAll(src, "\r\n", "\n")
	}

	curr := renameMembers(normalized)
	for pass := 0; pass < 8; pass++ {
		res, err := parser.ParsePackage([]string{curr})
		if err != nil || res.Package == nil || len(res.Package.Files) == 0 || res.Package.Files[0] == nil {
			break
		}
		edits := collectRoundExactEdits(curr, res.Package.Files[0])
		if len(edits) == 0 {
			break
		}
		applied := applyNonOverlappingEdits(curr, edits)
		if applied == curr {
			break
		}
		curr = applied
	}

	var notes []Note
	if res, err := parser.ParsePackage([]string{curr}); err == nil && res.Package != nil && len(res.Package.Files) > 0 && res.Package.Files[0] != nil {
		notes = collectRemovedStdlibNotes(res.Package.Files[0])
	}

	out := curr
	if useCRLF {
		out = strings.ReplaceAll(out, "\n", "\r\n")
	}
	return out, out != src, notes, nil
}

func collectRoundExactEdits(src string, file *program.PackageFile) []textEdit {
	fnIndex := buildFnIndex(file)
	var edits []textEdit

	visitCallsInFile(file, func(op *program.Operand, fc *program.FnCall) {
		if fc.Reference == nil || (fc.Reference.Prefix != "" && fc.Reference.Prefix != "builtin") {
			return
		}
		if !roundExactFns[fc.Reference.Name] {
			return
		}
		n := len(fc.Arguments)
		if n != 2 && n != 4 {
			return
		}
		arg0, arg1 := fc.Arguments[0], fc.Arguments[1]
		if !isNullaryThunkArg(arg0, fnIndex) && !isNullaryThunkArg(arg1, fnIndex) {
			return
		}
		if arg0 == nil || arg1 == nil || arg0.Start == nil || arg0.End == nil || arg1.Start == nil || arg1.End == nil {
			return
		}

		for i := 0; i < 2; i++ {
			if e, ok := unwrapThunkArgEdit(src, fc.Arguments[i], fnIndex); ok {
				edits = append(edits, e)
			}
		}
		if n == 2 {
			end1 := int(arg1.End.Offset)
			if end1 >= 0 && end1 <= len(src) {
				edits = append(edits, textEdit{start: end1, end: end1, text: ", 0"})
			}
		} else {
			arg2 := fc.Arguments[2]
			if arg2 != nil && arg2.Start != nil {
				start2 := int(arg2.Start.Offset)
				if start2 >= 0 && start2 <= len(src) {
					edits = append(edits, textEdit{start: start2, end: start2, text: "0, "})
				}
			}
		}
	})

	return edits
}

func unwrapThunkArgEdit(src string, arg *program.Operand, fnIndex map[string]*program.Function) (textEdit, bool) {
	if arg == nil || arg.Start == nil || arg.End == nil {
		return textEdit{}, false
	}
	ref, ok := arg.T.(*program.Operand_Reference)
	if !ok || ref.Reference == nil || ref.Reference.Prefix != "" {
		return textEdit{}, false
	}
	fn := fnIndex[ref.Reference.Name]
	if fn == nil || len(fn.Parameters) != 0 {
		return textEdit{}, false
	}
	start, end := int(arg.Start.Offset), int(arg.End.Offset)
	if start < 0 || end > len(src) || start > end {
		return textEdit{}, false
	}

	if strings.HasPrefix(ref.Reference.Name, program.AnonFnPrefix) {
		if len(fn.Assignments) == 0 && len(fn.Returns) == 1 {
			ret := fn.Returns[0]
			if ret != nil && ret.Start != nil && ret.End != nil {
				rStart, rEnd := int(ret.Start.Offset), int(ret.End.Offset)
				if rStart >= 0 && rEnd <= len(src) && rStart <= rEnd {
					return textEdit{start: start, end: end, text: src[rStart:rEnd]}, true
				}
			}
		}
		return textEdit{}, false
	}

	// Named zero-arg function passed as a value: call it directly.
	return textEdit{start: end, end: end, text: "()"}, true
}

func isNullaryThunkArg(arg *program.Operand, fnIndex map[string]*program.Function) bool {
	if arg == nil {
		return false
	}
	ref, ok := arg.T.(*program.Operand_Reference)
	if !ok || ref.Reference == nil || ref.Reference.Prefix != "" {
		return false
	}
	fn := fnIndex[ref.Reference.Name]
	return fn != nil && len(fn.Parameters) == 0
}

func buildFnIndex(file *program.PackageFile) map[string]*program.Function {
	index := make(map[string]*program.Function)
	var addFnTypes func(fts []*program.FnType)
	addFnTypes = func(fts []*program.FnType) {
		for _, ft := range fts {
			if ft == nil {
				continue
			}
			for _, g := range ft.Groups {
				if g == nil {
					continue
				}
				for _, fn := range g.Functions {
					if fn == nil {
						continue
					}
					if len(fn.Parameters) == 0 {
						index[ft.Name] = fn
					}
					addFnTypes(fn.FnTypes)
				}
			}
		}
	}
	addFnTypes(file.FnTypes)
	return index
}

func collectRemovedStdlibNotes(file *program.PackageFile) []Note {
	var notes []Note
	visitCallsInFile(file, func(op *program.Operand, fc *program.FnCall) {
		if fc.Reference == nil || (fc.Reference.Prefix != "" && fc.Reference.Prefix != "builtin") {
			return
		}
		if !removedSweepLoftFns[fc.Reference.Name] {
			return
		}
		line := 0
		if op.Start != nil {
			line = int(op.Start.Line)
		}
		notes = append(notes, Note{
			Line: line,
			Message: fmt.Sprintf(
				"%s was removed from builtin; for constant-section planar sweeps use sweep, sweepOpen, or sweepOpenEdges with edgeBezier, or keep a local copy in your library",
				fc.Reference.Name,
			),
		})
	})
	return notes
}

func applyNonOverlappingEdits(src string, edits []textEdit) string {
	sort.SliceStable(edits, func(i, j int) bool {
		if edits[i].start != edits[j].start {
			return edits[i].start < edits[j].start
		}
		return edits[i].end < edits[j].end
	})
	var kept []textEdit
	lastEnd := -1
	for _, e := range edits {
		if e.start < lastEnd {
			continue
		}
		kept = append(kept, e)
		lastEnd = e.end
	}
	out := src
	for i := len(kept) - 1; i >= 0; i-- {
		e := kept[i]
		out = out[:e.start] + e.text + out[e.end:]
	}
	return out
}

func visitCallsInFile(file *program.PackageFile, fn func(op *program.Operand, fc *program.FnCall)) {
	if file == nil {
		return
	}
	var walkOp func(op *program.Operand)
	walkOp = func(op *program.Operand) {
		if op == nil || op.T == nil {
			return
		}
		switch t := op.T.(type) {
		case *program.Operand_Expression:
			if t.Expression != nil {
				for _, child := range t.Expression.Operands {
					walkOp(child)
				}
			}
		case *program.Operand_FnCall:
			if t.FnCall != nil {
				fn(op, t.FnCall)
				for _, arg := range t.FnCall.Arguments {
					walkOp(arg)
				}
			}
		}
	}
	var walkFnTypes func(fts []*program.FnType)
	walkFnTypes = func(fts []*program.FnType) {
		for _, ft := range fts {
			if ft == nil {
				continue
			}
			for _, g := range ft.Groups {
				if g == nil {
					continue
				}
				for _, f := range g.Functions {
					if f == nil {
						continue
					}
					walkFnTypes(f.FnTypes)
					for _, a := range f.Assignments {
						walkOp(a.Expression)
						for _, v := range a.Variables {
							for _, p := range v.Properties {
								walkOp(p.Expression)
							}
						}
					}
					for _, r := range f.Returns {
						walkOp(r)
					}
				}
			}
		}
	}

	walkFnTypes(file.FnTypes)
	for _, a := range file.Assignments {
		walkOp(a.Expression)
		for _, v := range a.Variables {
			for _, p := range v.Properties {
				walkOp(p.Expression)
			}
		}
	}
	for _, r := range file.Returns {
		walkOp(r)
	}
}

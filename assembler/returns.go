package assembler

import (
	"fmt"
	"sort"

	"gradientcontrol.com/omega/core/assembler/meshplan"
	"gradientcontrol.com/omega/core/assembler/source"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/internal/link"
	"gradientcontrol.com/omega/core/internal/scope"
	"gradientcontrol.com/omega/core/program"
)

type variantSlice struct {
	name   string
	source string
	offset int
	count  int
}

// returnPartition tracks the contiguous slice layout of prog.Returns()
// (i.e. prog.Primary.Files[len(prog.Primary.Files)-1].Returns) across the
// single-pass compiler pipeline:
//
//  1. prog.Returns()[0 : geomCount]:
//     The primary package's geometry returns (0 for an audio-only
//     `return audio(left, right);` document or a variation-only library module).
//     prog.ColorReturns() and prog.MaterialReturns() are 1-to-1 with this slice.
//
//  2. prog.Returns()[geomCount : geomCount+audioCount]:
//     The 0, 1 (mono), or 2 (stereo) scalar channel operands unwrapped from a
//     trailing top-level `audio(...)` call in the primary package.
//
//  3. prog.Returns()[geomCount+audioCount : geomCount+audioCount+totalVariantCount]:
//     Any caller-supplied WithVariants return operands, packed contiguously in
//     variantSlices order.
type returnPartition struct {
	geomCount      int
	audioCount     int
	primarySources []string
	variantSlices  []variantSlice
	variantErrors  map[string]string
}

// partitionReturns validates and partitions top-level return expressions on prog
// at Step 1 of Assemble before reference linking.
//
// When a trailing `audio(...)` call is present as the final element of
// prog.Returns(), partitionReturns validates its placement and arity (1 or 2
// arguments) and unwraps the channel operands in-place into
// prog.Returns()[geomCount : geomCount+audioCount].
func partitionReturns(
	prog *program.Program,
	primarySources []string,
	variants map[string]string,
	enableDeadcodeRemoval bool,
) (*returnPartition, error) {
	rp := &returnPartition{
		primarySources: primarySources,
	}

	// Reject audio(...) anywhere in non-final files of the primary package.
	for fileIdx := 0; fileIdx < len(prog.Primary.Files)-1; fileIdx++ {
		if badOp := findAudioCallInFile(prog.Primary.Files[fileIdx], false); badOp != nil {
			return nil, rp.errAt(prog, fileIdx, badOp.Start, fmt.Errorf("audio(...) can only be used as the final top-level return expression"))
		}
	}

	lastIdx := len(prog.Primary.Files) - 1
	lastFile := prog.Primary.Files[lastIdx]
	if badOp := findAudioCallInFile(lastFile, true); badOp != nil {
		return nil, rp.errAt(prog, lastIdx, badOp.Start, fmt.Errorf("audio(...) can only be used as the final top-level return expression"))
	}

	returns := prog.Returns()
	hasTrailingAudio := false
	for idx, ret := range returns {
		if isAudioCall(ret) {
			if idx < len(returns)-1 {
				return nil, rp.errAt(prog, lastIdx, ret.Start, fmt.Errorf("audio(...) must be the last element in the return statement"))
			}
			args := ret.GetFnCall().GetArguments()
			if len(args) != 1 && len(args) != 2 {
				return nil, rp.errAt(prog, lastIdx, ret.Start, fmt.Errorf("audio(...) requires 1 (mono) or 2 (stereo) arguments, got %d", len(args)))
			}
			for _, arg := range args {
				if nested := findAudioCallInOperand(arg); nested != nil {
					return nil, rp.errAt(prog, lastIdx, nested.Start, fmt.Errorf("audio(...) can only be used as the final top-level return expression"))
				}
			}
			hasTrailingAudio = true
			rp.geomCount = len(returns) - 1
			rp.audioCount = len(args)
			unwrapped := make([]*program.Operand, 0, rp.geomCount+rp.audioCount)
			unwrapped = append(unwrapped, returns[:rp.geomCount]...)
			unwrapped = append(unwrapped, args...)
			prog.SetReturns(unwrapped)
			break
		}
		if nested := findAudioCallInOperand(ret); nested != nil {
			return nil, rp.errAt(prog, lastIdx, nested.Start, fmt.Errorf("audio(...) can only be used as the final top-level return expression"))
		}
	}

	if !hasTrailingAudio {
		rp.geomCount = len(returns)
	}

	if rp.geomCount == 0 && rp.audioCount == 0 && len(variants) == 0 && enableDeadcodeRemoval {
		return nil, fmt.Errorf("primary package must have a return statement")
	}

	return rp, nil
}

// AttachVariants parses each caller-supplied variant return statement in sorted
// key order, validates its symbol references against the linked primary package,
// records any per-variant error in rp.variantErrors without failing the overall
// assembly, and appends valid variant return operands to
// prog.Returns()[geomCount+audioCount : geomCount+audioCount+totalVariantCount].
func (rp *returnPartition) AttachVariants(
	prog *program.Program,
	srcMgr source.Manager,
	variants map[string]string,
	parsePkg ParsePackageFunc,
) {
	if len(variants) == 0 {
		return
	}
	rp.variantErrors = make(map[string]string)
	names := make([]string, 0, len(variants))
	for name := range variants {
		names = append(names, name)
	}
	sort.Strings(names)

	lastFile := prog.Primary.Files[len(prog.Primary.Files)-1]
	for _, name := range names {
		retSrc := variants[name]
		res, err := parsePkg([]string{retSrc})
		if err != nil {
			rp.variantErrors[name] = fmt.Sprintf("parse variant return %q: %v", retSrc, err)
			continue
		}
		pkg := res.Package
		if len(pkg.Returns()) == 0 {
			rp.variantErrors[name] = fmt.Sprintf("variant %s return expression has no returns", name)
			continue
		}
		if findAudioCallInFile(pkg.Files[0], false) != nil {
			rp.variantErrors[name] = "audio(...) cannot be used in a variant return"
			continue
		}

		varReturns := pkg.Returns()
		savedReturns := append([]*program.Operand(nil), prog.Returns()...)
		origImports := append([]*program.Import(nil), lastFile.Imports...)
		origDeps := append([]*program.Package(nil), prog.Dependencies...)

		offset := len(savedReturns)
		combined := make([]*program.Operand, 0, offset+len(varReturns))
		combined = append(combined, savedReturns...)
		combined = append(combined, varReturns...)
		prog.SetReturns(combined)

		if err := link.References(prog, srcMgr, parsePkg); err != nil {
			rp.variantErrors[name] = err.Error()
			prog.SetReturns(savedReturns)
			lastFile.Imports = origImports
			prog.Dependencies = origDeps
			continue
		}

		rp.variantSlices = append(rp.variantSlices, variantSlice{
			name:   name,
			source: retSrc,
			offset: offset,
			count:  len(varReturns),
		})
	}
}

// WithGeometryReturnsOnly temporarily detaches prog.Returns()[geomCount:] (audio
// and variant returns) while executing fn so passes that only apply to primary
// geometry returns (such as property.MarkParts, property.DropDeclaredBoxes, and
// property.LowerReturnProperties) observe only prog.Returns()[:geomCount] and
// produce 1-to-1 ColorReturns and MaterialReturns of length geomCount, then
// re-attaches the variation return slices before returning.
func (rp *returnPartition) WithGeometryReturnsOnly(
	prog *program.Program,
	fn func(*program.Program) (*program.Program, error),
) (*program.Program, error) {
	returns := prog.Returns()
	if len(returns) == rp.geomCount {
		return fn(prog)
	}
	tail := append([]*program.Operand(nil), returns[rp.geomCount:]...)
	prog.SetReturns(returns[:rp.geomCount])

	outProg, err := fn(prog)
	if outProg != nil {
		geomReturns := outProg.Returns()[:rp.geomCount]
		combined := make([]*program.Operand, 0, len(geomReturns)+len(tail))
		combined = append(combined, geomReturns...)
		combined = append(combined, tail...)
		outProg.SetReturns(combined)
	}
	return outProg, err
}

// ValidateAudioTypes verifies after Step 3 types.Analyze that every unwrapped
// audio channel operand in prog.Returns()[geomCount : geomCount+audioCount]
// has scalar numeric type.
func (rp *returnPartition) ValidateAudioTypes(prog *program.Program) error {
	if rp.audioCount == 0 {
		return nil
	}
	lastIdx := len(prog.Primary.Files) - 1
	returns := prog.Returns()
	for i := 0; i < rp.audioCount; i++ {
		op := returns[rp.geomCount+i]
		if !op.GetType().IsScalar() {
			return rp.errAt(
				prog,
				lastIdx,
				op.Start,
				fmt.Errorf("audio(...) argument %d must be a scalar number, got %s", i+1, op.GetType().TypeString()),
			)
		}
	}
	return nil
}

// PopVariations extracts the optimized audio return operands from
// prog.Returns()[geomCount : geomCount+audioCount] and slices each valid
// variant return in prog.Returns()[geomCount+audioCount:] via
// meshplan.SliceDocument after Step 7 deadcode.Remove, then restores
// prog.Returns() to prog.Returns()[:geomCount] (or nil when geomCount == 0)
// before textures.Extract and Step 8 ExtractMeshPlan run.
func (rp *returnPartition) PopVariations(
	prog *program.Program,
	idPrefix string,
) ([]meshplan.SdfSubtree, map[string]VariantResult) {
	hasVariants := len(rp.variantSlices) > 0 || len(rp.variantErrors) > 0
	if rp.audioCount == 0 && !hasVariants {
		return nil, nil
	}
	returns := append([]*program.Operand(nil), prog.Returns()...)

	var audioReturns []meshplan.SdfSubtree
	if rp.audioCount > 0 {
		audioReturns = make([]meshplan.SdfSubtree, rp.audioCount)
		for i := 0; i < rp.audioCount; i++ {
			audioReturns[i] = meshplan.SdfSubtree{
				ID:     fmt.Sprintf("audio-%d", i),
				Return: returns[rp.geomCount+i],
			}
		}
	}

	var variants map[string]VariantResult
	if hasVariants {
		variants = make(map[string]VariantResult, len(rp.variantSlices)+len(rp.variantErrors))
		for name, errMsg := range rp.variantErrors {
			variants[name] = VariantResult{Error: errMsg}
		}
		for _, vs := range rp.variantSlices {
			prog.SetReturns(returns[vs.offset : vs.offset+vs.count])
			plan, subtrees, err := meshplan.SliceDocument(prog, fmt.Sprintf("%svariant-%s-", idPrefix, vs.name), vs.source)
			if err != nil {
				variants[vs.name] = VariantResult{Error: fmt.Sprintf("prepare geometry: %v", err)}
				continue
			}
			variants[vs.name] = VariantResult{
				MeshPlan:    plan,
				SdfSubtrees: subtrees,
			}
		}
	}

	if rp.geomCount == 0 {
		prog.SetReturns(nil)
	} else {
		prog.SetReturns(returns[:rp.geomCount])
	}
	return audioReturns, variants
}

func (rp *returnPartition) errAt(prog *program.Program, fileIdx int, loc *program.Location, err error) error {
	// Invariant: fileIdx is a valid index into prog.Primary.Files.
	src := prog.Primary.Files[fileIdx].Source
	if src == "" && fileIdx < len(rp.primarySources) {
		src = rp.primarySources[fileIdx]
	}
	buildErr := scope.ErrLoc(prog.Primary.Path, "", loc, err)
	return build.ErrWithSrc(src, buildErr)
}

func isAudioCall(op *program.Operand) bool {
	fnCall := op.GetFnCall()
	if fnCall == nil {
		return false
	}
	ref := fnCall.GetReference()
	return ref != nil && ref.Prefix == "" && ref.Name == "audio"
}

func findAudioCallInFile(file *program.PackageFile, skipReturns bool) *program.Operand {
	for _, a := range file.Assignments {
		if op := findAudioCallInOperand(a.Expression); op != nil {
			return op
		}
		for _, v := range a.Variables {
			for _, prop := range v.Properties {
				if op := findAudioCallInOperand(prop.Expression); op != nil {
					return op
				}
			}
		}
	}
	for _, ft := range file.FnTypes {
		if op := findAudioCallInFnType(ft); op != nil {
			return op
		}
	}
	if !skipReturns {
		for _, ret := range file.Returns {
			if op := findAudioCallInOperand(ret); op != nil {
				return op
			}
		}
	}
	return nil
}

func findAudioCallInFnType(ft *program.FnType) *program.Operand {
	for _, g := range ft.Groups {
		for _, fn := range g.Functions {
			for _, a := range fn.Assignments {
				if op := findAudioCallInOperand(a.Expression); op != nil {
					return op
				}
			}
			for _, ret := range fn.Returns {
				if op := findAudioCallInOperand(ret); op != nil {
					return op
				}
			}
			for _, nested := range fn.FnTypes {
				if op := findAudioCallInFnType(nested); op != nil {
					return op
				}
			}
		}
	}
	return nil
}

func findAudioCallInOperand(op *program.Operand) *program.Operand {
	// Invariant: op is guaranteed non-nil in valid parsed ASTs.
	if isAudioCall(op) {
		return op
	}
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		for _, sub := range t.Expression.GetOperands() {
			if found := findAudioCallInOperand(sub); found != nil {
				return found
			}
		}
	case *program.Operand_FnCall:
		for _, arg := range t.FnCall.GetArguments() {
			if found := findAudioCallInOperand(arg); found != nil {
				return found
			}
		}
	}
	return nil
}

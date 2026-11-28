package fix

import (
	"fmt"
	"strings"

	"github.com/Gradient-Control-Laboratories/omega-core/program"
)

// conditionFn is a removed three-argument form that took a material condition:
// the package it is called through, and how its maximum and minimum material
// sizes are written without it.
type conditionFn struct {
	pkg string
	// size is the size the deviation is added to, from the first argument.
	size func(arg string) string
	// max and min name the tolerance package's deviation for each condition.
	max, min string
}

var conditionFns = map[string]conditionFn{
	"shaftDia":   {pkg: "tolerance", size: func(a string) string { return a }, max: "shaftUpper", min: "shaftLower"},
	"holeDia":    {pkg: "tolerance", size: func(a string) string { return a }, max: "holeLower", min: "holeUpper"},
	"seatDia":    {pkg: "bearing", size: func(a string) string { return "bearing.bore(" + a + ")" }, max: "shaftUpper", min: "shaftLower"},
	"housingDia": {pkg: "bearing", size: func(a string) string { return "bearing.outerDia(" + a + ")" }, max: "holeLower", min: "holeUpper"},
}

// collectConditionEdits rewrites each three-argument call of a removed
// condition form to the size it gave, exactly: nominal plus the condition's
// deviation, tolerance.shaftDia(25.0, tolerance.k5, tolerance.maxMaterial) ->
// (25.0 + tolerance.shaftUpper(25.0, tolerance.k5)). A condition it cannot
// read as maximum or minimum material is left, with a note.
func collectConditionEdits(src string, file *program.PackageFile) ([]textEdit, []Note) {
	var edits []textEdit
	var notes []Note
	visitCallsInFile(file, func(op *program.Operand, fc *program.FnCall) {
		if fc.Reference == nil || len(fc.Arguments) != 3 || op.Start == nil {
			return
		}
		cf, ok := conditionFns[fc.Reference.Name]
		if !ok || fc.Reference.Prefix != cf.pkg {
			return
		}
		start := int(op.Start.Offset)
		open := strings.IndexByte(src[start:], '(')
		if open < 0 {
			return
		}
		open += start
		end, args := splitCallArgs(src, open)
		if end < 0 || len(args) != 3 {
			return
		}
		dev := ""
		switch cond := strings.TrimSpace(args[2]); cond {
		case "tolerance.maxMaterial", "maxMaterial", "1", "1.0":
			dev = cf.max
		case "tolerance.minMaterial", "minMaterial", "2", "2.0":
			dev = cf.min
		default:
			notes = append(notes, Note{Line: int(op.Start.Line), Message: fmt.Sprintf(
				"%s.%s(...) lost its material condition argument; %q is neither tolerance.maxMaterial nor tolerance.minMaterial, so rewrite it by hand as the nominal plus tolerance.%s or tolerance.%s",
				cf.pkg, fc.Reference.Name, cond, cf.max, cf.min)})
			return
		}
		size := cf.size(strings.TrimSpace(args[0]))
		zone := strings.TrimSpace(args[1])
		edits = append(edits, textEdit{start: start, end: end + 1,
			text: fmt.Sprintf("(%s + tolerance.%s(%s, %s))", size, dev, size, zone)})
	})
	return edits, notes
}

// splitCallArgs reads the arguments of the call whose '(' is at open: the
// index of its matching ')' and each argument's text, split at top-level
// commas. end is -1 when the parentheses do not close.
func splitCallArgs(src string, open int) (end int, args []string) {
	depth, last := 0, open+1
	for i := open; i < len(src); i++ {
		switch src[i] {
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			depth--
			if depth == 0 {
				args = append(args, src[last:i])
				return i, args
			}
		case ',':
			if depth == 1 {
				args = append(args, src[last:i])
				last = i + 1
			}
		}
	}
	return -1, nil
}

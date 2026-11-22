package parser2

import (
	"strings"

	"gradientcontrol.com/omega/core/program"
)

// legacySliderDiagnostics is a warning for each top-level variable that sets its
// slider with `.min`, `.max` or `.period` property statements instead of an
// `@range` annotation, pointing at the first of those statements. The
// properties still compile; the warning is how a document learns they are
// deprecated, and `gsl fix -range` is how it is rewritten.
func legacySliderDiagnostics(source string, file *program.PackageFile) []program.Diagnostic {
	if file == nil {
		return nil
	}
	lines := strings.Split(source, "\n")
	var diags []program.Diagnostic
	for _, a := range file.Assignments {
		if a == nil {
			continue
		}
		for _, v := range a.Variables {
			if v == nil {
				continue
			}
			var first *program.Property
			for _, p := range v.Properties {
				if p == nil || p.Start == nil || !(p.Name == "min" || p.Name == "max" || p.Name == "period") {
					continue
				}
				if first == nil || p.Start.Offset < first.Start.Offset {
					first = p
				}
			}
			if first == nil {
				continue
			}
			diags = append(diags, legacySliderDiagnostic(lines, v.Name, first))
		}
	}
	return diags
}

func legacySliderDiagnostic(lines []string, name string, p *program.Property) program.Diagnostic {
	lineNum := int(p.Start.Line)
	column := int(p.Start.Column)
	line := ""
	if 1 <= lineNum && lineNum <= len(lines) {
		line = strings.TrimRight(lines[lineNum-1], "\r")
	}
	// Underline the `name.prop` the statement starts with.
	length := len(name) + 1 + len(p.Name)
	if column >= 1 && column-1 < len(line) {
		if eq := strings.IndexByte(line[column-1:], '='); eq > 0 {
			length = len(strings.TrimRight(line[column-1:column-1+eq], " \t"))
		}
	}
	return program.Diagnostic{
		Severity: program.SeverityWarning,
		Line:     line,
		LineNum:  lineNum,
		Column:   column,
		Length:   max(length, 1),
		Message: "the slider properties `.min`, `.max` and `.period` are deprecated; write `@range(min, max[, period])` " +
			"in front of `" + name + "` and run `gsl fix -range` to update",
	}
}

package link

import (
	"fmt"

	"gradientcontrol.com/omega/core/program"
)

// diagnosticSink collects compiler diagnostics with lazy deduplication map allocation.
type diagnosticSink struct {
	seen  map[string]struct{}
	diags []program.Diagnostic
}

func (s *diagnosticSink) add(d program.Diagnostic) {
	if s.seen == nil {
		s.seen = make(map[string]struct{})
	}
	key := fmt.Sprintf("%s:%d:%d:%d:%s", d.FilePath, d.Severity, d.LineNum, d.Column, d.Message)
	if _, exists := s.seen[key]; !exists {
		s.seen[key] = struct{}{}
		s.diags = append(s.diags, d)
	}
}

func (s *diagnosticSink) addWarning(loc *program.Location, pkgPath, filePath, msg string) {
	d := program.Diagnostic{
		Severity:    program.SeverityWarning,
		PackagePath: pkgPath,
		FilePath:    filePath,
		Message:     msg,
	}
	if loc != nil {
		d.LineNum = int(loc.Line)
		d.Column = int(loc.Column)
	}
	s.add(d)
}

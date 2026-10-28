package program

import (
	"fmt"
	"strings"
)

// DiagnosticSeverity indicates the severity level of a compiler diagnostic.
type DiagnosticSeverity int

const (
	SeverityInfo DiagnosticSeverity = iota
	SeverityWarning
	SeverityError
)

func (s DiagnosticSeverity) String() string {
	switch s {
	case SeverityInfo:
		return "info"
	case SeverityWarning:
		return "warning"
	case SeverityError:
		return "error"
	default:
		return fmt.Sprintf("severity(%d)", s)
	}
}

// Diagnostic represents a compiler diagnostic (warning, info, or structured error).
type Diagnostic struct {
	Severity    DiagnosticSeverity `json:"severity" yaml:"severity"`
	PackagePath string             `json:"packagePath,omitempty" yaml:"packagePath,omitempty"`
	FilePath    string             `json:"filePath,omitempty" yaml:"filePath,omitempty"`
	Line        string             `json:"line,omitempty" yaml:"line,omitempty"`
	LineNum     int                `json:"lineNum" yaml:"lineNum"`
	Column      int                `json:"column" yaml:"column"`
	Length      int                `json:"length,omitempty" yaml:"length,omitempty"`
	Message     string             `json:"message" yaml:"message"`
}

func (d Diagnostic) String() string {
	loc := ""
	if d.LineNum > 0 {
		loc = fmt.Sprintf("%d:%d: ", d.LineNum, d.Column)
	}
	msg := fmt.Sprintf("%s%s: %s", loc, d.Severity, d.Message)
	if d.Line != "" {
		underline := strings.Repeat(" ", max(d.Column-1, 0)) + strings.Repeat("^", max(d.Length, 1))
		msg = fmt.Sprintf("%s\n%s %s", d.Line, underline, msg)
	}
	if d.PackagePath != "" {
		msg = fmt.Sprintf("%s\n  in package %s", msg, d.PackagePath)
	}
	if d.FilePath != "" {
		msg = fmt.Sprintf("%s\n  in file %s", msg, d.FilePath)
	}
	return msg
}

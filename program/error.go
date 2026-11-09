package program

import (
	"fmt"
	"strings"
)

type Error struct {
	PackagePath string `json:"packagePath" yaml:"packagePath"`
	FilePath    string `json:"filePath" yaml:"filePath"`
	Line        string `json:"line" yaml:"line"`
	LineNum     int    `json:"lineNum" yaml:"lineNum"`
	Column      int    `json:"column" yaml:"column"`
	Length      int    `json:"length" yaml:"length"`
	Message     string `json:"message" yaml:"message"`
}

var _ error = (*Error)(nil)

func (e *Error) Error() string {
	errMsg := fmt.Sprintf("%d:%d: %s", e.LineNum, e.Column, e.Message)
	if e.Line != "" {
		underline := strings.Repeat(" ", max(e.Column-1, 0)) + strings.Repeat("^", e.Length)
		errMsg = fmt.Sprintf("%s\n%s %s", e.Line, underline, errMsg)
	}
	if e.PackagePath != "" {
		errMsg = fmt.Sprintf("%s\n  in package %s", errMsg, e.PackagePath)
	}
	if e.FilePath != "" {
		errMsg = fmt.Sprintf("%s\n  in file %s", errMsg, e.FilePath)
	}
	return errMsg
}

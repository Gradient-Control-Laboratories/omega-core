package parser

import (
	"gradientcontrol.com/omega/core/program"
)

const RequestPath = "parser/parse"
const ResponsePath = "parser/result"

type ParserRequest struct {
	ID      string   `json:"id"` // Client supplied
	Sources []string `json:"sources"`
}

type ParserResponse struct {
	ID      string          `json:"id"`      // Copied from ParserRequest
	Package []byte          `json:"package"` // package from new Program model
	Source  string          `json:"source"`
	Errors  []program.Error `json:"errors"`
}

type ParseResult struct {
	Package *program.Package
	Errors  []program.Error
}

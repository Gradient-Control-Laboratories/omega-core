package parser2

import (
	"strconv"

	participle "github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

func (d *TopLevel) parseImports(lex *lexer.PeekingLexer) error {
	imports, err := importsParser.ParseFromLexer(lex)
	if isImportError(lex, err) {
		for _, imp := range imports.Imports {
			if imp.EndPos.Line > 0 {
				d.Imports = append(d.Imports, imp)
			}
		}
		return &participle.ParseError{
			Pos: lex.Peek().Pos,
			Msg: "expected 'import <optional-prefix> \"<path>\";', where the path is enclosed in double quotes"}
	}

	d.Imports = imports.Imports
	return nil
}

func (d *TopLevel) parseAssetImports(lex *lexer.PeekingLexer) error {
	for {
		kind := lex.Peek().Value
		if kind != "importMesh" && kind != "importStep" && kind != "importWarp" {
			return nil
		}
		start := lex.Next()
		indexTok := lex.Peek()
		if indexTok.Type != gslInt {
			slot := "mesh slot index"
			if kind == "importWarp" {
				slot = "warp slot index"
			}
			return &participle.UnexpectedTokenError{Unexpected: *indexTok, Expect: slot}
		}
		lex.Next()
		index, err := strconv.Atoi(indexTok.Value)
		if err != nil {
			return err
		}
		pathTok := lex.Peek()
		if pathTok.Type != gslString {
			want := "mesh filename"
			if kind == "importWarp" {
				want = "simulation filename"
			}
			return &participle.UnexpectedTokenError{Unexpected: *pathTok, Expect: want}
		}
		lex.Next()
		end := lex.Peek()
		if end.Value != ";" {
			return &participle.UnexpectedTokenError{Unexpected: *end, Expect: ";"}
		}
		lex.Next()
		switch kind {
		case "importMesh":
			d.MeshImports = append(d.MeshImports, &MeshImport{
				Pos: start.Pos, Index: index, Path: pathTok.Value, EndPos: end.Pos,
			})
		case "importWarp":
			d.WarpImports = append(d.WarpImports, &WarpImport{
				Pos: start.Pos, Index: index, Path: pathTok.Value, EndPos: end.Pos,
			})
		default:
			d.StepImports = append(d.StepImports, &StepImport{
				Pos: start.Pos, Index: index, Path: pathTok.Value, EndPos: end.Pos,
			})
		}
	}
}

func isImportError(lex *lexer.PeekingLexer, err error) bool {
	if err == nil {
		return false
	}

	// Mark and restore the lexer state after checking for an import error

	checkpoint := lex.MakeCheckpoint()
	defer lex.LoadCheckpoint(checkpoint)

	// The error is probably because all imports have been parsed
	// and the next token is a declaration or return.
	// Check the error to see if it's related to an import statement
	// and if so, return a more helpful error message.

	token := lex.Peek()
	if token.Value != "import" {
		return false
	}
	value := token.Value
	for token.Type == gslIdent || token.Type == gslDot || token.Type == gslString {
		token = lex.Next()
		value = token.Value
	}
	return value == ";"
}

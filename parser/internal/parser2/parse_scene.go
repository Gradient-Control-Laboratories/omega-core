package parser2

import (
	"fmt"
	"strconv"

	participle "github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

func (d *TopLevel) parseScene(lex *lexer.PeekingLexer) error {
	peek := lex.Peek()
	if peek.Value != "@" {
		return nil
	}

	checkpoint := lex.MakeCheckpoint()
	atTok := lex.Next() // consume '@'

	identTok := lex.Peek()
	if identTok.Type != gslIdent || identTok.Value != "scene" {
		lex.LoadCheckpoint(checkpoint)
		return nil
	}
	lex.Next() // consume 'scene'

	parenTok := lex.Peek()
	if parenTok.Value != "(" {
		return &participle.UnexpectedTokenError{
			Unexpected: *parenTok,
			Expect:     "(",
		}
	}
	lex.Next() // consume '('

	sizeTok := lex.Peek()
	if sizeTok.Type != gslInt && sizeTok.Type != gslFloat {
		return &participle.UnexpectedTokenError{
			Unexpected: *sizeTok,
			Expect:     "number",
		}
	}
	lex.Next() // consume size
	size, err := strconv.ParseFloat(sizeTok.Value, 64)
	if err != nil {
		return err
	}
	if size <= 0 {
		return &participle.ParseError{
			Pos: sizeTok.Pos,
			Msg: fmt.Sprintf("scene size must be greater than zero, got %v", size),
		}
	}

	var grid *float64
	nextTok := lex.Peek()
	if nextTok.Value == "," {
		lex.Next() // consume ','
		gridTok := lex.Peek()
		if gridTok.Type != gslInt && gridTok.Type != gslFloat {
			return &participle.UnexpectedTokenError{
				Unexpected: *gridTok,
				Expect:     "number",
			}
		}
		lex.Next() // consume grid
		g, err := strconv.ParseFloat(gridTok.Value, 64)
		if err != nil {
			return err
		}
		if g <= 0 {
			return &participle.ParseError{
				Pos: gridTok.Pos,
				Msg: fmt.Sprintf("scene grid spacing must be greater than zero, got %v", g),
			}
		}
		grid = &g
		nextTok = lex.Peek()
	}

	if nextTok.Value != ")" {
		return &participle.UnexpectedTokenError{
			Unexpected: *nextTok,
			Expect:     ")",
		}
	}
	closeParen := lex.Next() // consume ')'

	if lex.Peek().Value == ";" {
		return &participle.ParseError{
			Pos: lex.Peek().Pos,
			Msg: "@scene directive must not end with a semicolon",
		}
	}

	d.Scene = &SceneDirective{
		Pos:    atTok.Pos,
		Size:   size,
		Grid:   grid,
		EndPos: closeParen.Pos,
	}
	return nil
}

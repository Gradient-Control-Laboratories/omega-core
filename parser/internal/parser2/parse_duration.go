package parser2

import (
	"fmt"
	"strconv"

	participle "github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"

	"gradientcontrol.com/omega/core/program"
)

// newDuration constructs and validates a Duration message.
func newDuration(seconds float64, start, end *program.Location) (*program.Duration, error) {
	if seconds <= 0 {
		return nil, fmt.Errorf("duration must be greater than zero, got %v", seconds)
	}
	return &program.Duration{Start: start, End: end, Seconds: seconds}, nil
}

// parseDuration reads @duration(seconds), which stands under @scene at the top
// of the first file: how long the document's animation runs. It is to time
// what @scene is to space, and is written the same way: one literal, and no
// semicolon. It consumes nothing when the next tokens are not '@' 'duration'.
func (d *TopLevel) parseDuration(lex *lexer.PeekingLexer) error {
	if lex.Peek().Value != "@" {
		return nil
	}

	checkpoint := lex.MakeCheckpoint()
	atTok := lex.Next() // consume '@'

	identTok := lex.Peek()
	if identTok.Type != gslIdent || identTok.Value != "duration" {
		lex.LoadCheckpoint(checkpoint)
		return nil
	}
	lex.Next() // consume 'duration'

	if err := expectPunct(lex, "("); err != nil {
		return err
	}

	secondsTok := lex.Peek()
	if secondsTok.Type != gslInt && secondsTok.Type != gslFloat {
		return &participle.UnexpectedTokenError{Unexpected: *secondsTok, Expect: "a number of seconds"}
	}
	lex.Next() // consume the seconds
	seconds, err := strconv.ParseFloat(secondsTok.Value, 64)
	if err != nil {
		return err
	}
	if seconds <= 0 {
		return &participle.ParseError{Pos: secondsTok.Pos,
			Msg: fmt.Sprintf("duration must be greater than zero, got %v", seconds)}
	}

	closeTok := lex.Peek()
	if closeTok.Value != ")" {
		return &participle.UnexpectedTokenError{Unexpected: *closeTok, Expect: `")"`}
	}
	lex.Next() // consume ')'

	if lex.Peek().Value == ";" {
		return &participle.ParseError{Pos: lex.Peek().Pos, Msg: "@duration directive must not end with a semicolon"}
	}

	d.Duration = &DurationDirective{Pos: atTok.Pos, Seconds: seconds, EndPos: closeTok.Pos}
	return nil
}

package parser2

import (
	"fmt"
	"strconv"

	participle "github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"

	"github.com/Gradient-Control-Laboratories/omega-core/program"
)

// defaultCaptionFade is the fade, in seconds, a caption gets when it names
// none. A caption shorter than twice this fades for half its length instead.
const defaultCaptionFade = 0.5

// CaptionDirective is @caption("text", startTime, endTime[, fade]): a line of
// text the viewer shows centred at the bottom of the canvas while the
// animation clock is in [startTime, endTime].
//
// Like @scene it is a directive rather than an annotation: it marks no
// declaration, takes literals only, and ends without a semicolon. It may stand
// among the top-level declarations, or ahead of the imports, under @scene.
// Inside a function body it is an error (buildDeclaration).
type CaptionDirective struct {
	Pos       lexer.Position
	Text      string // unquoted
	StartTime float64
	EndTime   float64
	Fade      float64 // resolved: the default applied, or the one given
	EndPos    lexer.Position
}

func (*CaptionDirective) isDeclaration() {}

// Parse reads one @caption, or answers participle.NextMatch without consuming
// anything when the next tokens are not '@' 'caption'.
func (c *CaptionDirective) Parse(lex *lexer.PeekingLexer) error {
	checkpoint := lex.MakeCheckpoint()
	atTok := lex.Peek()
	if atTok.Value != "@" {
		return participle.NextMatch
	}
	lex.Next() // consume '@'
	if ident := lex.Peek(); ident.Type != gslIdent || ident.Value != "caption" {
		lex.LoadCheckpoint(checkpoint)
		return participle.NextMatch
	}
	lex.Next() // consume 'caption'

	if err := expectPunct(lex, "("); err != nil {
		return err
	}

	textTok := lex.Peek()
	if textTok.Type != gslString {
		return &participle.UnexpectedTokenError{Unexpected: *textTok, Expect: "caption text in double quotes"}
	}
	lex.Next()
	text, err := strconv.Unquote(textTok.Value)
	if err != nil {
		return &participle.ParseError{Pos: textTok.Pos, Msg: fmt.Sprintf("invalid caption text %s: %v", textTok.Value, err)}
	}

	if err := expectPunct(lex, ","); err != nil {
		return err
	}
	_, start, err := captionNumber(lex, "start time")
	if err != nil {
		return err
	}
	if err := expectPunct(lex, ","); err != nil {
		return err
	}
	endTok, end, err := captionNumber(lex, "end time")
	if err != nil {
		return err
	}
	if end <= start {
		return &participle.ParseError{Pos: endTok.Pos,
			Msg: fmt.Sprintf("caption end time (%g) must be greater than its start time (%g)", end, start)}
	}

	half := (end - start) / 2
	fade := min(defaultCaptionFade, half)
	if lex.Peek().Value == "," {
		lex.Next() // consume ','
		fadeTok, f, err := captionNumber(lex, "fade")
		if err != nil {
			return err
		}
		if f > half {
			return &participle.ParseError{Pos: fadeTok.Pos,
				Msg: fmt.Sprintf("caption fade (%g) must be at most half its interval (%g), since it fades in and out within it", f, half)}
		}
		fade = f
	}

	closeTok := lex.Peek()
	if closeTok.Value != ")" {
		return &participle.UnexpectedTokenError{Unexpected: *closeTok, Expect: `")"`}
	}
	lex.Next() // consume ')'

	if lex.Peek().Value == ";" {
		return &participle.ParseError{Pos: lex.Peek().Pos, Msg: "@caption directive must not end with a semicolon"}
	}

	*c = CaptionDirective{
		Pos:       atTok.Pos,
		Text:      text,
		StartTime: start,
		EndTime:   end,
		Fade:      fade,
		EndPos:    closeTok.Pos,
	}
	return nil
}

func expectPunct(lex *lexer.PeekingLexer, value string) error {
	tok := lex.Peek()
	if tok.Value != value {
		return &participle.UnexpectedTokenError{Unexpected: *tok, Expect: strconv.Quote(value)}
	}
	lex.Next()
	return nil
}

// captionNumber reads a number literal. A caption's times are literals, as
// @scene's sizes are, so a name or an expression is refused here.
func captionNumber(lex *lexer.PeekingLexer, what string) (*lexer.Token, float64, error) {
	tok := lex.Peek()
	if tok.Type != gslInt && tok.Type != gslFloat {
		return nil, 0, &participle.ParseError{Pos: tok.Pos,
			Msg: fmt.Sprintf("caption %s must be a number literal in seconds, got %q", what, tok.Value)}
	}
	lex.Next()
	v, err := strconv.ParseFloat(tok.Value, 64)
	if err != nil {
		return nil, 0, &participle.ParseError{Pos: tok.Pos, Msg: err.Error()}
	}
	return tok, v, nil
}

// parseLeadingCaptions reads the @captions that stand ahead of the imports.
func (d *TopLevel) parseLeadingCaptions(lex *lexer.PeekingLexer) error {
	for {
		c := &CaptionDirective{}
		err := c.Parse(lex)
		if err == participle.NextMatch {
			return nil
		}
		if err != nil {
			return err
		}
		d.Captions = append(d.Captions, c)
	}
}

// takeCaptions removes the @captions from a file's top-level declarations and
// returns them after the leading ones, all in source order, as the program
// records them.
func takeCaptions(topLevel *TopLevel) []*program.Caption {
	captions := make([]*program.Caption, 0, len(topLevel.Captions))
	for _, c := range topLevel.Captions {
		captions = append(captions, c.toProgram())
	}
	n := 0
	for _, decl := range topLevel.Declarations {
		if c, ok := decl.(*CaptionDirective); ok {
			captions = append(captions, c.toProgram())
			continue
		}
		topLevel.Declarations[n] = decl
		n++
	}
	topLevel.Declarations = topLevel.Declarations[:n]
	return captions
}

func (c *CaptionDirective) toProgram() *program.Caption {
	return &program.Caption{
		Start:     posToLocation(c.Pos),
		End:       posToLocation(c.EndPos),
		Text:      c.Text,
		StartTime: c.StartTime,
		EndTime:   c.EndTime,
		Fade:      c.Fade,
	}
}

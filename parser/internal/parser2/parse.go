package parser2

import (
	"errors"

	participle "github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"gradientcontrol.com/omega/core/parser/internal/annotation"
)

func (d *TopLevel) Parse(lex *lexer.PeekingLexer) error {
	err := d.parseScene(lex)
	if err != nil {
		return err
	}

	err = d.parseImports(lex)
	if err != nil {
		return err
	}

	err = d.parseAssetImports(lex)
	if err != nil {
		return err
	}

	decls, err := declarationsParser.ParseFromLexer(lex)
	d.Declarations = decls.Declarations

	peek := lex.Peek().Value
	if err != nil {

		var tokenErr *participle.UnexpectedTokenError
		if errors.As(err, &tokenErr) && tokenErr.Unexpected.Value == "return" {
			d.Declarations = filterCompleteDeclarations(d.Declarations)
			return &participle.ParseError{Pos: tokenErr.Position(),
				Msg: "missing \";\" before return"}
		}

		if peek != "return" {
			d.Declarations = filterCompleteDeclarations(d.Declarations)
			if errors.As(err, &tokenErr) {
				if tokenErr.Unexpected.Value == "=" || tokenErr.Unexpected.Value == "}" {
					return errors.Join(err, errors.New("  perhaps forgot \";\" at the end of a previous statement?"))
				}
			}
			return errors.Join(err, errors.New("  expected assignment, named function declaration, or return"))
		}
	}

	if peek != "return" {
		if !lex.Peek().EOF() {
			return &participle.UnexpectedTokenError{
				Unexpected: *lex.Peek(),
				Expect:     "assignment, named function declaration, or return",
			}
		}
		return nil
	}

	lex.Next() // consume 'return'
	returnExprs, err := returnExpressionsParser.ParseFromLexer(lex)
	if err != nil {
		return err
	}
	d.Return = returnExprs.Expressions
	return nil
}

func filterCompleteDeclarations(decls []Declaration) []Declaration {
	n := 0
	for _, decl := range decls {
		if isCompleteDeclaration(decl) {
			decls[n] = decl
			n++
		}
	}
	return decls[:n]
}

func isCompleteDeclaration(decl Declaration) bool {
	switch d := decl.(type) {
	case *NamedFunction:
		return d.EndPos.Line > 0
	case *PropertyAssignment:
		return d.EndPos.Line > 0
	case *Assignment:
		return d.EndPos.Line > 0
	case *AnnotatedDeclaration:
		return d.EndPos.Line > 0 && isCompleteDeclaration(d.Declaration)
	default:
		return false
	}
}

func (vn *VarNames) Parse(lex *lexer.PeekingLexer) error {
	names := []VarName{}

	for {
		peek := lex.Peek()
		if peek.Type != gslIdent || peek.Value == "return" {
			return &participle.UnexpectedTokenError{
				Unexpected: *peek,
				Expect:     "identifier",
			}
		}

		next := lex.Next()
		names = append(names, VarName{
			Pos:  next.Pos,
			Name: next.Value,
		})

		peek = lex.Peek()
		if peek.Value == "," {
			lex.Next()
			if lex.Peek().Value == "=" {
				break
			}
			continue
		}

		break
	}

	*vn = VarNames{Names: names}

	return nil
}

func (p *Parameters) Parse(lex *lexer.PeekingLexer) error {

	openCheckpoint := lex.Checkpoint

	// parse '('

	if lex.Peek().Value != "(" {
		lex.Checkpoint = openCheckpoint
		return &participle.UnexpectedTokenError{
			Unexpected: *lex.Peek(),
			Expect:     "(",
		}
	}

	lex.Next()

	if lex.Peek().Value == ")" {
		return verifyOpenBraceAfterCloseParen(lex, openCheckpoint)
	}

	// parse parameter declarations
	for {
		peek := lex.Peek()

		var annot *AnnotationDirective
		if peek.Value == "@" {
			atTok := lex.Next()
			nameTok := lex.Peek()
			if nameTok.Type != gslIdent {
				lex.Checkpoint = openCheckpoint
				return &participle.UnexpectedTokenError{
					Unexpected: *nameTok,
					Expect:     "annotation identifier",
				}
			}
			annotName := lex.Next().Value
			if err := annotation.ValidateTarget(annotName, annotation.TargetParameter); err != nil {
				return &participle.ParseError{
					Pos: atTok.Pos,
					Msg: err.Error(),
				}
			}

			var args []*AnnotationArg
			endPos := lex.Peek().Pos
			if lex.Peek().Value == "(" {
				lex.Next()
				for {
					if lex.Peek().Value == ")" {
						endPos = lex.Next().Pos
						break
					}
					argExpr, err := expressionParser.ParseFromLexer(lex)
					peek := lex.Peek()
					if err != nil && peek.Value != "," && peek.Value != ")" {
						return err
					}
					args = append(args, &AnnotationArg{
						Pos:    argExpr.Pos,
						Expr:   argExpr,
						EndPos: lex.Peek().Pos,
					})
					if lex.Peek().Value == "," {
						lex.Next()
						if lex.Peek().Value == ")" {
							endPos = lex.Next().Pos
							break
						}
						continue
					} else if lex.Peek().Value == ")" {
						endPos = lex.Next().Pos
						break
					} else {
						return &participle.UnexpectedTokenError{
							Unexpected: *lex.Peek(),
							Expect:     "\",\" or \")\"",
						}
					}
				}
			}
			annot = &AnnotationDirective{
				Pos:    atTok.Pos,
				Name:   annotName,
				Args:   args,
				EndPos: endPos,
			}
			peek = lex.Peek()
		}

		if peek.Type != gslIdent {
			break
		}
		next := lex.Next()
		p.Params = append(p.Params, ParameterDecl{
			Pos:        next.Pos,
			Annotation: annot,
			Name:       next.Value,
		})

		peek = lex.Peek()

		switch peek.Value {

		case ",":
			lex.Next()
			continue

		case ")":
			return verifyOpenBraceAfterCloseParen(lex, openCheckpoint)

		default:
			lex.Checkpoint = openCheckpoint
			return &participle.UnexpectedTokenError{
				Unexpected: *peek,
				Expect:     "expected \",\" or \")\"",
			}
		}
	}

	first := lex.Peek()

	if first.Value == ")" {
		return verifyOpenBraceAfterCloseParen(lex, openCheckpoint)
	}

	if first.Value == "." {
		dotCheckpoint := lex.Checkpoint
		dotCount := 0
		for {
			peek := lex.Peek()
			if peek.Value != "." {
				break
			}
			lex.Next()
			dotCount++
		}
		if dotCount == 3 {
			lex.Checkpoint = dotCheckpoint
			return &participle.ParseError{
				Pos: first.Pos,
				Msg: "varargs (...) are not supported",
			}
		}
		lex.Checkpoint = dotCheckpoint
	}

	lex.Checkpoint = openCheckpoint
	return &participle.UnexpectedTokenError{
		Unexpected: *first,
		Expect:     "identifier or \")\"",
	}
}

func verifyOpenBraceAfterCloseParen(lex *lexer.PeekingLexer, openCheckpoint lexer.Checkpoint) error {

	lex.Next() // Assume ')'

	next := lex.Next()

	if next.Value != "{" {
		lex.Checkpoint = openCheckpoint
		return &participle.UnexpectedTokenError{
			Unexpected: *next,
			Expect:     "{",
		}
	}

	return nil
}

// Mark the current position without consuming any tokens
func (m *MarkPos) Parse(lex *lexer.PeekingLexer) error {
	m.Pos = lex.Peek().Pos
	return nil
}

func (r *ReturnExpressions) Parse(lex *lexer.PeekingLexer) error {
	if lex.Peek().Value == ";" {
		return &participle.ParseError{Pos: lex.Peek().Pos,
			Msg: "expected at least one return element"}
	}
	for {
		expr, err := expressionParser.ParseFromLexer(lex)
		peek := lex.Peek()
		if err != nil && peek.Value != "," && peek.Value != ";" {
			return err
		}
		r.Expressions = append(r.Expressions, expr)

		if peek.Value == "," {
			lex.Next()
			if lex.Peek().Value == ";" {
				lex.Next()
				break
			}
			continue
		} else if peek.Value == ";" {
			lex.Next()
			break
		} else {
			return &participle.UnexpectedTokenError{
				Unexpected: *peek,
				Expect:     "\",\" or \";\"",
			}
		}
	}
	return nil
}

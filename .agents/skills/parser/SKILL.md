---
name: parser-author
description: Expert in building parsers using the github.com/alecthomas/participle/v2 parser framework. Use when building or modifying custom lexers, EBNF grammar struct tags, parser error tests, or AST representations.
---

# Skill: Participle Parser Author

## Description

The `github.com/alecthomas/participle/v2` framework constructs parsers directly from Go structs and
struct tags using an EBNF-like grammar to parse text into an Abstract Syntax Tree (AST).

---

## Core Concepts

- **AST Structs:** Go structs reflected upon by Participle to build the grammar and populate fields.
- **Lexer:** Separates lexing from parsing via custom regex-based lexers (`lexer.MustSimple`) or
  stateful lexers.
- **Struct Tags:** `parser:"..."` struct tags define how grammar productions map to struct fields.
- **Capture (`@`):** Captures a matching token or sub-expression into the struct field.
- **Recursive Parsing (`@@`):** Recursively parses the type of the struct field.

---

## Syntax & Grammar

### Declarations (Struct Tags)

- **Literal Matches:** `"string"` or `'string'` (e.g., `parser:"'('"`).
- **Token Matches / Captures:** `@TokenName` captures the matched token into the field.
- **Capture Sub-expressions:** `@(...)` captures the result of a sub-expression.
- **Delegate / Recursive:** `@@` recursively parses the field's type.

### Operators and Quantifiers

- `?` — Optional (0 or 1 time); valid on pointer or boolean fields.
- `*` — Zero or more times; valid on slice fields.
- `+` — One or more times; valid on slice fields.
- `|` — Alternation; matches either left or right side.
- `(...)` — Groups tokens together for quantification or capture.
- `!(...)` / `~(...)` — Negated token match.

---

## Best Practices & Guidelines

### Positional Information

Embed `lexer.Position` in AST structs so Participle populates file, line, and column automatically:
```go
type MyNode struct {
	Pos  lexer.Position
	Name string `parser:"@Ident"`
}
```

### Native Boolean Captures

Participle evaluates `bool` properties marked `?` into `true` automatically when present:
```go
type Factor struct {
	Not      bool  `parser:"@'!'?"`
	Negative bool  `parser:"@'-'?"`
	Value    Value `parser:"@@"`
}
```

### Union Types (Interfaces)

When an AST node branches into mutually exclusive possibilities, use `participle.Union` backing a
common Go interface rather than monolithic structs filled with `nil` pointers:
```go
type Node interface{ Node() }
type A struct{ A string `parser:"'A' @Ident"` }
type B struct{ B string `parser:"'B' @Ident"` }
func (A) Node() {}
func (B) Node() {}
```

### Custom Unmarshaling & Lookahead

Implement `Capture(values []string) error` or `Parse(lex *lexer.PeekingLexer) error` when custom
parsing logic is needed, and factor out shared prefixes across branches to avoid backtracking.

---

## Parser Safety Rule

- Any changes to the parser MUST pass all parser and assembler tests (`go test ./parser/... ./assembler/...`).
- Do not recompile the Fidget/Manifold native libraries if no changes were made to Rust/C++ sources.

---

## Testing Parser Errors

In `parser/internal/parser2/parser_test.go`, there are two ways to test parser failure conditions:

1. **Whole-File Negative Tests**: Prefix the first line with `// ERROR: <optional_expected_message>`.
2. **Inline Error Assertions**: Encode localized mutation tests in valid `.gsl` files via
   `// <bad code> // ERROR: <id> <optional_expected_message>`. Omitting the message auto-records
   the exact parser error string on `go test`.


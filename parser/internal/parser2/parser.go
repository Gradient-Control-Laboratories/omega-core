package parser2

import (
	"errors"
	"strings"

	participle "github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"

	"gradientcontrol.com/omega/core/internal/scope"
	api "gradientcontrol.com/omega/core/parser/internal/api"
	"gradientcontrol.com/omega/core/program"
	"gradientcontrol.com/omega/core/program/scene"
)

var gslLexer = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "Comment", Pattern: `//.*|/\*[\s\S]*?\*/`},
	{Name: "Float", Pattern: `(?:\d+\.\d*|\.\d+)(?:[eE][+-]?\d+)?|\d+[eE][+-]?\d+`},
	{Name: "Int", Pattern: `\d+`},
	{Name: "Ident", Pattern: `[a-zA-Z_]\w*`},
	{Name: "Dot", Pattern: `\.`},
	{Name: "String", Pattern: `"(?:\\.|[^"])*"`},
	{Name: "Punct", Pattern: `[-+*/%,;:=><!&|?(){}\[\]@]`},
	{Name: "Whitespace", Pattern: `[ \t\n\r]+`},
})
var gslIdent = gslLexer.Symbols()["Ident"]
var gslInt = gslLexer.Symbols()["Int"]
var gslFloat = gslLexer.Symbols()["Float"]
var gslDot = gslLexer.Symbols()["Dot"]
var gslString = gslLexer.Symbols()["String"]

var parser = participle.MustBuild[TopLevel](options...)
var importsParser = participle.MustBuild[Imports](options...)
var declarationsParser = participle.MustBuild[Declarations](options...)
var returnExpressionsParser = participle.MustBuild[ReturnExpressions](options...)
var expressionParser = participle.MustBuild[Expression](options...)
var options = []participle.Option{
	participle.Lexer(gslLexer),
	participle.Elide("Whitespace", "Comment"),
	// Lookahead 3 for distinguishing between qualified function calls and qualified references
	participle.UseLookahead(3),
	participle.Union[Declaration](&NamedFunction{}, &PropertyAssignment{}, &Assignment{}, &AnnotatedDeclaration{}),
	// Order is important here. ImportRef is first so the `importStep`/`importMesh`
	// keywords are matched before they could be consumed as a plain Reference.
	participle.Union[Value](&ImportRef{}, &Number{}, &FunctionCall{}, &QualifiedRef{}, &Reference{}, &AnonCall{}, &SubExpression{}, &Collection{}, &String{}, &AnnotatedExpr{}),
}

// Parse parses all source files in a package into a program.Package.
//
// Invariants:
//   - The returned *api.ParseResult and its Package are never nil.
//   - Package.Files always contains at least one non-nil *program.PackageFile
//     (len(res.Package.Files) >= 1 and res.Package.Files[0] != nil).
//   - If a syntax or semantic error occurs, res.Errors contains the error and
//     res.Package preserves all files, scene metadata, imports, and declarations
//     completed prior to the error.
func Parse(sources []string) (*api.ParseResult, error) {
	pkg := &program.Package{
		VersionMajor: program.VersionMajor,
		VersionMinor: program.VersionMinor,
		Files:        make([]*program.PackageFile, 0, max(1, len(sources))),
	}
	res := &api.ParseResult{
		Package: pkg,
	}
	if len(sources) == 0 {
		pkgFile := &program.PackageFile{}
		pkg.Files = append(pkg.Files, pkgFile)
		return recordFileError(res, pkgFile, "", errors.New("no sources provided to parse"))
	}

	prog := &program.Program{
		Primary: pkg,
	}
	progScope, err := scope.Init(prog)
	if err != nil {
		pkgFile := &program.PackageFile{}
		pkg.Files = append(pkg.Files, pkgFile)
		return recordFileError(res, pkgFile, "", err)
	}
	pkgScope := progScope.Primary

	for i, source := range sources {
		pkgFile := &program.PackageFile{}
		pkg.Files = append(pkg.Files, pkgFile)
		pkgFileScope := pkgScope.NewPkgFileScope(pkgFile)

		topLevel, perr := parser.ParseString("", source)
		if topLevel == nil {
			if perr == nil {
				perr = program.InternalErr(errors.New("nil AST with no parse error"))
			}
			return recordFileError(res, pkgFile, source, perr)
		}

		if i == 0 {
			if topLevel.Scene != nil {
				sc, err := scene.New(topLevel.Scene.Size, topLevel.Scene.Grid,
					posToLocation(topLevel.Scene.Pos), posToLocation(topLevel.Scene.EndPos))
				if err != nil {
					return recordFileError(res, pkgFile, source, participle.Errorf(topLevel.Scene.Pos, "%s", err.Error()))
				}
				pkg.Scene = sc
			} else {
				pkg.Scene = fromLegacyComments(source)
			}
		} else if topLevel.Scene != nil {
			return recordFileError(res, pkgFile, source, participle.Errorf(topLevel.Scene.Pos, "@scene directive is only allowed in the first source file"))
		}

		if berr := buildPkgFile(pkgFileScope, topLevel); berr != nil {
			return recordFileError(res, pkgFile, source, berr)
		}

		if perr != nil {
			return recordFileError(res, pkgFile, source, perr)
		}

		if i < len(sources)-1 && len(topLevel.Return) > 0 {
			return recordFileError(res, pkgFile, source, participle.Errorf(topLevel.Return[0].Pos, "return statement is only allowed in the last source file"))
		}
	}

	return res, nil
}

func recordFileError(res *api.ParseResult, pkgFile *program.PackageFile, source string, err error) (*api.ParseResult, error) {
	pkgFile.Source = source
	res.Errors = append(res.Errors, ParseError(source, err))
	return res, nil
}

func ParseError(source string, err error) program.Error {
	pErr, ok := err.(participle.Error)
	if ok {

		errPos := pErr.Position()
		lineNum := errPos.Line
		column := errPos.Column

		lines := strings.Split(source, "\n")
		if 1 <= lineNum && lineNum <= len(lines) {
			line := lines[lineNum-1]
			msg := pErr.Message()

			// GSL's ?: branches are LogicOr, not Expression — so nested
			// ternaries are illegal inline, and the bare participle error
			// ("unexpected token ?") doesn't tell the model how to fix it.
			// When the failing token is a ?, append a decomposition hint so
			// the Fix-error follow-up turn has actionable guidance.
			if column >= 1 && column <= len(line) && line[column-1] == '?' {
				msg += " — nested ternaries cannot be inlined; assign the inner conditional to a variable first, e.g. `inner = (b > 0) ? y : z; result = (a > 0) ? x : inner;`"
			}

			return program.Error{
				Line:    line,
				LineNum: lineNum,
				Column:  column,
				Length:  1,
				Message: msg,
			}
		}
	}
	return program.Error{
		LineNum: 1,
		Column:  1,
		Message: err.Error(),
	}
}

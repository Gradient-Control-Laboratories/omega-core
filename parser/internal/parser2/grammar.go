package parser2

import (
	participle "github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type TopLevel struct {
	Scene        *SceneDirective
	Imports      []*Import     // `parser:"@@*"`
	MeshImports  []*MeshImport // `parser:"@@*"`
	StepImports  []*StepImport // `parser:"@@*"`
	WarpImports  []*WarpImport // `parser:"@@*"`
	Declarations []Declaration // `parser:"@@*"`
	Return       []*Expression // `parser:"( 'return' @@ ( ',' @@ )* ','? ';' )?"`
}

type SceneDirective struct {
	Pos    lexer.Position
	Size   float64
	Grid   *float64
	EndPos lexer.Position
}

type Imports struct {
	Imports []*Import `parser:"@@*"`
}

type Import struct {
	Pos    lexer.Position
	Prefix string `parser:" 'import' ( @Ident | @Dot )? "`
	Path   string `parser:" @String ';' "`
	EndPos lexer.Position
}

type MeshImport struct {
	Pos    lexer.Position
	Index  int    `parser:" 'importMesh' @Int "`
	Path   string `parser:" @String ';' "`
	EndPos lexer.Position
}

type StepImport struct {
	Pos    lexer.Position
	Index  int    `parser:" 'importStep' @Int "`
	Path   string `parser:" @String ';' "`
	EndPos lexer.Position
}

type WarpImport struct {
	Pos    lexer.Position
	Index  int    `parser:" 'importWarp' @Int "`
	Path   string `parser:" @String ';' "`
	EndPos lexer.Position
}

type Declarations struct {
	Declarations []Declaration `parser:"@@*"`
}

type ReturnExpressions struct {
	Expressions []*Expression
}

type Declaration interface{ isDeclaration() }

func (*NamedFunction) isDeclaration()        {}
func (*PropertyAssignment) isDeclaration()   {}
func (*Assignment) isDeclaration()           {}
func (*AnnotatedDeclaration) isDeclaration() {}

type AnnotationArg struct {
	Pos    lexer.Position
	Expr   *Expression `parser:"@@"`
	EndPos lexer.Position
}

type AnnotationDirective struct {
	Pos    lexer.Position
	Name   string           `parser:"'@' @( 'range' | 'values' | 'point' | 'transform' | 'spline' | 'material' | 'finish' | 'color' | 'direction' | 'frame' | 'surface' )"`
	Args   []*AnnotationArg `parser:"[ '(' ( @@ ( ',' @@ )* ','? )? ')' ]"`
	EndPos lexer.Position
}

type AnnotatedDeclaration struct {
	Pos         lexer.Position
	Annotation  *AnnotationDirective `parser:"@@"`
	Declaration Declaration          `parser:"@@"`
	EndPos      lexer.Position
}

type FunctionName string

func (f *FunctionName) Parse(lex *lexer.PeekingLexer) error {
	peek := lex.Peek()
	if peek.Type != gslIdent {
		return &participle.UnexpectedTokenError{
			Unexpected: *peek,
			Expect:     "identifier",
		}
	}
	if peek.Value == "return" || peek.Value == "audio" || peek.Value == "import" {
		return &participle.UnexpectedTokenError{
			Unexpected: *peek,
			Expect:     "non-keyword identifier",
		}
	}
	lex.Next()
	*f = FunctionName(peek.Value)
	return nil
}

type NamedFunction struct {
	Pos        lexer.Position
	Name       FunctionName  `parser:"@@"`
	Parameters Parameters    `parser:"@@"`
	Body       []Declaration `parser:"@@*"`
	Return     []*Expression `parser:"'return' @@ ( ',' @@ )* ','? ';' '}'"`
	EndPos     lexer.Position
}

type AnonFunction struct {
	Pos        lexer.Position
	Parameters Parameters    `parser:"@@"`
	Body       []Declaration `parser:"@@*"`
	Return     []*Expression `parser:"'return' @@ ( ',' @@ )* ','? ';' '}'"`
	EndPos     lexer.Position
}

type ParameterDecl struct {
	Pos        lexer.Position
	Annotation *AnnotationDirective // @range(min, max), or a control kind: @color, @point, @direction, @transform, @frame
	Name       string
}

type Parameters struct {
	Params []ParameterDecl
}

type Assignment struct {
	Pos        lexer.Position
	Names      VarNames    `parser:"@@"`
	Expression *Expression `parser:" '=' @@ ';'"`
	EndPos     lexer.Position
}

type PropertyAssignment struct {
	Pos        lexer.Position
	VarName    string      `parser:"@Ident '.'"`
	PropName   string      `parser:"@Ident"`
	Expression *Expression `parser:" '=' @@ ';'"`
	EndPos     lexer.Position
}

type VarNames struct {
	Names []VarName // `parser:"@Ident ( ',' @Ident )*"`
}

type VarName struct {
	Pos    lexer.Position
	Name   string // `parser:"@Ident"`
	EndPos lexer.Position
}

type Expression struct {
	Pos     lexer.Position
	Value   *LogicOr `parser:"@@"`
	Ternary *Ternary `parser:"@@?"`
}

type Ternary struct {
	QMark    *MarkPos    `parser:"@@"`
	Question bool        `parser:"@'?'"`
	Positive *Expression `parser:"@@"`

	Mark1  *MarkPos    `parser:"@@"`
	Colon1 bool        `parser:"@':'?"`
	Equal  *Expression `parser:"@@?"`

	Mark2    *MarkPos    `parser:"@@"`
	Colon2   bool        `parser:"@':'?"`
	Negative *Expression `parser:"@@?"`
}

func (e *Expression) StartPos() lexer.Position {
	return e.Pos
}

type LogicOr struct {
	Left  *LogicAnd   `parser:"@@"`
	Right []*LogicAnd `parser:"( '|' '|' @@)*"`
}

type LogicAnd struct {
	Left  *Equality   `parser:"@@"`
	Right []*Equality `parser:"( '&' '&' @@)*"`
}

type Equality struct {
	Left  *Comparison   `parser:"@@"`
	Right *OpComparison `parser:"(@@)?"`
}

type OpComparison struct {
	Comparator Comparator  `parser:"@( '=' '=' | '!' '=' )"`
	Comparison *Comparison `parser:"@@"`
}

type Comparison struct {
	Left  *Terms   `parser:"@@"`
	Right *OpTerms `parser:"(@@)?"`
}

type OpTerms struct {
	Pos        lexer.Position // automatically assigned by parser
	Comparator Comparator     `parser:"@( '>' '=' | '>' | '<' '=' '>' | '<' '=' | '<' )"`
	Terms      *Terms         `parser:"@@"`
}

type Terms struct {
	Left  *Term     `parser:"@@"`
	Right []*OpTerm `parser:"@@*"`
}

type OpTerm struct {
	Operator Operator `parser:"@( '+' | '-' )"`
	Term     *Term    `parser:"@@"`
}

type Term struct {
	Left  *Factor     `parser:"@@"`
	Right []*OpFactor `parser:"@@*"`
}

type OpFactor struct {
	Operator Operator `parser:"@( '*' | '/' | '%' )"`
	Factor   *Factor  `parser:"@@"`
}

type Factor struct {
	Pos      lexer.Position // automatically assigned by parser
	Not      bool           `parser:"@'!'?"`
	Negative bool           `parser:"@'-'?"`
	Value    Value          `parser:"@@"`
	Access   string         `parser:"( '.' @Ident )?"`
	Index    *Expression    `parser:"( '[' @@ ']' )?"`
	EndPos   lexer.Position // automatically assigned by parser
}

type Value interface{ isValue() }

func (*ImportRef) isValue()     {}
func (*Number) isValue()        {}
func (*QualifiedRef) isValue()  {}
func (*FunctionCall) isValue()  {}
func (*Reference) isValue()     {}
func (*AnonCall) isValue()      {}
func (*SubExpression) isValue() {}
func (*Collection) isValue()    {}
func (*String) isValue()        {}
func (*AnnotatedExpr) isValue() {}

type AnnotatedExpr struct {
	Pos        lexer.Position
	Annotation *AnnotationDirective `parser:"@@"`
	Expression *Expression          `parser:"@@"`
	EndPos     lexer.Position
}

// ImportRef is the expression form of an asset import: `importStep "file.stp"`,
// `importMesh "file.stl"`, `importWarp "run.castsim"`. Unlike the numbered
// statement forms (MeshImport/StepImport/WarpImport, which carry an
// author-chosen slot), the expression evaluates to a compiler-assigned slot id,
// so it can be assigned to a variable and returned (e.g.
// `id = importStep "f.stp";`). It is lowered to a FunctionCall named after the
// keyword whose single argument carries the path; a slot-assignment pass
// resolves it to an id — meshes and STEP bodies share one namespace, warps have
// their own, and each is resolved by the pass that owns it.
type ImportRef struct {
	Pos    lexer.Position
	Kind   string `parser:"@( 'importStep' | 'importMesh' | 'importWarp' )"`
	Path   string `parser:"@String"`
	EndPos lexer.Position
}

type Number struct {
	Pos    lexer.Position
	Value  *float64 `parser:"@( Float | Int )"`
	EndPos lexer.Position
}

type QualifiedRef struct {
	Pos    lexer.Position // automatically assigned by parser
	Prefix string         `parser:"@Ident '.'"`
	Name   string         `parser:"@Ident"`
	EndPos lexer.Position // automatically assigned by parser
}

type FunctionCall struct {
	Pos       lexer.Position
	PkgName   string        `parser:"( @Ident '.' )?"`
	Name      string        `parser:"@Ident"`
	Arguments []*Expression `parser:"'(' ( @@ ( ',' @@ )* ','? )? ')'"`
	EndPos    lexer.Position
}

type Reference struct {
	Pos    lexer.Position
	Name   string `parser:"@Ident"`
	EndPos lexer.Position
}

type AnonCall struct {
	Pos          lexer.Position
	AnonFunction *AnonFunction `parser:"@@"`
	Arguments    *Arguments    `parser:"@@?"`
	EndPos       lexer.Position
}

type SubExpression struct {
	Pos        lexer.Position
	Expression *Expression `parser:"'(' @@ ')'"`
	EndPos     lexer.Position
}

type Collection struct {
	Pos         lexer.Position
	Expressions []*Expression `parser:"'[' ( @@ ( ',' @@ )* ','? )? ']'"`
	EndPos      lexer.Position
}

type String struct {
	Pos    lexer.Position
	Value  string `parser:"@String"`
	EndPos lexer.Position
}

type Arguments struct {
	OpenParen   string        `parser:"'('"`
	Expressions []*Expression `parser:"( @@ ( ',' @@ )* ','? )?"`
	CloseParen  string        `parser:"')'"`
}

// Used to mark the current token position for better error reporting
type MarkPos struct {
	Pos lexer.Position
}

type Comparator int

type Operator int

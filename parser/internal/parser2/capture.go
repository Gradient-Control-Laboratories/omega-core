package parser2

import "fmt"

func (c *Comparator) Capture(s []string) error {
	switch len(s) {
	case 1:
		*c = comparatorMap[s[0]]
	case 2:
		*c = comparatorMap[s[0]+s[1]]
	case 3:
		*c = comparatorMap[s[0]+s[1]+s[2]]
	default:
		return fmt.Errorf("invalid comparator: %v", s)
	}
	return nil
}

const (
	CompUnknown Comparator = iota
	CompGT
	CompGE
	CompEQ
	CompNE
	CompLE
	CompLT
	CompCompare
)

var comparatorMap = map[string]Comparator{
	">": CompGT, ">=": CompGE,
	"==": CompEQ, "!=": CompNE,
	"<=": CompLE, "<": CompLT,
	"<=>": CompCompare,
}

// ===================================================

func (o *Operator) Capture(s []string) error {
	*o = operatorMap[s[0]]
	return nil
}

const (
	OpUnknown Operator = iota
	OpMul
	OpDiv
	OpAdd
	OpSub
	OpMod
)

var operatorMap = map[string]Operator{
	"+": OpAdd, "-": OpSub,
	"*": OpMul, "/": OpDiv, "%": OpMod,
}

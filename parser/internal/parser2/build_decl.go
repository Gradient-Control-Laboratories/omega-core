package parser2

import (
	"errors"
	"fmt"

	"gradientcontrol.com/omega/core/internal/scope"
)

func buildDeclarations(buildScope scope.BuilderScope, declaration []Declaration) error {
	var err error
	for _, decl := range declaration {
		err = buildDeclaration(buildScope, decl)
		if err != nil {
			return err
		}
	}
	return nil
}

func buildDeclaration(buildScope scope.BuilderScope, declNode Declaration) error {
	if declNode == nil {
		return errors.New("missing declaration")
	}

	switch decl := declNode.(type) {
	case *AnnotatedDeclaration:
		return buildAnnotatedDeclaration(buildScope, decl)
	case *Assignment:
		return buildAssignment(buildScope, decl, nil)
	case *PropertyAssignment:
		return buildPropertyAssignment(buildScope, decl)
	case *NamedFunction:
		return buildFunction(buildScope, decl)
	default:
		return fmt.Errorf("unhandled declaration type: %T", declNode)
	}
}

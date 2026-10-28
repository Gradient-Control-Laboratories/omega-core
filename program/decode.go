package program

import (
	"fmt"

	"google.golang.org/protobuf/proto"
)

// DecodePackage decodes program.Package bytes, falling back to legacy model.Package decoding and conversion.
func DecodePackage(data []byte) (*Package, error) {
	pkg := &Package{}
	err := proto.Unmarshal(data, pkg)
	if err != nil || len(pkg.Files) == 0 {
		return nil, fmt.Errorf("failed to decode package bytes: %v", err)
	}
	return pkg, nil
}

// DecodeProgram decodes program.Program bytes.
func DecodeProgram(data []byte) (*Program, error) {
	prog := &Program{}
	err := proto.Unmarshal(data, prog)
	if err != nil {
		return nil, fmt.Errorf("failed to decode program bytes: %v", err)
	}
	return prog, nil
}

package compiler

import (
	"errors"
	"fmt"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/assembler/variation"
	"gradientcontrol.com/omega/core/backend/tape"
	"gradientcontrol.com/omega/core/internal/compile"
	"gradientcontrol.com/omega/core/internal/meshimport"
)

// ErrTapeRequest is wrapped by the errors Tape returns for a document it will not make a tape of, as opposed to a
// document that does not compile.
var ErrTapeRequest = errors.New("bad tape request")

// TapeRequest describes a server-side tape: the viewer's document as a self-contained evaluation tape, for the
// viewer's Fidget draw device, which runs fidget's GPU pipeline on it in the page.
type TapeRequest struct {
	ID       string `json:"id"`
	ClientID string `json:"clientId"`
	Source   string `json:"source"`
	// Appearance puts each return's colour and material fields in the tape, which the Fidget device lights its frame
	// with.
	Appearance bool `json:"appearance"`
}

// Tape assembles the request's source and compiles every return of it, as the viewer draws them, into a tape. A
// document that reads an imported mesh (bvhSdf, vdbSdf) has no tape here: fidget's GPU path has no encoding for the
// host callback such a leaf needs.
func Tape(req TapeRequest, sourcePaths []string) ([]byte, error) {
	asmRes, err := assembler.Assemble([]string{req.Source}, compile.AssemblerOptions(sourcePaths, nil)...)
	if err != nil {
		return nil, err
	}
	prog := asmRes.Program
	if len(asmRes.AudioReturns) > 0 {
		prog = variation.PrunePrimary(prog)
	}
	if meshimport.References(prog) {
		return nil, fmt.Errorf("%w: the Fidget device does not draw a document that reads an imported mesh (bvhSdf, vdbSdf)", ErrTapeRequest)
	}
	return tape.Compile(prog, tape.WithAppearance(req.Appearance))
}

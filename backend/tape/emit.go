package tape

import (
	"fmt"

	"gradientcontrol.com/omega/core/internal/fidget"
	"gradientcontrol.com/omega/core/program"
)

// Option configures bytecode tape compilation.
type Option func(*options)

type options struct {
	progressFn ProgressFunc
}

// ProgressFunc is invoked during tape compilation milestones.
type ProgressFunc func(stage string, completed, total uint32)

// WithProgress supplies a progress callback to report tape compilation progress.
func WithProgress(fn ProgressFunc) Option {
	return func(o *options) {
		o.progressFn = fn
	}
}

// Compile translates an assembled program into a bytecode evaluation tape.
func Compile(prog *program.Program, opts ...Option) ([]byte, error) {
	if prog == nil {
		return nil, fmt.Errorf("cannot compile nil program into tape")
	}

	var o options
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}

	if o.progressFn != nil {
		o.progressFn("Compiling bytecode tape", 0, 1)
	}

	tapeBytes, _, err := fidget.GenerateTape(prog)
	if err != nil {
		return nil, fmt.Errorf("tape generation failed: %w", err)
	}

	if o.progressFn != nil {
		o.progressFn("Compiling bytecode tape", 1, 1)
	}

	return tapeBytes, nil
}

package assembler

import (
	"fmt"

	"gradientcontrol.com/omega/core/assembler/source"
	lib "gradientcontrol.com/omega/core/internal/assembler"
)

// Result contains the assembled program AST and compilation products.
type Result = lib.Result

// Assemble collects packages directly and indirectly referenced by primarySources,
// executes compiler passes, and assembles them into an optimized Program AST.
func Assemble(primarySources []string, opts ...Option) (*Result, error) {
	if len(primarySources) == 0 {
		return nil, fmt.Errorf("no primary sources provided")
	}

	var cfg config
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	srcMgr := cfg.srcMgr
	if srcMgr == nil {
		if cfg.sourcePath != "" {
			srcMgr = source.NewLocalManager(cfg.sourcePath)
		} else {
			srcMgr = source.NewLocalManager(".")
		}
	}

	return lib.Assemble(primarySources, srcMgr, cfg.libOpts...)
}

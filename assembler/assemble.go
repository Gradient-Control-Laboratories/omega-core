package assembler

import (
	"fmt"

	lib "gradientcontrol.com/omega/core/internal/assembler"
	internalsource "gradientcontrol.com/omega/core/internal/source"
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

	var srcMgr internalsource.Manager = internalsource.NewStdLibManager()
	if len(cfg.managers) > 0 {
		managers := make([]internalsource.Manager, 0, len(cfg.managers)+1)
		managers = append(managers, cfg.managers...)
		managers = append(managers, srcMgr)
		srcMgr = internalsource.NewMultiSourceManager(managers...)
	}

	return lib.Assemble(primarySources, srcMgr, cfg.libOpts...)
}

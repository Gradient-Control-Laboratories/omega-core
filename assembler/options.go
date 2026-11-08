package assembler

import (
	"gradientcontrol.com/omega/core/assembler/source"
	lib "gradientcontrol.com/omega/core/internal/assembler"
	internalsource "gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/program"
)

// Option configures program assembly.
type Option func(*config)

type config struct {
	managers []internalsource.Manager
	libOpts  []lib.Option
}

// Parser supplies a custom package parser (such as parser.CachingParser.ParsePackage)
// to use when parsing primary and imported packages during assembly.
func Parser(parsePkg program.ParsePackageFunc) Option {
	return func(c *config) {
		c.libOpts = append(c.libOpts, lib.Parser(parsePkg))
	}
}

// SourcePaths adds local directory root paths used to resolve imported packages and assets.
func SourcePaths(paths ...string) Option {
	return func(c *config) {
		if len(paths) > 0 {
			c.managers = append(c.managers, internalsource.NewLocalManager(paths...))
		}
	}
}

// SourceManager adds a custom source manager used to resolve imported packages and assets.
func SourceManager(mgr source.Manager) Option {
	return func(c *config) {
		if mgr != nil {
			c.managers = append(c.managers, mgr)
		}
	}
}

// DisableMeshSlotFolding disables the automatic mesh slot assignment and folding pass.
func DisableMeshSlotFolding() Option {
	return func(c *config) {
		c.libOpts = append(c.libOpts, lib.DisableMeshSlotFolding())
	}
}

// DisableDeadcodeRemoval disables dead code removal during assembly.
func DisableDeadcodeRemoval() Option {
	return func(c *config) {
		c.libOpts = append(c.libOpts, lib.DisableDeadcodeRemoval())
	}
}

// ExtractUI enables extraction of sliders, UI controls, and scene metadata into the Result.
func ExtractUI() Option {
	return func(c *config) {
		c.libOpts = append(c.libOpts, lib.ExtractUI())
	}
}

// ExtractMeshPlan configures the assembler to partition geometry into a MeshExportPlan
// and analytical SdfSubtrees as Step 8.
func ExtractMeshPlan(idPrefix string) Option {
	return func(c *config) {
		c.libOpts = append(c.libOpts, lib.ExtractMeshPlan(idPrefix))
	}
}

// ProgressFunc is invoked at discrete compilation milestones during program assembly.
type ProgressFunc = lib.ProgressFunc

// WithProgress supplies a progress callback to report compiler pass progress.
func WithProgress(fn ProgressFunc) Option {
	return func(c *config) {
		c.libOpts = append(c.libOpts, lib.WithProgress(fn))
	}
}

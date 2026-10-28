package assembler

import (
	"gradientcontrol.com/omega/core/assembler/source"
	lib "gradientcontrol.com/omega/core/internal/assembler"
)

// Option configures program assembly.
type Option func(*config)

type config struct {
	sourcePath string
	srcMgr     source.Manager
	libOpts    []lib.Option
}

// SourcePath sets the local directory path used to resolve imported packages.
func SourcePath(path string) Option {
	return func(c *config) {
		c.sourcePath = path
	}
}

// SourceManager sets a custom source manager used to resolve imported packages and assets.
func SourceManager(mgr source.Manager) Option {
	return func(c *config) {
		c.srcMgr = mgr
	}
}

// PrimaryPackage supplies a primary package path assigned to the assembled program's primary package.
func PrimaryPackage(pkg string) Option {
	return func(c *config) {
		c.libOpts = append(c.libOpts, lib.PrimaryPackage(pkg))
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

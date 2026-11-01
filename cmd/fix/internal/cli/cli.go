package cli

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gradientcontrol.com/omega/core/cmd/fix/internal/config"
	"gradientcontrol.com/omega/core/cmd/fix/internal/fix"
)

// ErrDiffFound is returned in -diff mode when one or more files have pending fixes.
var ErrDiffFound = errors.New("diffs found")

// PrintUsage writes the brief usage message shown when fix is invoked with no arguments.
func PrintUsage(w io.Writer) {
	fmt.Fprintf(w, `fix is a tool for refactoring and updating GSL programs.

Usage of fix:
	fix [flags] <path ...>	# execute fixers on specified .gsl files or directories
	fix help              	# general help, including listing fixers and flags
	fix help <name>       	# help on specific fixer and its flags
`)
}

// PrintHelp writes the full help documentation including registered fixers and core flags.
func PrintHelp(w io.Writer) {
	PrintUsage(w)
	fmt.Fprintf(w, "\nRegistered fixers:\n\n")
	for _, f := range fix.All() {
		fmt.Fprintf(w, "    %-12s %s\n", f.Name, f.Doc)
	}
	fmt.Fprintf(w, `
By default all fixers are run.
To select specific fixers, use the -NAME flag for each one,
or -NAME=false to run all fixers not explicitly disabled.

Core flags:

  -diff
    	instead of applying each fix, print the patch as a unified diff;
    	exit with a non-zero status if the diff is not empty
`)
	for _, f := range fix.All() {
		fmt.Fprintf(w, "  -%s\n    \tenable %s fixer\n", f.Name, f.Name)
	}
	fmt.Fprintf(w, "\nTo see details and examples of a specific fixer, run 'fix help <name>'.\n")
}

// Run executes the fix CLI workflow for the given arguments and output writer.
func Run(args []string, stdout io.Writer) error {
	cfg, err := config.ParseFlags(args)
	if err != nil {
		return err
	}

	switch cfg.Mode {
	case config.ModeUsage:
		PrintUsage(stdout)
		return nil

	case config.ModeHelp:
		PrintHelp(stdout)
		return nil

	case config.ModeFixerHelp:
		// Invariant: config.ParseFlags validates HelpTopic against fix.ByName before setting ModeFixerHelp.
		f, _ := fix.ByName(cfg.HelpTopic)
		fmt.Fprintf(stdout, "%s: %s\n\n%s", f.Name, f.Doc, f.Help)
		return nil
	}

	files, err := collectGSLFiles(cfg.Paths)
	if err != nil {
		return err
	}

	anyDiff := false
	for _, filePath := range files {
		changed, fileErr := processFile(filePath, cfg.ActiveFixers, cfg.Diff, stdout)
		if fileErr != nil {
			return fileErr
		}
		if changed && cfg.Diff {
			anyDiff = true
		}
	}

	if anyDiff {
		return ErrDiffFound
	}
	return nil
}

func collectGSLFiles(paths []string) ([]string, error) {
	var files []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, fmt.Errorf("stat %s: %w", p, err)
		}
		if !info.IsDir() {
			files = append(files, p)
			continue
		}

		walkErr := filepath.WalkDir(p, func(curPath string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				name := d.Name()
				if curPath != p && strings.HasPrefix(name, ".") {
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(strings.ToLower(d.Name()), ".gsl") {
				files = append(files, curPath)
			}
			return nil
		})
		if walkErr != nil {
			return nil, walkErr
		}
	}
	return files, nil
}

func processFile(filePath string, activeFixers []fix.Fixer, diffMode bool, stdout io.Writer) (bool, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return false, err
	}

	rawBytes, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("reading %s: %w", filePath, err)
	}

	orig := string(rawBytes)
	updated, changed, err := fix.Apply(orig, activeFixers)
	if err != nil {
		return false, fmt.Errorf("fixing %s: %w", filePath, err)
	}
	if !changed || updated == orig {
		return false, nil
	}

	if diffMode {
		fmt.Fprint(stdout, formatUnifiedDiff(filePath, orig, updated))
		return true, nil
	}

	if err := os.WriteFile(filePath, []byte(updated), info.Mode().Perm()); err != nil {
		return false, fmt.Errorf("writing %s: %w", filePath, err)
	}
	return true, nil
}

func formatUnifiedDiff(filePath, oldText, newText string) string {
	var sb strings.Builder
	slashPath := filepath.ToSlash(filePath)
	fmt.Fprintf(&sb, "--- a/%s\n+++ b/%s\n", slashPath, slashPath)

	oldLines := strings.Split(strings.TrimSuffix(strings.ReplaceAll(oldText, "\r\n", "\n"), "\n"), "\n")
	newLines := strings.Split(strings.TrimSuffix(strings.ReplaceAll(newText, "\r\n", "\n"), "\n"), "\n")

	// Find common prefix and suffix to emit a compact unified hunk
	prefix := 0
	for prefix < len(oldLines) && prefix < len(newLines) && oldLines[prefix] == newLines[prefix] {
		prefix++
	}
	oldEnd := len(oldLines)
	newEnd := len(newLines)
	for oldEnd > prefix && newEnd > prefix && oldLines[oldEnd-1] == newLines[newEnd-1] {
		oldEnd--
		newEnd--
	}

	fmt.Fprintf(&sb, "@@ -%d,%d +%d,%d @@\n",
		prefix+1, oldEnd-prefix,
		prefix+1, newEnd-prefix)

	for i := prefix; i < oldEnd; i++ {
		fmt.Fprintf(&sb, "- %s\n", oldLines[i])
	}
	for i := prefix; i < newEnd; i++ {
		fmt.Fprintf(&sb, "+ %s\n", newLines[i])
	}
	return sb.String()
}

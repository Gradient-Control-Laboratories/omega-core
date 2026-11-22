// Command vendor-fidget copies the Fidget source trees from a checkout of the
// pinned fork revision into internal/fidget/vendor, and reports how each fork
// manifest differs from the vendored one.
//
// The pin is recorded in internal/fidget/vendor/FIDGET_PIN. The tool refuses a
// checkout that is not at the pinned commit or that has local changes, so the
// vendored sources always correspond to a commit anyone can fetch.
//
// Manifests (Cargo.toml) are never rewritten: the vendored ones are
// de-workspaced by hand, so the tool only reports versions, dependencies and
// features that differ, and the differences are reconciled by hand.
//
// Run from the repository root:
//
//	go run ./tool/vendor-fidget -fork <checkout of the fork> [-dry-run]
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// crates lists the vendored Fidget crates.
var crates = []string{
	"fidget",
	"fidget-bytecode",
	"fidget-core",
	"fidget-gui",
	"fidget-jit",
	"fidget-mesh",
	"fidget-raster",
	"fidget-rhai",
	"fidget-shapes",
	"fidget-solver",
	"fidget-wgpu",
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "vendor-fidget:", err)
		os.Exit(1)
	}
}

func run(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("vendor-fidget", flag.ContinueOnError)
	fork := fs.String("fork", "", "path to a checkout of the fork at the pinned commit (required)")
	vendor := fs.String("vendor", "internal/fidget/vendor", "vendored fidget directory, relative to the repository root")
	pinPath := fs.String("pin", "", "pin file (default: FIDGET_PIN inside -vendor)")
	dryRun := fs.Bool("dry-run", false, "report what would change without writing anything")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *fork == "" {
		return fmt.Errorf("-fork is required")
	}
	if *pinPath == "" {
		*pinPath = filepath.Join(*vendor, "FIDGET_PIN")
	}

	f, err := os.Open(*pinPath)
	if err != nil {
		return err
	}
	pin, err := ParsePin(f)
	f.Close()
	if err != nil {
		return fmt.Errorf("%s: %w", *pinPath, err)
	}

	if err := checkForkRevision(*fork, pin); err != nil {
		return err
	}
	fmt.Fprintf(out, "fork checkout is at pinned commit %s\n", pin.Commit[:8])

	if same, err := sameFile(filepath.Join(*fork, "LICENSE.txt"), filepath.Join(*vendor, "LICENSE")); err != nil {
		return err
	} else if !same {
		fmt.Fprintln(out, "NOTE: the fork's LICENSE.txt differs from the vendored LICENSE")
	}

	for _, name := range crates {
		forkDir := filepath.Join(*fork, name)
		vendorDir := filepath.Join(*vendor, name)
		res, err := syncCrate(forkDir, vendorDir, *dryRun)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		verb := "copied"
		if *dryRun {
			verb = "would copy"
		}
		fmt.Fprintf(out, "%-16s %s %d files\n", name, verb, res.Files)
		for _, e := range res.VendoredOnly {
			fmt.Fprintf(out, "%-16s   kept vendored-only entry: %s\n", name, e)
		}
		for _, e := range res.NotVendored {
			fmt.Fprintf(out, "%-16s   not vendored (fork has it): %s\n", name, e)
		}

		lines, err := reportManifest(filepath.Join(vendorDir, "Cargo.toml"), filepath.Join(forkDir, "Cargo.toml"))
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		for _, l := range lines {
			fmt.Fprintf(out, "%-16s   manifest: %s\n", name, l)
		}
	}
	if *dryRun {
		fmt.Fprintln(out, "dry run: nothing was written")
	}
	return nil
}

// checkForkRevision requires the checkout to be at the pinned commit and clean.
func checkForkRevision(fork string, pin Pin) error {
	head, err := git(fork, "rev-parse", "HEAD")
	if err != nil {
		return fmt.Errorf("%s is not a git checkout: %w", fork, err)
	}
	if head != pin.Commit {
		return fmt.Errorf("fork checkout is at %s, but the pin is %s", head, pin.Commit)
	}
	dirty, err := git(fork, "status", "--porcelain")
	if err != nil {
		return err
	}
	if dirty != "" {
		return fmt.Errorf("fork checkout has local changes; vendor only committed, pinned sources")
	}
	return nil
}

func git(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	b, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(string(b)), nil
}

func sameFile(a, b string) (bool, error) {
	x, err := os.ReadFile(a)
	if err != nil {
		return false, err
	}
	y, err := os.ReadFile(b)
	if err != nil {
		return false, err
	}
	return bytes.Equal(x, y), nil
}

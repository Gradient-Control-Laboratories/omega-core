package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"
)

// manifestSummary is the part of a Cargo.toml the report compares: the package
// version and the names listed in the dependency and feature tables.
type manifestSummary struct {
	Version string
	// Names maps a table kind to the names listed in it.
	Names map[string]map[string]bool
}

// comparedTables are the table kinds the report compares. Every
// target-specific dependency table is merged into "target-dependencies".
var comparedTables = []string{
	"dependencies",
	"dev-dependencies",
	"build-dependencies",
	"target-dependencies",
	"features",
}

var (
	tableRE   = regexp.MustCompile(`^\[+\s*([^\]]+?)\s*\]+\s*(#.*)?$`)
	versionRE = regexp.MustCompile(`^version\s*=\s*"([^"]*)"`)
	nameRE    = regexp.MustCompile(`^([A-Za-z0-9_-]+)`)
)

// summarize reads the names in a manifest. The manifests are regular enough
// that a line-based reading is sufficient for a report; it is not a TOML
// parser, and the report only says what to look at.
func summarize(r io.Reader) (manifestSummary, error) {
	s := manifestSummary{Names: map[string]map[string]bool{}}
	table := ""
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if m := tableRE.FindStringSubmatch(line); m != nil {
			table = tableKind(m[1])
			continue
		}
		if table == "package" {
			if m := versionRE.FindStringSubmatch(line); m != nil {
				s.Version = m[1]
			}
			continue
		}
		if !slices.Contains(comparedTables, table) {
			continue
		}
		m := nameRE.FindStringSubmatch(line)
		if m == nil || m[1] == "workspace-hack" {
			continue
		}
		if s.Names[table] == nil {
			s.Names[table] = map[string]bool{}
		}
		s.Names[table][m[1]] = true
	}
	return s, sc.Err()
}

// tableKind normalizes a table header such as `target.'cfg(...)'.dependencies`.
func tableKind(header string) string {
	if strings.HasPrefix(header, "target.") {
		if strings.HasSuffix(header, "dependencies") {
			return "target-dependencies"
		}
		return "target-other"
	}
	return header
}

// compareManifests returns the differences, one line each, between the
// vendored summary and the fork's.
func compareManifests(vendored, fork manifestSummary) []string {
	var lines []string
	if vendored.Version != fork.Version {
		lines = append(lines, fmt.Sprintf("version %s -> %s", vendored.Version, fork.Version))
	}
	for _, t := range comparedTables {
		if only := difference(vendored.Names[t], fork.Names[t]); len(only) > 0 {
			lines = append(lines, fmt.Sprintf("%s only in vendored: %s", t, strings.Join(only, ", ")))
		}
		if only := difference(fork.Names[t], vendored.Names[t]); len(only) > 0 {
			lines = append(lines, fmt.Sprintf("%s only in fork: %s", t, strings.Join(only, ", ")))
		}
	}
	return lines
}

// difference returns the sorted names in a that are not in b.
func difference(a, b map[string]bool) []string {
	var out []string
	for k := range a {
		if !b[k] {
			out = append(out, k)
		}
	}
	slices.Sort(out)
	return out
}

// reportManifest compares two manifest files.
func reportManifest(vendoredPath, forkPath string) ([]string, error) {
	read := func(path string) (manifestSummary, error) {
		f, err := os.Open(path)
		if err != nil {
			return manifestSummary{}, err
		}
		defer f.Close()
		return summarize(f)
	}
	v, err := read(vendoredPath)
	if err != nil {
		return nil, err
	}
	f, err := read(forkPath)
	if err != nil {
		return nil, err
	}
	return compareManifests(v, f), nil
}

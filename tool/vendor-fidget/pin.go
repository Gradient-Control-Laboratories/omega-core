package main

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// Pin is the Fidget revision recorded in the pin file.
type Pin struct {
	// Fork is the repository the sources are vendored from.
	Fork string
	// Branch is the fork branch the commit was taken from.
	Branch string
	// Commit is the full commit of the fork that is vendored.
	Commit string
	// Upstream is the repository the fork was made from.
	Upstream string
	// UpstreamBase is the upstream commit the fork's commits sit on.
	UpstreamBase string
}

var sha40 = regexp.MustCompile(`^[0-9a-f]{40}$`)

// ParsePin reads a pin file: "key: value" lines, with blank lines and lines
// starting with "#" ignored. The keys are fork, branch, commit, upstream and
// upstream-base; commit and upstream-base must be full 40-digit commits.
func ParsePin(r io.Reader) (Pin, error) {
	var p Pin
	sc := bufio.NewScanner(r)
	for n := 1; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			return Pin{}, fmt.Errorf("line %d: expected \"key: value\"", n)
		}
		key, val = strings.TrimSpace(key), strings.TrimSpace(val)
		switch key {
		case "fork":
			p.Fork = val
		case "branch":
			p.Branch = val
		case "commit":
			p.Commit = val
		case "upstream":
			p.Upstream = val
		case "upstream-base":
			p.UpstreamBase = val
		default:
			return Pin{}, fmt.Errorf("line %d: unknown key %q", n, key)
		}
	}
	if err := sc.Err(); err != nil {
		return Pin{}, err
	}
	if !sha40.MatchString(p.Commit) {
		return Pin{}, fmt.Errorf("commit must be a full 40-digit lowercase commit, got %q", p.Commit)
	}
	if !sha40.MatchString(p.UpstreamBase) {
		return Pin{}, fmt.Errorf("upstream-base must be a full 40-digit lowercase commit, got %q", p.UpstreamBase)
	}
	return p, nil
}

package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const sha1x = "0123456789abcdef0123456789abcdef01234567"
const sha2x = "89abcdef0123456789abcdef0123456789abcdef"

func TestParsePin(t *testing.T) {
	good := "# a comment\n\nfork: https://example.com/fidget\nbranch: omega\ncommit: " + sha1x +
		"\nupstream: https://example.com/upstream\nupstream-base: " + sha2x + "\n"
	p, err := ParsePin(strings.NewReader(good))
	if err != nil {
		t.Fatal(err)
	}
	want := Pin{
		Fork:         "https://example.com/fidget",
		Branch:       "omega",
		Commit:       sha1x,
		Upstream:     "https://example.com/upstream",
		UpstreamBase: sha2x,
	}
	if p != want {
		t.Errorf("got %+v, want %+v", p, want)
	}

	for name, in := range map[string]string{
		"short commit":   "commit: abc123\nupstream-base: " + sha2x + "\n",
		"missing commit": "upstream-base: " + sha2x + "\n",
		"missing base":   "commit: " + sha1x + "\n",
		"unknown key":    "commit: " + sha1x + "\nupstream-base: " + sha2x + "\ncomit: x\n",
		"no colon":       "commit " + sha1x + "\n",
	} {
		if _, err := ParsePin(strings.NewReader(in)); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func exists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

func TestSyncCrate(t *testing.T) {
	fork, vend := t.TempDir(), t.TempDir()

	// The fork's crate: sources, a build script, a swap file, and entries that
	// are not vendored.
	write(t, filepath.Join(fork, "Cargo.toml"), "fork manifest\n")
	write(t, filepath.Join(fork, "src", "lib.rs"), "new lib\n")
	write(t, filepath.Join(fork, "src", "sub", "mod.rs"), "new mod\n")
	write(t, filepath.Join(fork, "src", "sub", ".mod.rs.swp"), "swap\n")
	write(t, filepath.Join(fork, "build.rs"), "new build\n")
	write(t, filepath.Join(fork, "benches", "b.rs"), "bench\n")
	write(t, filepath.Join(fork, "tests", "t.rs"), "test\n")

	// The vendored crate: its own manifest and README, and stale sources.
	write(t, filepath.Join(vend, "Cargo.toml"), "vendored manifest\n")
	write(t, filepath.Join(vend, "README.md"), "vendored readme\n")
	write(t, filepath.Join(vend, "src", "lib.rs"), "old lib\n")
	write(t, filepath.Join(vend, "src", "gone.rs"), "removed upstream\n")

	res, err := syncCrate(fork, vend, false)
	if err != nil {
		t.Fatal(err)
	}

	read := func(p string) string {
		b, err := os.ReadFile(filepath.Join(vend, p))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	if got := read("src/lib.rs"); got != "new lib\n" {
		t.Errorf("src/lib.rs = %q", got)
	}
	if got := read("src/sub/mod.rs"); got != "new mod\n" {
		t.Errorf("src/sub/mod.rs = %q", got)
	}
	if got := read("build.rs"); got != "new build\n" {
		t.Errorf("build.rs = %q", got)
	}
	if exists(filepath.Join(vend, "src", "gone.rs")) {
		t.Error("a source the fork no longer has was kept")
	}
	if exists(filepath.Join(vend, "src", "sub", ".mod.rs.swp")) {
		t.Error("a swap file was copied")
	}
	if got := read("Cargo.toml"); got != "vendored manifest\n" {
		t.Errorf("the vendored manifest was touched: %q", got)
	}
	if got := read("README.md"); got != "vendored readme\n" {
		t.Errorf("a vendored-only README was touched: %q", got)
	}
	if exists(filepath.Join(vend, "benches")) || exists(filepath.Join(vend, "tests")) {
		t.Error("benches or tests were copied")
	}

	if res.Files != 3 {
		t.Errorf("Files = %d, want 3 (lib.rs, mod.rs, build.rs)", res.Files)
	}
	if want := []string{"README.md"}; !reflect.DeepEqual(res.VendoredOnly, want) {
		t.Errorf("VendoredOnly = %v, want %v", res.VendoredOnly, want)
	}
	if want := []string{"benches", "tests"}; !reflect.DeepEqual(res.NotVendored, want) {
		t.Errorf("NotVendored = %v, want %v", res.NotVendored, want)
	}
}

func TestSyncCrateDryRun(t *testing.T) {
	fork, vend := t.TempDir(), t.TempDir()
	write(t, filepath.Join(fork, "src", "lib.rs"), "new\n")
	write(t, filepath.Join(vend, "src", "lib.rs"), "old\n")

	res, err := syncCrate(fork, vend, true)
	if err != nil {
		t.Fatal(err)
	}
	if res.Files != 1 {
		t.Errorf("Files = %d, want 1", res.Files)
	}
	b, _ := os.ReadFile(filepath.Join(vend, "src", "lib.rs"))
	if string(b) != "old\n" {
		t.Errorf("a dry run wrote: %q", b)
	}
}

const vendoredManifest = `[package]
name = "fidget-core"
version = "0.5.0"
edition = "2024"

[lib]
name = "fidget_core"

[features]
eval-tests = []

[dependencies]
facet = { version = "0.46", features = ["reflect", "doc"] }
bincode = "1.3"
rayon = "1.10"

[target."cfg(target_arch = \"wasm32\")".dependencies]
getrandom-04 = { version = "0.4", package = "getrandom" }
`

const forkManifest = `[package]
name = "fidget-core"
description = "Core infrastructure for Fidget"
version = "0.5.1"
edition.workspace = true

[dependencies]
facet.workspace = true
rayon.workspace = true
workspace-hack.workspace = true

[target.'cfg(target_arch = "wasm32")'.dependencies]
getrandom-04 = { package = "getrandom", version = "0.4" }
getrandom-03 = { package = "getrandom", version = "0.3", features = ["wasm_js"] }

[dev-dependencies]
facet.workspace = true

[features]
## docs
eval-tests = []
`

func TestCompareManifests(t *testing.T) {
	v, err := summarize(strings.NewReader(vendoredManifest))
	if err != nil {
		t.Fatal(err)
	}
	f, err := summarize(strings.NewReader(forkManifest))
	if err != nil {
		t.Fatal(err)
	}
	got := compareManifests(v, f)
	want := []string{
		"version 0.5.0 -> 0.5.1",
		"dependencies only in vendored: bincode",
		"dev-dependencies only in fork: facet",
		"target-dependencies only in fork: getrandom-03",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got:\n  %s\nwant:\n  %s", strings.Join(got, "\n  "), strings.Join(want, "\n  "))
	}

	// Identical manifests report nothing, whatever their formatting.
	if got := compareManifests(v, v); len(got) != 0 {
		t.Errorf("identical manifests reported %v", got)
	}
}

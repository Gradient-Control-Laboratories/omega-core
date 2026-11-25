package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// syncedEntries are the entries of a crate directory that are copied from the
// fork. Each is replaced whole: files vendored earlier that the fork no longer
// has are removed with it.
var syncedEntries = []string{"src", "build.rs", "README.md"}

// ignoredEntries are crate-directory entries that are neither copied nor
// reported: the manifests are reconciled by hand, and the rest is build output.
var ignoredEntries = []string{"Cargo.toml", "Cargo.lock", "target"}

// syncResult describes one crate's copy.
type syncResult struct {
	// Files is the number of files copied (or that would be, in a dry run).
	Files int
	// VendoredOnly lists synced entries the vendored crate has and the fork
	// does not; they are kept untouched.
	VendoredOnly []string
	// NotVendored lists entries of the fork's crate that are not vendored
	// (benches, tests): they are not copied.
	NotVendored []string
}

// syncCrate replaces the synced entries of vendorDir with those of forkDir.
// In a dry run nothing is written.
func syncCrate(forkDir, vendorDir string, dryRun bool) (syncResult, error) {
	var res syncResult
	for _, e := range syncedEntries {
		src := filepath.Join(forkDir, e)
		dst := filepath.Join(vendorDir, e)
		if _, err := os.Lstat(src); err != nil {
			if !os.IsNotExist(err) {
				return res, err
			}
			if _, err := os.Lstat(dst); err == nil {
				res.VendoredOnly = append(res.VendoredOnly, e)
			}
			continue
		}
		n, err := countFiles(src)
		if err != nil {
			return res, err
		}
		res.Files += n
		if dryRun {
			continue
		}
		if err := os.RemoveAll(dst); err != nil {
			return res, err
		}
		if err := copyTree(src, dst); err != nil {
			return res, err
		}
	}

	entries, err := os.ReadDir(forkDir)
	if err != nil {
		return res, err
	}
	for _, de := range entries {
		name := de.Name()
		if slices.Contains(syncedEntries, name) || slices.Contains(ignoredEntries, name) || skipFile(name) {
			continue
		}
		res.NotVendored = append(res.NotVendored, name)
	}
	return res, nil
}

// skipFile reports whether a file is never vendored: editor swap files, which
// upstream has committed by accident.
func skipFile(name string) bool {
	return strings.HasSuffix(name, ".swp")
}

// countFiles counts the files copyTree would copy from src.
func countFiles(src string) (int, error) {
	n := 0
	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || skipFile(d.Name()) {
			return nil
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("%s: not a regular file", path)
		}
		n++
		return nil
	})
	return n, err
}

// copyTree copies the file or directory tree at src to dst, keeping file
// permissions and leaving out editor swap files. It refuses anything that is
// not a regular file or directory.
func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if skipFile(d.Name()) {
			return nil
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("%s: not a regular file", path)
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		return copyFile(path, target, info.Mode().Perm())
	})
}

func copyFile(src, dst string, perm fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

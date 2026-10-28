package accel

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	parser "gradientcontrol.com/omega/core/parser"
)

// SiblingGSLPath returns the path to the companion .gsl file for a given sidecar YAML file.
// It trims compound extensions such as .<target>.yaml or .yaml and appends .gsl.
func SiblingGSLPath(yamlPath string) string {
	dir := filepath.Dir(yamlPath)
	base := filepath.Base(yamlPath)
	stem := strings.TrimSuffix(base, filepath.Ext(base))
	if subExt := filepath.Ext(stem); subExt != "" {
		stem = strings.TrimSuffix(stem, subExt)
	}
	return filepath.Join(dir, stem+".gsl")
}

// ValidateDescriptor validates that a descriptor aligns with the functions declared in GSL source.
func ValidateDescriptor(desc *PackageDescriptor, gslSource string, expectedPkgName string) error {
	if expectedPkgName != "" && desc.Package != expectedPkgName {
		return fmt.Errorf("package mismatch: descriptor specifies %q but expected %q", desc.Package, expectedPkgName)
	}

	pkg, err := parser.ParsePackage([]string{gslSource})
	if err != nil {
		return fmt.Errorf("failed to parse GSL source for package %q: %w", desc.Package, err)
	}

	// Collect function names and parameter counts from parsed GSL AST
	fnArities := make(map[string]map[int]bool)
	for _, file := range pkg.Files {
		for _, fnType := range file.FnTypes {
			if _, ok := fnArities[fnType.Name]; !ok {
				fnArities[fnType.Name] = make(map[int]bool)
			}
			for _, group := range fnType.Groups {
				fnArities[fnType.Name][int(group.ParamCount)] = true
			}
		}
	}

	for _, accel := range desc.Accelerations {
		arities, ok := fnArities[accel.Name]
		if !ok {
			return fmt.Errorf("package %q: acceleration function %q not found in GSL declarations", desc.Package, accel.Name)
		}

		for _, sig := range accel.Signatures {
			argCount := len(sig.Args)
			if !arities[argCount] {
				return fmt.Errorf("package %q: function %q signature arity %d does not match any declared parameter count in GSL source", desc.Package, accel.Name, argCount)
			}
		}
	}

	return nil
}

// ValidateSidecar validates a sidecar file against its sibling .gsl file on disk
// and returns the parsed descriptor.
func ValidateSidecar(yamlPath string) (*PackageDescriptor, error) {
	desc, err := ParseFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse sidecar %s: %w", yamlPath, err)
	}

	gslPath := SiblingGSLPath(yamlPath)
	gslBytes, err := os.ReadFile(gslPath)
	if err != nil {
		return nil, fmt.Errorf("sibling GSL file not found for %s: %w", yamlPath, err)
	}

	dirName := filepath.Base(filepath.Dir(yamlPath))
	if err := ValidateDescriptor(desc, string(gslBytes), dirName); err != nil {
		return nil, fmt.Errorf("%s: %w", yamlPath, err)
	}

	return desc, nil
}

// ValidateAll scans rootDir for all *.<target>.yaml sidecar files, validates each against its
// sibling .gsl file, and returns the slice of validated package descriptors.
func ValidateAll(rootDir string, target string) ([]*PackageDescriptor, error) {
	suffix := "." + target + ".yaml"
	var descriptors []*PackageDescriptor

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), suffix) {
			desc, err := ValidateSidecar(path)
			if err != nil {
				return err
			}
			descriptors = append(descriptors, desc)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return descriptors, nil
}

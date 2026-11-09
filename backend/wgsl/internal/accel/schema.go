package accel

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Signature defines the parameter type pattern and return type for an accelerated function.
type Signature struct {
	Args   []string `yaml:"args"`
	Return string   `yaml:"return"`
}

// Acceleration defines a hardware acceleration mapping for a function.
type Acceleration struct {
	Name       string      `yaml:"name"`
	Kind       string      `yaml:"kind"`
	Symbol     string      `yaml:"symbol,omitempty"`
	Builtin    string      `yaml:"builtin,omitempty"`
	Precedence string      `yaml:"precedence,omitempty"`
	Unary      bool        `yaml:"unary,omitempty"`
	Signatures []Signature `yaml:"signatures"`
}

// PackageDescriptor represents a parsed <file>.wgsl.yaml sidecar file.
type PackageDescriptor struct {
	Package       string         `yaml:"package"`
	Accelerations []Acceleration `yaml:"accelerations"`
}

// ParseYAML parses and validates a sidecar descriptor from YAML bytes.
func ParseYAML(data []byte) (*PackageDescriptor, error) {
	var desc PackageDescriptor
	if err := yaml.Unmarshal(data, &desc); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}

	if strings.TrimSpace(desc.Package) == "" {
		return nil, fmt.Errorf("missing package name in descriptor")
	}

	for i, a := range desc.Accelerations {
		if strings.TrimSpace(a.Name) == "" {
			return nil, fmt.Errorf("acceleration at index %d: missing acceleration name", i)
		}
		switch a.Kind {
		case "operator":
			if strings.TrimSpace(a.Symbol) == "" {
				return nil, fmt.Errorf("acceleration %q: operator missing symbol", a.Name)
			}
		case "builtin":
			if strings.TrimSpace(a.Builtin) == "" {
				return nil, fmt.Errorf("acceleration %q: builtin missing builtin name", a.Name)
			}
		default:
			return nil, fmt.Errorf("acceleration %q: invalid kind %q (must be 'operator' or 'builtin')", a.Name, a.Kind)
		}
	}

	return &desc, nil
}

// ParseFile reads and parses a sidecar descriptor file from the filesystem.
func ParseFile(path string) (*PackageDescriptor, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}
	return ParseYAML(data)
}

package identifier2

import (
	"fmt"
	"strings"
	"unicode"

	"gradientcontrol.com/omega/core/program"
)

func Validate(name string) error {
	if err := validateStructure(name); err != nil {
		return err
	}
	if program.IsReservedWord(name) || program.IsPrimitive(name) {
		if name == "auto" {
			return fmt.Errorf("\"auto\" is reserved and only permitted within .box property assignments")
		}
		return fmt.Errorf("identifier '%s' is a reserved word; try '%s' or another descriptive name", name, suggestRename(name))
	}
	if program.BuiltinNamed(name) != nil {
		return fmt.Errorf("identifier '%s' is a builtin function; try '%s' or another descriptive name", name, suggestRename(name))
	}
	return nil
}

// ValidatePackageName validates an import prefix / package name. Unlike
// Validate, it permits names that are reserved words, primitives, or builtins
// — e.g. a package named "map". A package qualifier appears only in
// `pkg.member` position, which the parser distinguishes from a bare `map(...)`
// primitive call (the primitive/builtin check keys off the member name, not
// the prefix), so there is no ambiguity. The structural rules still apply,
// including the reserved gcl_/gsl_ codegen prefixes.
func ValidatePackageName(name string) error {
	return validateStructure(name)
}

func validateStructure(name string) error {
	if name == "" {
		return fmt.Errorf("identifier cannot be empty")
	}
	firstRune := rune(name[0])
	if !unicode.IsLetter(firstRune) && firstRune != '_' {
		return fmt.Errorf("identifier '%s' must start with a letter or underscore", name)
	}
	for i, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return fmt.Errorf("identifier '%s' contains invalid character '%c' at position %d", name, r, i)
		}
	}
	if strings.HasPrefix(name, "gcl_") {
		return fmt.Errorf("identifier '%s' cannot start with the reserved prefix 'gcl_'", name)
	}
	if strings.Contains(name, "_gcl") {
		return fmt.Errorf("identifier '%s' cannot contain the reserved substring '_gcl'", name)
	}
	if strings.HasPrefix(name, "gsl_") {
		return fmt.Errorf("identifier '%s' cannot start with the reserved prefix 'gsl_'", name)
	}
	if strings.Contains(name, "_gsl") {
		return fmt.Errorf("identifier '%s' cannot contain the reserved substring '_gsl'", name)
	}
	return nil
}

// suggestRename returns a concrete replacement for a reserved or builtin
// name. The map covers names we've watched the LLM repeatedly trip on;
// anything else falls through to a trailing-underscore form so the model
// always has *something* concrete to copy in the follow-up turn.
func suggestRename(name string) string {
	if alt, ok := commonRenames[name]; ok {
		return alt
	}
	return name + "_"
}

var commonRenames = map[string]string{
	// Future-GSL keywords the model often grabs as the obvious noun.
	"shape":    "myShape",
	"out":      "result",
	"oneof":    "choice",
	"closest":  "nearest",
	"furthest": "farthest",

	// Builtin functions / primitives that read as nice nouns.
	"length":    "lenVal",
	"distance":  "dist",
	"dot":       "dotVal",
	"cross":     "crossVec",
	"normalize": "normVec",
	"clamp":     "clamped",
	"mix":       "mixVal",
	"step":      "stepVal",
	"sign":      "signVal",
	"min":       "minVal",
	"max":       "maxVal",
	"abs":       "absVal",
	"floor":     "floorVal",
	"ceil":      "ceilVal",
	"round":     "rounded",
	"sqrt":      "sqrtVal",
	"square":    "sq",
	"pow":       "powVal",
	"exp":       "expVal",
	"ln":        "lnVal",
	"sin":       "sinVal",
	"cos":       "cosVal",
	"tan":       "tanVal",
	"asin":      "asinVal",
	"acos":      "acosVal",
	"atan":      "atanVal",
	"mesh":      "meshOut",
	"remap":     "remapped",

	// Globals.
	"time": "t",
	"p":    "pt",

	// Type tags from WGSL/Rust/etc.
	"f32": "value",
	"f16": "halfVal",
	"u8":  "byteVal",
	"u32": "uintVal",
	"i32": "intVal",
	"i64": "longVal",
}

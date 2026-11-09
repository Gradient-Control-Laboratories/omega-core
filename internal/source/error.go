package source

import (
	"fmt"

	pubsource "gradientcontrol.com/omega/core/assembler/source"
)

var ErrPkgNotFound = pubsource.ErrPkgNotFound

type PackageError struct {
	Path string
	Err  error
}

func wrapError(path string, err error) error {
	if err == nil {
		return nil
	}
	return &PackageError{
		Path: path,
		Err:  err,
	}
}

func (e *PackageError) Error() string {
	return fmt.Sprintf("package %q: %s", e.Path, e.Err.Error())
}

func (e *PackageError) Unwrap() error {
	return e.Err
}

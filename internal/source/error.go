package source

import (
	"errors"
	"fmt"
)

var ErrPkgNotFound = errors.New("package not found")

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

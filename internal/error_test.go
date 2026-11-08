package internal

import (
	"errors"
	"fmt"
	"testing"
)

func TestInternalError(t *testing.T) {
	baseErr := fmt.Errorf("unexpected failure")
	ie := Err(baseErr)

	if ie == nil {
		t.Fatal("expected non-nil InternalError")
	}

	expectedMsg := "internal error: unexpected failure"
	if ie.Error() != expectedMsg {
		t.Errorf("expected %q, got %q", expectedMsg, ie.Error())
	}

	if !errors.Is(ie, baseErr) {
		t.Error("expected errors.Is(ie, baseErr) to be true")
	}

	if errors.Unwrap(ie) != baseErr {
		t.Errorf("expected unwrapped error to be %v, got %v", baseErr, errors.Unwrap(ie))
	}

	var target *InternalError
	if !errors.As(ie, &target) {
		t.Error("expected errors.As to match *InternalError")
	}
}


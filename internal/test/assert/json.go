package assert

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// JSONEq asserts that two JSON strings are equivalent.
func JSONEq(t *testing.T, expected, actual string) bool {
	t.Helper()
	var expectedVal, actualVal any

	if err := json.Unmarshal([]byte(expected), &expectedVal); err != nil {
		t.Logf("failed to unmarshal expected JSON: %v", err)
		return false
	}
	if err := json.Unmarshal([]byte(actual), &actualVal); err != nil {
		t.Logf("failed to unmarshal actual JSON: %v", err)
		return false
	}

	if !cmp.Equal(actualVal, expectedVal) {
		t.Logf("JSON mismatch (-want +got):\n%s", cmp.Diff(expectedVal, actualVal))
		return false
	}
	return true
}

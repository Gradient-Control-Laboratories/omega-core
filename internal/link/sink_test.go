package link

import (
	"testing"

	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/program"
)

func TestDiagnosticSinkLazyAllocation(t *testing.T) {
	sink := &diagnosticSink{}
	assert.Nil(t, sink.seen, "expected seen to be nil on construction")
	assert.Equal(t, 0, len(sink.diags))

	sink.add(program.Diagnostic{
		Severity: program.SeverityWarning,
		Message:  "test warning",
	})

	assert.NotNil(t, sink.seen, "expected seen to be allocated after adding diagnostic")
	assert.Equal(t, 1, len(sink.diags))

	// Adding duplicate diagnostic does not duplicate
	sink.add(program.Diagnostic{
		Severity: program.SeverityWarning,
		Message:  "test warning",
	})
	assert.Equal(t, 1, len(sink.diags))
}

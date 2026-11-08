package codegen_test

import (
	"fmt"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

func TestConstFoldingSimple(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{

		// ---- math ----

		{
			input:    "return 1 + 2 + 3;",
			expected: "6.0",
		},
		{
			input:    "return 1 * 2 * 3;",
			expected: "6.0",
		},
		{
			input:    "return 2 * 3 + 4 * 5;",
			expected: "26.0",
		},
		{
			input:    "return (10 - 2) * 3;",
			expected: "24.0",
		},
		{
			input:    "return 10 - 2 * 3;",
			expected: "4.0",
		},
		{
			input:    "return 5 % 2;",
			expected: "1.0",
		},

		// ---- math on lists ----

		{
			input:    "return 1 + [2, 3, 4];",
			expected: "vec3<f32>(3.0, 4.0, 5.0)",
		},
		{
			input:    "return [1, 2] + [3, 4];",
			expected: "vec2<f32>(4.0, 6.0)",
		},
		{
			input:    "return [10, 20] * 2;",
			expected: "vec2<f32>(20.0, 40.0)",
		},
		{
			input:    "return 2 * [10, 20];",
			expected: "vec2<f32>(20.0, 40.0)",
		},
		{
			input:    "return [1, 2, 3, 4] / 2;",
			expected: "vec4<f32>(0.5, 1.0, 1.5, 2.0)",
		},
		{
			input:    "return [1, 2, 3, 4, 5] + 10;",
			expected: "array<f32, 5>(11.0, 12.0, 13.0, 14.0, 15.0)",
		},

		// ---- boolean -----

		{
			input:    "return 1 && 1;",
			expected: "1.0",
		},
		{
			input:    "return 0 || 1;",
			expected: "1.0",
		},
		{
			input:    "return 1.0 ? 10.0 : 20.0;",
			expected: "10.0",
		},
		{
			input:    "return 0.0 ? 10.0 : 20.0;",
			expected: "20.0",
		},
		{
			input:    "return -1.0 ? 10.0 : 20.0 : 30.0;",
			expected: "30.0",
		},
		{
			input:    "return select(1, 10.0, 20.0);",
			expected: "10.0",
		},
		{
			input:    "return select(0, 10.0, 20.0);",
			expected: "20.0",
		},
		{
			input:    "return select(-1, 10.0, 20.0, 30.0);",
			expected: "30.0",
		},
		{
			input:    "return select(0, 10.0, 20.0, 30.0);",
			expected: "20.0",
		},
		{
			input:    "return select(1, 10.0, 20.0, 30.0);",
			expected: "10.0",
		},

		// ----- boolean on lists -----

		{
			input:    "return [1, 0] && [0, 1];",
			expected: "vec2<f32>(0.0, 0.0)",
		},
		{
			input:    "return [1, 0] || [0, 1];",
			expected: "vec2<f32>(1.0, 1.0)",
		},
		{
			input:    "return 1 ? [10, 20] : [30, 40];",
			expected: "vec2<f32>(10.0, 20.0)",
		},
		{
			input:    "return 0 ? [10, 20] : [30, 40];",
			expected: "vec2<f32>(30.0, 40.0)",
		},
		{
			input:    "return -1 ? [10, 20] : [30, 40] : [50, 60];",
			expected: "vec2<f32>(50.0, 60.0)",
		},
		{
			input:    "return 0 ? [10, 20] : [30, 40] : [50, 60];",
			expected: "vec2<f32>(30.0, 40.0)",
		},
		{
			input:    "return 1 ? [10, 20] : [30, 40] : [50, 60];",
			expected: "vec2<f32>(10.0, 20.0)",
		},
		{
			input:    "return select([1, 0], [10, 20], [30, 40]);",
			expected: "vec2<f32>(10.0, 40.0)",
		},
		{
			input:    "return [1, 0] ? [10, 20] : [30, 40];",
			expected: "vec2<f32>(10.0, 40.0)",
		},
		{
			input:    "return select([0, 1], [10, 20], [30, 40]);",
			expected: "vec2<f32>(30.0, 20.0)",
		},
		{
			input:    "return [-1, 0] ? [10, 20] : [30, 40] : [50, 60];",
			expected: "vec2<f32>(50.0, 40.0)",
		},

		// ----- op calls (min, max, pow, atan2) -----

		{
			input:    "return min(2, 3, 1);",
			expected: "1.0",
		},
		{
			input:    "return min([10, 20], [5, 30]);",
			expected: "vec2<f32>(5.0, 20.0)",
		},
		{
			input:    "return max(2, 3, 1);",
			expected: "3.0",
		},
		{
			input:    "return max([10, 20], [5, 30]);",
			expected: "vec2<f32>(10.0, 30.0)",
		},
		{
			input:    "return pow(2, 3);",
			expected: "8.0",
		},
		{
			input:    "return pow([2, 3], [3, 2]);",
			expected: "vec2<f32>(8.0, 9.0)",
		},
		{
			input:    "return atan(1, 1);",
			expected: "0.7853982",
		},
		{
			input:    "return atan([1, 0], [1, 1]);",
			expected: "vec2<f32>(0.7853982, 0.0)",
		},

		// ----- unary calls (square, sqrt, ...) -----

		{
			input:    "return square(3.0);",
			expected: "9.0",
		},
		{
			input:    "return square([2, 4]);",
			expected: "vec2<f32>(4.0, 16.0)",
		},
		{
			input:    "return sqrt(16.0);",
			expected: "4.0",
		},
		{
			input:    "return floor(1.8);",
			expected: "1.0",
		},
		{
			input:    "return ceil(1.2);",
			expected: "2.0",
		},
		{
			input:    "return round(1.5);",
			expected: "2.0",
		},
		{
			input:    "return abs(-5.0);",
			expected: "5.0",
		},
		{
			input:    "return abs([-1.0, 2.0]);",
			expected: "vec2<f32>(1.0, 2.0)",
		},
		{
			input:    "return exp(0.0);",
			expected: "1.0",
		},
		{
			input:    "return ln(1.0);",
			expected: "0.0",
		},
		{
			input:    "return sin(0.0);",
			expected: "0.0",
		},
		{
			input:    "return cos(0.0);",
			expected: "1.0",
		},
		{
			input:    "return not(0.0);",
			expected: "1.0",
		},
		{
			input:    "return not(1.0);",
			expected: "0.0",
		},
		{
			input:    "return neg(1.0);",
			expected: "-1.0",
		},

		// ----- GSL builtins -----

		{
			input:    "return cosh(0.0);",
			expected: "1.0",
		},
		{
			input:    "return cosh([0.0, 0.0]);",
			expected: "vec2<f32>(1.0, 1.0)",
		},
		{
			input:    "return tanh(0.0);",
			expected: "0.0",
		},
		{
			input:    "return tanh([0.0, 0.0]);",
			expected: "vec2<f32>(0.0, 0.0)",
		},
		{
			input:    "return degrees(3.141592653589793);",
			expected: "180.0",
		},
		{
			input:    "return degrees([3.141592653589793, 0.0]);",
			expected: "vec2<f32>(180.0, 0.0)",
		},
		{
			input:    "return radians(180.0);",
			expected: "3.1415927",
		},
		{
			input:    "return radians([180.0, 0.0]);",
			expected: "vec2<f32>(3.1415927, 0.0)",
		},
		{
			input:    "return sign(-15.5);",
			expected: "-1.0",
		},
		{
			input:    "return sign([-15.5, 1.0, 0.0]);",
			expected: "vec3<f32>(-1.0, 1.0, 0.0)",
		},
		{
			input:    "return length([3.0, 4.0]);",
			expected: "5.0",
		},
		{
			input:    "return distance([1.0, 2.0], [4.0, 6.0]);",
			expected: "5.0",
		},
		{
			input:    "return dot([2.0, 3.0], [4.0, 5.0]);",
			expected: "23.0",
		},
		{
			input:    "return cross([1.0, 0.0, 0.0], [0.0, 1.0, 0.0]);",
			expected: "vec3<f32>(0.0, 0.0, 1.0)",
		},
		{
			input:    "return normalize([0.0, 5.0, 0.0]);",
			expected: "vec3<f32>(0.0, 1.0, 0.0)",
		},
		{
			input:    "return clamp([0.5, 12.0, -3.0], 0.0, 10.0);",
			expected: "vec3<f32>(0.5, 10.0, 0.0)",
		},
		{
			input:    "return clamp([0.5, 12.0, -3.0], [0.0, 0.0, 0.0], [10.0, 10.0, 10.0]);",
			expected: "vec3<f32>(0.5, 10.0, 0.0)",
		},

		// ---- other ----
		{
			input:    "return (1 + 2 + 3);",
			expected: "6.0",
		},
		{
			input:    "a = 3; return a + 7;",
			expected: "10.0",
		},
		{
			input:    "a = [3, 4]; return a.y;",
			expected: "4.0",
		},
		{
			input:    "a = [3, 4]; return a.yxy;",
			expected: "vec3<f32>(4.0, 3.0, 4.0)",
		},
		{
			input:    "a = [3, 4]; return a[1];",
			expected: "4.0",
		},
		{
			input:    "a = [3, 4]; return [a[1], a[0]];",
			expected: "vec2<f32>(4.0, 3.0)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			res, err := assembler.Assemble([]string{tc.input}, assembler.SourceManager(&source.MockManager{}))
			require.NoError(t, err)
			require.NotNil(t, res)
			require.NotNil(t, res.Program)

			renderResult, err := renderHelper(tc.input, res.Program)
			require.NoError(t, err)
			require.NotNil(t, renderResult)

			expectedShape := fmt.Sprintf("let gsl_s1 = gsl_shape(%s, 1u);", tc.expected)
			assert.True(t, strings.Contains(renderResult.WGSL, expectedShape),
				"Expected WGSL to contain %q, but got:\n%s", expectedShape, renderResult.WGSL)
		})
	}
}

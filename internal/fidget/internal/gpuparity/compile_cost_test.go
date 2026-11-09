package gpuparity

import (
	"encoding/binary"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/program/noise"
)

// TestShaderCompileCost measures what each example's shader costs to compile,
// the time a material takes to appear, and writes one CSV row per example. It
// runs only when GSL_COMPILE_COST names the CSV to write; GSL_COMPILE_COST_RUNS
// sets how many cold compiles each takes (3 without it), and
// GSL_COMPILE_COST_MATCH, when set, keeps only the examples whose names
// contain it.
//
// Each row: the example; its WGSL golden's size in bytes and lines; the
// assembler's time on its source; and gpu-eval's time creating the shader
// module (naga's parse and validation) and the pipeline (the backend's
// translation and the driver's compile), as the median of the cold runs and
// one warm run. The shader is the field-parity test's (assembleShader), a
// compute pass over the golden, which the viewers' render pipeline does not
// match exactly but compiles the same functions. A cold run renames a private
// variable the entry point reads, so the backend's source differs every time
// and the driver's shader cache (Metal's, across processes) cannot answer it;
// the warm run repeats the last cold run's source.
func TestShaderCompileCost(t *testing.T) {
	out := os.Getenv("GSL_COMPILE_COST")
	if out == "" {
		t.Skip("GSL_COMPILE_COST names the CSV to write")
	}
	runs := 3
	if v := os.Getenv("GSL_COMPILE_COST_RUNS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			t.Fatalf("GSL_COMPILE_COST_RUNS: %q", v)
		}
		runs = n
	}
	bin := filepath.Join(repoRoot, "target/release/gpu-eval")
	if _, err := os.Stat(bin); err != nil {
		t.Skipf("gpu-eval is not built (tool/build-fidget builds it): %v", err)
	}
	host := readHostWGSL(t)
	noisePath := filepath.Join(t.TempDir(), "noise_tile.bin")
	if err := os.WriteFile(noisePath, noise.Bake(), 0o644); err != nil {
		t.Fatal(err)
	}

	files, err := filepath.Glob(filepath.Join(repoRoot, "internal/testdata/example/*.gsl"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no examples found: %v", err)
	}
	f, err := os.Create(out)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	w.Write([]string{"example", "wgsl_bytes", "wgsl_lines", "assemble_ms",
		"cold_module_ms", "cold_pipeline_ms", "warm_module_ms", "warm_pipeline_ms", "note"})

	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".gsl")
		if m := os.Getenv("GSL_COMPILE_COST_MATCH"); m != "" && !strings.Contains(name, m) {
			continue
		}
		row := []string{name, "", "", "", "", "", "", "", ""}
		note := func(s string) { row[8] = s }
		func() {
			if why, ok := skipExamples[name]; ok {
				note("skipped: " + why)
				return
			}
			golden, err := os.ReadFile(filepath.Join(repoRoot, "backend/wgsl/internal/codegen/testdata/example", name+".wgsl"))
			if err != nil {
				note("no WGSL golden")
				return
			}
			wgsl := string(golden)
			row[1] = strconv.Itoa(len(wgsl))
			row[2] = strconv.Itoa(strings.Count(wgsl, "\n"))
			src, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			start := time.Now()
			if _, err := assembler.Assemble([]string{string(src)}, assembler.SourceManager(pkgMgr)); err != nil {
				note("assemble: " + err.Error())
				return
			}
			row[3] = ms(time.Since(start))

			sig := sdShapeRE.FindStringSubmatch(wgsl)
			if sig == nil {
				note("no sdShape")
				return
			}
			usesNoise := false
			switch sig[1] {
			case "":
			case ", noiseTex: texture_3d<f32>":
				usesNoise = true
			default:
				note("sdShape binds textures the harness does not provide")
				return
			}
			n := maxShapeIndex(wgsl)
			hasColor := strings.Contains(wgsl, "fn gsl_colorField(")
			hasMaterial := strings.Contains(wgsl, "fn gsl_materialField(")
			shader := assembleShader(host, wgsl, n, usesNoise, hasColor, hasMaterial)

			var mods, pipes []float64
			var last string
			for r := 0; r < runs; r++ {
				last = withNonce(shader)
				m, p, err := compileTime(bin, last, stride(n, hasColor, hasMaterial), usesNoise, noisePath)
				if err != nil {
					note(err.Error())
					return
				}
				mods, pipes = append(mods, m), append(pipes, p)
			}
			row[4], row[5] = msF(median(mods)), msF(median(pipes))
			m, p, err := compileTime(bin, last, stride(n, hasColor, hasMaterial), usesNoise, noisePath)
			if err != nil {
				note(err.Error())
				return
			}
			row[6], row[7] = msF(m), msF(p)
		}()
		w.Write(row)
		w.Flush()
		t.Logf("%s: %v", name, row[1:])
	}
}

// withNonce gives the shader a private variable of a name no run has used,
// which its entry point reads, so the backend's source is new.
func withNonce(shader string) string {
	name := fmt.Sprintf("gsl_nonce_%x", rand.Uint64())
	shader = strings.Replace(shader, "\n@compute", fmt.Sprintf("\nvar<private> %s: f32 = 1.0;\n@compute", name), 1)
	return strings.Replace(shader, "    gsl_out[o] = s.distance;\n", fmt.Sprintf("    gsl_out[o] = s.distance * %s;\n", name), 1)
}

// compileTime runs gpu-eval once with --timing over a single workgroup of
// points and returns its module and pipeline milliseconds.
func compileTime(bin, shader string, stride int, usesNoise bool, noisePath string) (float64, float64, error) {
	dir, err := os.MkdirTemp("", "compile-cost-*")
	if err != nil {
		return 0, 0, err
	}
	defer os.RemoveAll(dir)
	shaderPath := filepath.Join(dir, "shader.wgsl")
	pointsPath := filepath.Join(dir, "points.f32")
	if err := os.WriteFile(shaderPath, []byte(shader), 0o644); err != nil {
		return 0, 0, err
	}
	buf := make([]byte, 16*64)
	for i := 0; i < 64; i++ {
		binary.LittleEndian.PutUint32(buf[16*i:], math.Float32bits(float32(i)))
	}
	if err := os.WriteFile(pointsPath, buf, 0o644); err != nil {
		return 0, 0, err
	}
	args := []string{"--shader", shaderPath, "--points", pointsPath, "--stride", strconv.Itoa(stride),
		"--out", filepath.Join(dir, "out.f32"), "--timing", "1"}
	if usesNoise {
		args = append(args, "--noise", noisePath)
	}
	cmd := exec.Command(bin, args...)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return 0, 0, fmt.Errorf("gpu-eval: %v: %s", err, strings.TrimSpace(stderr.String()))
	}
	for _, line := range strings.Split(stderr.String(), "\n") {
		var tm struct {
			Module   float64 `json:"module_ms"`
			Pipeline float64 `json:"pipeline_ms"`
		}
		if json.Unmarshal([]byte(line), &tm) == nil && tm.Pipeline > 0 {
			return tm.Module, tm.Pipeline, nil
		}
	}
	return 0, 0, fmt.Errorf("gpu-eval printed no timing: %s", strings.TrimSpace(stderr.String()))
}

func median(v []float64) float64 {
	s := append([]float64(nil), v...)
	sort.Float64s(s)
	return s[len(s)/2]
}

func ms(d time.Duration) string { return msF(float64(d.Microseconds()) / 1000) }

func msF(v float64) string { return strconv.FormatFloat(v, 'f', 1, 64) }

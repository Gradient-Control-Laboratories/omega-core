package wgsl

import (
	api "gradientcontrol.com/omega/core/backend/wgsl/internal/lib"
	"gradientcontrol.com/omega/core/backend/wgsl/internal/lib/lib"
	"gradientcontrol.com/omega/core/program"
)

// RenderResult contains the emitted WGSL shader source code and geometry metadata.
type RenderResult = api.RenderResult

// ShapeLocations contains source spans that contribute to a shape.
type ShapeLocations = api.ShapeLocations

// Part represents a part bounding box in world coordinates.
type Part = api.Part

// Option configures WGSL shader rendering.
type Option func(*config)

type config struct {
	standalone bool
	progressFn ProgressFunc
}

// ProgressFunc reports compilation progress.
type ProgressFunc func(stage string, completed, total uint32)

// WithStandalone configures WGSL emission in standalone mode.
func WithStandalone(standalone bool) Option {
	return func(c *config) {
		c.standalone = standalone
	}
}

// WithProgress supplies a progress callback during WGSL emission.
func WithProgress(fn ProgressFunc) Option {
	return func(c *config) {
		c.progressFn = fn
	}
}

const standaloneHarness = `
struct CameraUniforms {
  viewProj: mat4x4<f32>,
  invViewProj: mat4x4<f32>,
  eyePos: vec3<f32>,
  time: f32,
};

@group(0) @binding(0) var<uniform> camera: CameraUniforms;

struct VertexOutput {
  @builtin(position) position: vec4<f32>,
  @location(0) uv: vec2<f32>,
};

@vertex
fn vs_main(@builtin(vertex_index) vertexIndex: u32) -> VertexOutput {
  var pos = array<vec2<f32>, 3>(
    vec2<f32>(-1.0, -1.0),
    vec2<f32>( 3.0, -1.0),
    vec2<f32>(-1.0,  3.0)
  );
  var out: VertexOutput;
  out.position = vec4<f32>(pos[vertexIndex], 0.0, 1.0);
  out.uv = pos[vertexIndex] * 0.5 + vec2<f32>(0.5, 0.5);
  return out;
}

@fragment
fn fs_main(in: VertexOutput) -> @location(0) vec4<f32> {
  let ndc = vec4<f32>(in.uv * 2.0 - 1.0, 1.0, 1.0);
  let worldTarget = camera.invViewProj * ndc;
  let rayDir = normalize(worldTarget.xyz / worldTarget.w - camera.eyePos);
  
  var t: f32 = 0.01;
  let maxDist: f32 = 100.0;
  for (var i: i32 = 0; i < 128; i = i + 1) {
    let p = camera.eyePos + rayDir * t;
    let res = sdShape(p, camera.time);
    if (res.d < 0.001) {
      let eps = 0.001;
      let nx = sdShape(p + vec3<f32>(eps, 0.0, 0.0), camera.time).d - sdShape(p - vec3<f32>(eps, 0.0, 0.0), camera.time).d;
      let ny = sdShape(p + vec3<f32>(0.0, eps, 0.0), camera.time).d - sdShape(p - vec3<f32>(0.0, eps, 0.0), camera.time).d;
      let nz = sdShape(p + vec3<f32>(0.0, 0.0, eps), camera.time).d - sdShape(p - vec3<f32>(0.0, 0.0, eps), camera.time).d;
      let normal = normalize(vec3<f32>(nx, ny, nz));
      let lightDir = normalize(vec3<f32>(1.0, 2.0, 3.0));
      let diff = max(dot(normal, lightDir), 0.1);
      return vec4<f32>(vec3<f32>(0.8, 0.8, 0.9) * diff, 1.0);
    }
    t = t + res.d;
    if (t > maxDist) {
      break;
    }
  }
  return vec4<f32>(0.05, 0.05, 0.08, 1.0);
}
`

// Render emits WebGPU shader code from the given GSL source and assembled program AST.
func Render(source string, prog *program.Program, opts ...Option) (*RenderResult, error) {
	var cfg config
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.progressFn != nil {
		cfg.progressFn("Emitting WGSL shader", 1, 1)
	}
	res, err := lib.Render(source, prog)
	if err != nil {
		return nil, err
	}
	if cfg.standalone {
		res.WGSL = res.WGSL + "\n" + standaloneHarness
	}
	return res, nil
}

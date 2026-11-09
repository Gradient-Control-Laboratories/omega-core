// Package gpuparity holds the GPU field-parity test: it checks that the WGSL a
// GSL document compiles to evaluates to the same distance, color and material
// as fidget does for the same document, point by point.
//
// The two sides share only the GSL front end (common/build and the property
// resolver). Everything after it is a separate backend: agent/wgsl's codegen
// on one side, agent/fidget/internal/convert on the other. A mismatch here is
// a bug in one of them, found without a camera, a march or a tone map in the
// way.
//
// The WGSL comes from the example goldens in
// agent/wgsl/internal/codegen/testdata/example, which TestBuildExamples pins
// to the codegen output. It runs on the GPU through gpu-eval
// (agent/fidget/internal/gpu-eval), a headless wgpu runner that
// tool/build-fidget builds. When the binary or a GPU adapter is missing, the
// test skips.
//
// wgpu compiles WGSL with naga, and Chrome compiles it with Tint, so this
// proves the generated WGSL, not a browser's compiler.
//
// TestGpuImageParity is the second layer: the viewer's standard frame against
// litrender. The frame is the shader the viewers draw with, built by Node from
// common/web/gsl-render-graph.js (js/standard-frame.mjs) and run by gpu-eval
// one pixel per point; litrender draws the same camera in its viewer mode.
// It compares which return each pixel shows and its tone-mapped colour, so it
// covers the march, the normals, the ambient occlusion and the shading that
// the field test leaves out. It also skips when node is missing.
package gpuparity

// Builds the viewer's standard frame for one compiled program as a gpu-eval
// shader, for TestGpuImageParity:
//
//   node --import ./register-three.mjs standard-frame.mjs < job.json > frame.wgsl
//
// The frame is gsl_wgslFragment exactly as the viewers build it
// (common/web/gsl-render-graph.js), with every uniform baked in from job.json
// and the shading knobs at their defaults. One invocation per point: the
// point's xy is the pixel's uv (0,0 top left), and the fragment's vec4 goes
// to gsl_out: tone-mapped colour, and shapeIndex + depth in alpha. A miss
// writes zeros, which is what the viewer's discard leaves in its cleared
// target.
//
// job.json:
//   golden       path of the program's WGSL
//   shapeCount   the number of returns
//   height       image height in pixels (for the pixel footprint)
//   camera       {pos, right, up, fwd, tanFov: [x, y], ortho}, world Z-up
//   worldExtent  the scene box half-size
//   far          the far plane depth is measured against
//   colors       each return's colour, [[r, g, b], ...]
//   sky, ground, background   the environment, [r, g, b]
import fs from "node:fs";
import { buildRenderGraph } from "../../../../../common/web/gsl-render-graph.js";
import { snapshotCall } from "../../../../../common/web/gsl-render-job.js";
import { shadingKnobsU, glassKnobsU, surfaceKnobsU, renderKnobsU } from "../../../../../common/web/gsl-shading.js";
import { envMatrices } from "../../../../../common/web/gsl-environment.js";

const job = JSON.parse(fs.readFileSync(0, "utf8"));
const wgslCode = fs.readFileSync(job.golden, "utf8");
const noise = /\bnoiseTex\b/.test(wgslCode);
const graph = buildRenderGraph({
  wgslCode,
  shapeCount: job.shapeCount,
  extraParamsWGSL: noise ? ", noiseTex: texture_3d<f32>" : "",
  extraArgsWGSL: noise ? ", noiseTex" : "",
});

const c = job.camera;
// World Z-up directions into view space (x right, y up, z toward the eye),
// column-major. The shading only turns directions with it.
const zUpToView = [
  c.right[0], c.up[0], -c.fwd[0], 0,
  c.right[1], c.up[1], -c.fwd[1], 0,
  c.right[2], c.up[2], -c.fwd[2], 0,
  0, 0, 0, 1,
];
const overrides = {
  camPosZUp: c.pos,
  camRight: c.right,
  camUp: c.up,
  camFwd: c.fwd,
  camTanFov: c.tanFov,
  camOrtho: c.ortho ? 1 : 0,
  projViewZUp: [1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1],
  zUpToView,
  skyCol: job.sky,
  groundCol: job.ground,
  // The Viewer environment: the sky and ground above, the rig at full
  // strength (gsl-environment.js).
  ...(({ a, b, c }) => ({ envA: a, envB: b, envC: c }))(envMatrices("Viewer")),
  shadingKnobs: shadingKnobsU.value,
  glassKnobs: glassKnobsU.value,
  surfaceKnobs: surfaceKnobsU.value,
  renderKnobs: renderKnobsU.value,
  // Every rung of the Detail ladder off: parts keep their full detail at any
  // size on screen, as the CPU renderer draws them.
  detailKnobs: [0, 0, 0, 0],
  detailKnobs2: [0, 0, 0, 0],
  bgColor: job.background,
  // The viewer's pixel footprint at unit distance (updateCameraUniforms).
  pixelScale: (2 * c.tanFov[1]) / job.height,
  worldExtent: job.worldExtent,
  // Depth rides in alpha under the shape index as depth / far, clamped below
  // 0.498; job.far keeps it inside that across the scene box.
  nearFar: [0.01, job.far],
  time: 0,
  fieldActive: 0,
  fieldHeight: 0,
  fieldNormal: [0, 0, 1],
};
for (let i = 1; i <= job.shapeCount; i++) {
  overrides[`c${i}`] = job.colors[i - 1];
  overrides[`vis${i}`] = 1;
}

const call = snapshotCall({
  fn: graph.wgslFragmentShader,
  name: "gsl_wgslFragment",
  overrides,
  entryArgs: { uv: "uv", ...(noise ? { noiseTex: "gsl_noise" } : {}) },
});
if ((call.src.match(/\bdiscard;/g) || []).length !== 1) {
  throw new Error("expected the standard fragment's one discard");
}
const src = call.src.replace(/\bdiscard;/, "return vec4<f32>(0.0);");

process.stdout.write(`${src}

@group(0) @binding(0) var<storage, read> gsl_points: array<vec4<f32>>;
@group(0) @binding(1) var<storage, read_write> gsl_out: array<f32>;
${noise ? "@group(0) @binding(2) var gsl_noise: texture_3d<f32>;" : ""}
${call.decls.join("\n")}

@compute @workgroup_size(64)
fn main(@builtin(global_invocation_id) gid: vec3<u32>) {
    let i = gid.x;
    if (i >= arrayLength(&gsl_points)) { return; }
    let uv = gsl_points[i].xy;
    let c = gsl_wgslFragment(${call.args.join(", ")});
    gsl_out[i * 4u] = c.x;
    gsl_out[i * 4u + 1u] = c.y;
    gsl_out[i * 4u + 2u] = c.z;
    gsl_out[i * 4u + 3u] = c.w;
}
`);

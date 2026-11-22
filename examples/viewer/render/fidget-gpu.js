// The fidget rest device: an omega tape drawn in the page by fidget's own GPU
// pipeline (fidget-wgpu compiled to wasm, vendor/fidget-wgpu, built by
// tool/build-fidget.sh), from the viewer's camera, and lit as the standard frame
// lights a hit.
//
//   const frame = await fidgetGpuFrame(tapeBytes, snap, colors);
//
// snap is the viewer's CPU snapshot ({width, height, cam: {pos, right, up, fwd,
// tanFov, ortho}, camArgs, backDim, shading: {zUpToView, sky, ground, knobs}},
// camArgs ending in the world extent); colors the viewer's colour per return,
// for a return whose tape gives none. The frame is the CPU tier's: {width,
// height, data}, five floats a pixel from the top left, the lit colour, the
// return hit (0 for none) and the depth along fwd; plus, per pixel, slice (how
// deep one of fidget's voxels is there, along fwd) and transmission (the
// material's, 0 for opaque), for a comparison to judge it by. options.occlusion
// false lights every hit unoccluded, leaving SSAO out.
//
// fidget finds each pixel's hit, the return it belongs to, and that return's
// colour and material fields there (the tape carries them: its appearance),
// plus SSAO; gsl_shadeSurface, the standard frame's own WGSL, lights it here in
// a compute pass. Glass shades as an opaque surface of its material, its
// transmission set to 0, and the occlusion is fidget's SSAO rather than the
// standard frame's marched AO.

import { GSL_SHADING_WGSL, GSL_MATERIAL_STRUCT_WGSL } from "./gsl-shading.js";

// fidget's normal is the field's gradient in its pixel coordinates (x right,
// y down, z toward the viewer), not in its [-1, 1] screen coordinates.
const NORMAL_IN_PIXELS = true;

let module = null;
let ready = null;
let loadedKey = null;
let light = null;

async function initOnce() {
  if (!navigator.gpu) throw new Error("this browser has no WebGPU");
  module ??= await import("../vendor/fidget-wgpu/fidget_wgpu.js").catch(() => {
    console.log(
      `[fidget-gpu] /vendor/fidget-wgpu/fidget_wgpu.js is not built.\n` +
      `To build fidget-wgpu:\n` +
      `  1. Install the Rust WebAssembly target and the wasm-bindgen-cli that\n` +
      `     internal/fidget/internal/wgpu-web/Cargo.lock names:\n` +
      `       rustup target add wasm32-unknown-unknown\n` +
      `       cargo install wasm-bindgen-cli --version <the lock's wasm-bindgen>\n` +
      `  2. Build the bundle from the omega-core root:\n` +
      `       ./tool/build-fidget.sh --wasm-out examples/viewer/vendor\n` +
      `  3. Reload this page.`
    );
    throw new Error("fidget-wgpu is not built; see browser console for build instructions");
  });
  ready ??= module.default().then(() => module.init());
  await ready;
}

// depthRange is where fidget's voxel columns start and stop along the view:
// the span of the sphere around the scene cube (worldExtent on a side, as the
// viewer frames it), not the tape's box, which is padded, extended to cover
// the box around the guarded parts (bounds, six floats, or null) where those
// reach out of it: a scene wider than the viewer's cube has parts outside the
// sphere, which the near and far planes would cut. The sphere stays the floor
// because a part with no box (one a remap moved) is not in bounds. Perspective
// slices are uniform in inverse depth, so a range reaching back to the camera
// would crowd every slice beside it and leave the scene between two; a camera
// inside the range starts a hundredth of the way to its far side.
function depthRange(cam, worldExtent, bounds) {
  const r = (worldExtent / 2) * Math.sqrt(3);
  const center = -(cam.pos[0] * cam.fwd[0] + cam.pos[1] * cam.fwd[1] + cam.pos[2] * cam.fwd[2]);
  let nearD = center - r;
  let farD = center + r;
  if (bounds?.length === 6) {
    // The depth along the view of each corner of the box.
    for (const x of [bounds[0], bounds[3]]) {
      for (const y of [bounds[1], bounds[4]]) {
        for (const z of [bounds[2], bounds[5]]) {
          const d = (x - cam.pos[0]) * cam.fwd[0] + (y - cam.pos[1]) * cam.fwd[1] + (z - cam.pos[2]) * cam.fwd[2];
          if (Number.isFinite(d)) {
            nearD = Math.min(nearD, d);
            farD = Math.max(farD, d);
          }
        }
      }
    }
  }
  const far = Math.max(farD, 1e-6);
  const near = cam.ortho > 0.5 ? nearD : Math.max(nearD, far * 0.01);
  return { near, far };
}

// worldToModel takes fidget's screen coordinates to world space: a pixel
// (px, py) at voxel depth pz sits at ((px - W/2) k, -(py - H/2 + 1) k,
// (pz - D/2) k), k = 2 / min(W, H, D), with z toward the viewer. Those are
// rescaled to the camera's [-1, 1] square at pixel centres, then taken along
// the view: linearly for an orthographic camera, and for a perspective one by
// a projective map whose w is the inverse depth, near at z = 1 and far at
// z = -1, the way a projection matrix slices a frustum. Column-major.
function worldToModel(snap, depth, near, far) {
  const { width: W, height: H, cam } = snap;
  const k = 2 / Math.min(W, H, depth);
  const [tx, ty] = cam.tanFov;
  const { pos, right: R, up: U, fwd: F } = cam;
  // N maps (u, v, w, 1) to homogeneous world space, columns first.
  let N;
  if (cam.ortho > 0.5) {
    const half = (far - near) / 2;
    N = [
      [tx * R[0], tx * R[1], tx * R[2], 0],
      [ty * U[0], ty * U[1], ty * U[2], 0],
      [-half * F[0], -half * F[1], -half * F[2], 0],
      [0, 1, 2].map((i) => pos[i] + (near + half) * F[i]).concat(1),
    ];
  } else {
    const a = (1 / near - 1 / far) / 2;
    const b = (1 / near + 1 / far) / 2;
    N = [
      [tx * R[0], tx * R[1], tx * R[2], 0],
      [ty * U[0], ty * U[1], ty * U[2], 0],
      [a * pos[0], a * pos[1], a * pos[2], a],
      [0, 1, 2].map((i) => F[i] + b * pos[i]).concat(b),
    ];
  }
  // S maps fidget's screen coordinates to (u, v, w): u = 2 sx / (W k) + 1/W,
  // v = 2 sy / (H k) + 1/H, w = 2 sz / (D k).
  const su = 2 / (W * k);
  const sv = 2 / (H * k);
  const sw = 2 / (depth * k);
  const m = new Float32Array(16);
  for (let r = 0; r < 4; r++) {
    m[0 + r] = N[0][r] * su;
    m[4 + r] = N[1][r] * sv;
    m[8 + r] = N[2][r] * sw;
    m[12 + r] = N[0][r] / W + N[1][r] / H + N[3][r];
  }
  return m;
}

// The standard frame's lighting, run over the G-buffer: gsl_shadeSurface as
// the standard fragment calls it for a front hit, the normal flipped and the
// colour dimmed for a face seen from behind.
const LIGHT_WGSL = `
${GSL_MATERIAL_STRUCT_WGSL}
${GSL_SHADING_WGSL}
struct Params {
    zUpToView: mat4x4<f32>,
    camPos: vec4<f32>,
    sky: vec4<f32>,
    ground: vec4<f32>,
    knobs: vec4<f32>,
    backDim: f32,
    count: u32,
    rows: u32,
    _pad: u32,
}
@group(0) @binding(0) var<uniform> params: Params;
@group(0) @binding(1) var<storage, read> hit: array<vec4<f32>>;       // world xyz, return
@group(0) @binding(2) var<storage, read> normal: array<vec4<f32>>;    // world xyz, occlusion
@group(0) @binding(3) var<storage, read> base: array<vec4<f32>>;      // sRGB colour
@group(0) @binding(4) var<storage, read> material: array<vec4<f32>>;  // params.rows rows a pixel
@group(0) @binding(5) var<storage, read_write> lit: array<vec4<f32>>;

@compute @workgroup_size(64)
fn light_main(@builtin(global_invocation_id) id: vec3<u32>, @builtin(num_workgroups) groups: vec3<u32>) {
    // The dispatch is a grid of groups a row wide (a row holds 65535 at most).
    let i = id.y * groups.x * 64u + id.x;
    if (i >= params.count || hit[i].w < 0.5) { return; }
    let hitPos = hit[i].xyz;
    let n = normalize(normal[i].xyz);
    let viewDirWorld = normalize(params.camPos.xyz - hitPos);
    let facingAway = dot(-viewDirWorld, n) > 0.0;
    let localNormal = select(n, -n, facingAway);
    // Opaque: gsl_shadeSurface takes a transmissive material's transmission
    // out of its diffuse lobe for the glass path to add back from behind, and
    // nothing here traces behind, so glass would come out nearly black.
    // The record is params.rows vec4 rows a pixel: the surface, emissive and
    // volume rows, and from a four-row record the lay (zero from a three-row one).
    let r = params.rows;
    var volume = material[r * i + 2u];
    volume.x = 0.0;
    var lay = vec4<f32>(0.0);
    if (r > 3u) { lay = material[r * i + 3u]; }
    let mat = gsl_material(material[r * i], material[r * i + 1u], volume, lay, 1.0);
    let c = gsl_shadeSurface(hitPos, localNormal, viewDirWorld, base[i].rgb, mat, normal[i].w,
                             select(1.0, params.backDim, facingAway), params.zUpToView,
                             params.sky.xyz, params.ground.xyz, params.knobs);
    lit[i] = vec4<f32>(c, 1.0);
}
`;

async function lightOnce() {
  if (light) return light;
  const adapter = await navigator.gpu.requestAdapter({ powerPreference: "high-performance" });
  if (!adapter) throw new Error("no WebGPU adapter for the lighting pass");
  const device = await adapter.requestDevice({
    // The material record is 64 bytes a pixel: past about 2 megapixels it
    // passes the default storage binding (128 MiB), and past 4 the default
    // buffer size (256 MiB), which a Retina window reaches.
    requiredLimits: {
      maxStorageBufferBindingSize: adapter.limits.maxStorageBufferBindingSize,
      maxBufferSize: adapter.limits.maxBufferSize,
    },
  });
  const shader = device.createShaderModule({ code: LIGHT_WGSL });
  const info = await shader.getCompilationInfo();
  const errors = info.messages.filter((msg) => msg.type === "error");
  if (errors.length) throw new Error(`lighting shader: ${errors.map((e) => e.message).join("; ")}`);
  const pipeline = device.createComputePipeline({ layout: "auto", compute: { module: shader, entryPoint: "light_main" } });
  light = { device, pipeline };
  return light;
}

// lightFrame runs the lighting pass and returns the lit colour, 4 floats a pixel.
async function lightFrame(snap, count, hit, normal, base, material) {
  const { device, pipeline } = await lightOnce();
  const make = (data, usage) => {
    const b = device.createBuffer({ size: Math.max(16, data.byteLength), usage: usage | GPUBufferUsage.COPY_DST });
    device.queue.writeBuffer(b, 0, data);
    return b;
  };
  // The bundle sends the material record as 3 vec4 rows a pixel (surface,
  // emissive, volume) or 4 (and the lay); the shader reads params.rows of them.
  const rows = count ? material.length / (4 * count) : 3;
  if (!Number.isInteger(rows) || rows < 3 || rows > 4) {
    throw new Error(`the fidget material record is ${rows} rows a pixel, expected 3 or 4`);
  }
  const params = new ArrayBuffer(160);
  const f = new Float32Array(params);
  f.set(snap.shading.zUpToView, 0);
  f.set([...snap.cam.pos, 0], 16);
  f.set([...snap.shading.sky, 0], 20);
  f.set([...snap.shading.ground, 0], 24);
  f.set(snap.shading.knobs, 28);
  f[32] = snap.backDim;
  new Uint32Array(params)[33] = count;
  new Uint32Array(params)[34] = rows;
  // A buffer or dispatch past a device limit is a validation error that
  // leaves the output zero, a black frame: it is caught below, not left silent.
  device.pushErrorScope("validation");
  const buffers = [
    make(params, GPUBufferUsage.UNIFORM),
    make(hit, GPUBufferUsage.STORAGE),
    make(normal, GPUBufferUsage.STORAGE),
    make(base, GPUBufferUsage.STORAGE),
    make(material, GPUBufferUsage.STORAGE),
  ];
  const out = device.createBuffer({ size: count * 16, usage: GPUBufferUsage.STORAGE | GPUBufferUsage.COPY_SRC });
  const read = device.createBuffer({ size: count * 16, usage: GPUBufferUsage.MAP_READ | GPUBufferUsage.COPY_DST });
  const bind = device.createBindGroup({
    layout: pipeline.getBindGroupLayout(0),
    entries: [...buffers, out].map((buffer, binding) => ({ binding, resource: { buffer } })),
  });
  const enc = device.createCommandEncoder();
  const pass = enc.beginComputePass();
  pass.setPipeline(pipeline);
  pass.setBindGroup(0, bind);
  // A dispatch is at most 65535 groups wide, 4.19 megapixels of 64: a frame
  // past that is a grid of rows (light_main indexes it by num_workgroups).
  const groups = Math.ceil(count / 64);
  const row = Math.max(1, Math.min(groups, 65535));
  pass.dispatchWorkgroups(row, Math.ceil(groups / row));
  pass.end();
  enc.copyBufferToBuffer(out, 0, read, 0, count * 16);
  device.queue.submit([enc.finish()]);
  const failure = await device.popErrorScope();
  if (failure) {
    for (const b of [...buffers, out, read]) b.destroy();
    throw new Error(`lighting pass: ${failure.message}`);
  }
  await read.mapAsync(GPUMapMode.READ);
  const lit = new Float32Array(read.getMappedRange().slice(0));
  read.unmap();
  for (const b of [...buffers, out, read]) b.destroy();
  return lit;
}

// jacobian is d(world)/d(s) at screen point s, for world = (m s).xyz / (m s).w.
function jacobian(m, s) {
  const p = [0, 1, 2, 3].map((r) => m[r] * s[0] + m[4 + r] * s[1] + m[8 + r] * s[2] + m[12 + r]);
  const w = p[3];
  const J = [[0, 0, 0], [0, 0, 0], [0, 0, 0]];
  for (let i = 0; i < 3; i++) {
    for (let j = 0; j < 3; j++) J[i][j] = (m[4 * j + i] - (p[i] / w) * m[4 * j + 3]) / w;
  }
  return { J, world: [p[0] / w, p[1] / w, p[2] / w] };
}

// solveTranspose returns x with J^T x = g (the world gradient, from the
// screen-space one), by Cramer's rule.
function solveTranspose(J, g) {
  const a = [[J[0][0], J[1][0], J[2][0]], [J[0][1], J[1][1], J[2][1]], [J[0][2], J[1][2], J[2][2]]];
  const det = (M) =>
    M[0][0] * (M[1][1] * M[2][2] - M[1][2] * M[2][1]) -
    M[0][1] * (M[1][0] * M[2][2] - M[1][2] * M[2][0]) +
    M[0][2] * (M[1][0] * M[2][1] - M[1][1] * M[2][0]);
  const d = det(a);
  return [0, 1, 2].map((c) => det(a.map((row, r) => row.map((v, k) => (k === c ? g[r] : v)))) / d);
}

export async function fidgetGpuFrame(tape, snap, colors, { occlusion = true } = {}) {
  await initOnce();
  const palette = new Float32Array(colors.flat());
  const key = palette.join(",");
  if (loadedKey?.tape !== tape || loadedKey.palette !== key) {
    loadedKey = null;
    module.load(tape, palette);
    loadedKey = { tape, palette: key };
  }
  const { width: W, height: H, cam } = snap;
  const D = Math.max(W, H);
  const { near, far } = depthRange(cam, snap.camArgs[snap.camArgs.length - 1], snap.bounds);
  const m = worldToModel(snap, D, near, far);
  const g = await module.render(W, H, D, m);

  const k = 2 / Math.min(W, H, D);
  const count = W * H;
  const matStride = g.material.length / count; // floats a pixel: 12 (3 rows) or 16 (4)
  const hit = new Float32Array(4 * count);
  const normal = new Float32Array(4 * count);
  const base = new Float32Array(4 * count);
  const depth = new Float32Array(count);
  const slice = new Float32Array(count);
  const transmission = new Float32Array(count);
  const along = (p) => (p[0] - cam.pos[0]) * cam.fwd[0] + (p[1] - cam.pos[1]) * cam.fwd[1] + (p[2] - cam.pos[2]) * cam.fwd[2];
  for (let py = 0; py < H; py++) {
    for (let px = 0; px < W; px++) {
      const i = py * W + px;
      const idx = g.index[i];
      if (!idx) continue;
      const s = [(px - W / 2) * k, -(py - H / 2 + 1) * k, (g.depth[i] - D / 2) * k];
      const { J, world } = jacobian(m, s);
      // The gradient in s from fidget's, which is in pixels (y down) or in s.
      const gn = [g.normal[3 * i], g.normal[3 * i + 1], g.normal[3 * i + 2]];
      const gs = NORMAL_IN_PIXELS ? [gn[0] / k, -gn[1] / k, gn[2] / k] : gn;
      const nw = solveTranspose(J, gs);
      hit.set([...world, idx], 4 * i);
      const occ = g.occlusion[i];
      normal.set([...nw, occlusion && Number.isFinite(occ) && occ >= 0 ? occ : 1], 4 * i);
      base.set([g.color[4 * i] / 255, g.color[4 * i + 1] / 255, g.color[4 * i + 2] / 255, 1], 4 * i);
      depth[i] = along(world);
      slice[i] = Math.abs(along(jacobian(m, [s[0], s[1], s[2] + k]).world) - depth[i]);
      transmission[i] = g.material[matStride * i + 8];
    }
  }
  const lit = await lightFrame(snap, count, hit, normal, base, g.material);

  const data = new Float32Array(count * 5);
  for (let i = 0; i < count; i++) {
    if (!g.index[i]) continue;
    const o = 5 * i;
    data[o] = lit[4 * i];
    data[o + 1] = lit[4 * i + 1];
    data[o + 2] = lit[4 * i + 2];
    data[o + 3] = g.index[i];
    data[o + 4] = depth[i];
  }
  return { width: W, height: H, data, slice, transmission };
}

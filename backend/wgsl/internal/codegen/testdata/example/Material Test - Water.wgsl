fn ease_quadratic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * (3.0 - 2.0 * t);
}

fn texture_mask(field: f32, edge: f32, width: f32, noiseTex: texture_3d<f32>) -> f32 {
  let t = clamp((field - edge) / width * 0.5 + 0.5, 0.0, 1.0);
  return ease_quadratic(t, noiseTex);
}

fn material_withRoughness(m: mat4x4<f32>, rough: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let surface = m[0];
  return mat4x4<f32>(vec4<f32>(rough, surface.y, surface.z, surface.w), m[1], m[2], m[3]);
}

fn ease_cubic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * t * (t * (t * 6.0 - 15.0) + 10.0);
}

fn noise_cellHash4(cell4: vec4<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  let cx = cell4.x;
  let cw = cell4.w;
  let cy = cell4.y;
  let cz = cell4.z;
  let hA = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(cx + 17.0 * cw, cy + 29.0 * cw, cz + 41.0 * cw))) & vec3<i32>(63), 0).rgb);
  let hB = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(cy + 53.0 * cw, cz + 61.0 * cw, cx + 73.0 * cw))) & vec3<i32>(63), 0).rgb);
  return vec4<f32>(hA.x, hA.y, hA.z, hB.x);
}

fn noise_cornerGradient4(cell4: vec4<f32>, noiseTex: texture_3d<f32>) -> vec4<f32> {
  return normalize(noise_cellHash4(cell4, noiseTex) * 2.0 - 1.0);
}

fn noise_gradientLattice4_gsl_anon(pi: vec4<f32>, pf: vec4<f32>, u: f32, v: f32, w: f32, t: f32, acc: f32, o: vec4<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let wx = o.x * u + (1.0 - o.x) * (1.0 - u);
  let wy = o.y * v + (1.0 - o.y) * (1.0 - v);
  let wz = o.z * w + (1.0 - o.z) * (1.0 - w);
  let wt = o.w * t + (1.0 - o.w) * (1.0 - t);
  let g = noise_cornerGradient4(pi + o, noiseTex);
  let d = pf - o;
  return acc + wx * wy * wz * wt * dot(g, d);
}

fn noise_gradientLattice4(q4: vec4<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let pi = floor(q4);
  let pf = q4 - pi;
  let u = ease_cubic(pf.x, noiseTex);
  let v = ease_cubic(pf.y, noiseTex);
  let w = ease_cubic(pf.z, noiseTex);
  let t = ease_cubic(pf.w, noiseTex);
  var gsl_fold = 0.0;
  for (var gsl_i = 0; gsl_i < 16; gsl_i = gsl_i + 1) {
    gsl_fold = noise_gradientLattice4_gsl_anon(pi, pf, u, v, w, t, gsl_fold, gsl_foldK[gsl_i], noiseTex);
  }
  return gsl_fold;
}

fn material_withDisplacement(m: mat4x4<f32>, height: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let volume = m[2];
  return mat4x4<f32>(m[0], m[1], vec4<f32>(volume.x, volume.y, height, volume.w), m[3]);
}

fn sdShape(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-25.0, -25.0, -25.0);
  let gsl_partHi = vec3<f32>(25.0, 25.0, 25.0);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = length(p) - 25.0 + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let ball = gsl_part;
  let gsl_s1 = gsl_shape(ball, 1u);
  return gsl_s1;
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  if (gsl_colorIdx == 1u) { return vec4<f32>(vec3<f32>(0.8, 0.93, 0.97), 1.0); }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32, noiseTex: texture_3d<f32>) -> gsl_material {
  if (gsl_materialIdx == 1u) {
    var gsl_simpleValue = mat4x4<f32>(vec4<f32>(0.09486833, 0.0, 1.333, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 200.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleDetail > 0.0) {
      let gsl_simpleFull = material_withDisplacement(material_withRoughness(mat4x4<f32>(vec4<f32>(0.0, 0.0, 1.333, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 200.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), sqrt(sqrt(0.0 + (1.0 - (1.0 - texture_mask(gsl_pw, 3.0, 1.0, noiseTex)) * gsl_dispDrawn) * 0.000081)), noiseTex), 0.15 * noise_gradientLattice4(vec4<f32>(p.x / 8.0, p.y / 8.0, p.z / 8.0, time / 3.0), noiseTex) * 8.0 / 10.0 * (1.0 - texture_mask(gsl_pw, 3.0, 1.0, noiseTex)), noiseTex);
      gsl_simpleValue = mat4x4<f32>(mix(gsl_simpleValue[0], gsl_simpleFull[0], gsl_simpleDetail), mix(gsl_simpleValue[1], gsl_simpleFull[1], gsl_simpleDetail), mix(gsl_simpleValue[2], gsl_simpleFull[2], gsl_simpleDetail), mix(gsl_simpleValue[3], gsl_simpleFull[3], gsl_simpleDetail));
    }
    let waterMat = gsl_simpleValue;
    let gsl_m = waterMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  return gsl_material(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), 0.0);
}
var<private> gsl_vanishEye: vec4<f32>;
var<private> gsl_vanishView: vec4<f32>;
var<private> gsl_partNear: vec3<f32>;

fn gsl_vanish(lo: vec3<f32>, hi: vec3<f32>) -> vec4<f32> {
  if (gsl_vanishEye.w <= 0.0) { return vec4<f32>(1.0, 0.0, 1e10, 1e10); }
  let size = hi - lo;
  let dist = length(gsl_vanishEye.xyz - clamp(gsl_vanishEye.xyz, lo, hi));
  let pixel = gsl_vanishEye.w * mix(max(dist, 1.0), 1.0, gsl_vanishView.x);
  let edge = min(size.x, min(size.y, size.z));
  let px = edge / pixel;
  let w = smoothstep(gsl_vanishView.y, max(gsl_vanishView.z, gsl_vanishView.y + 1e-3), px);
  return vec4<f32>(w, (1.0 - w) * 0.5 * edge, px, edge);
}

fn gsl_boxDistance(q: vec3<f32>, lo: vec3<f32>, hi: vec3<f32>) -> f32 {
  return length(max(max(lo - q, q - hi), vec3<f32>(0.0)));
}

var<private> gsl_simpleDetail: f32 = 1.0;

var<private> gsl_dispDrawn: f32 = 0.0;

const gsl_foldK = array<vec4<f32>, 16>(vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(1.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 1.0, 0.0, 0.0), vec4<f32>(1.0, 1.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 1.0, 0.0), vec4<f32>(1.0, 0.0, 1.0, 0.0), vec4<f32>(0.0, 1.0, 1.0, 0.0), vec4<f32>(1.0, 1.0, 1.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 1.0), vec4<f32>(1.0, 0.0, 0.0, 1.0), vec4<f32>(0.0, 1.0, 0.0, 1.0), vec4<f32>(1.0, 1.0, 0.0, 1.0), vec4<f32>(0.0, 0.0, 1.0, 1.0), vec4<f32>(1.0, 0.0, 1.0, 1.0), vec4<f32>(0.0, 1.0, 1.0, 1.0), vec4<f32>(1.0, 1.0, 1.0, 1.0));

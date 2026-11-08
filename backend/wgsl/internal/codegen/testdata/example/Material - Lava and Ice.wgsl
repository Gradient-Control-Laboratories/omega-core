fn gsl_floored(a: f32, n: f32) -> f32 {
  return a - n * floor(a / n);
}

fn noise_driftingCellXY(base: vec3<f32>, i: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return base + vec3<f32>(gsl_floored(i, 3.0) - 1.0, floor(i / 3.0) - 1.0, 0.0);
}

fn noise_driftingHashXY(cell: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(cell.x + 211.0, cell.y + 5.0, -97.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(cell.x + 211.0, cell.y + 5.0, -97.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
}

fn noise_driftingLife(time: f32, h: vec3<f32>, life: f32, noiseTex: texture_3d<f32>) -> f32 {
  return fract(time / life + fract(dot(h, vec3<f32>(7.1, 3.7, 5.3))));
}

fn noise_driftingStart(time: f32, h: vec3<f32>, life: f32, noiseTex: texture_3d<f32>) -> f32 {
  let grown = sin(3.1415927 * noise_driftingLife(time, h, life, noiseTex));
  return 3.0 * (1.0 - grown) * (1.0 - grown);
}

fn noise_driftingWall(q: vec3<f32>, a: vec3<f32>, b: vec3<f32>, startA: f32, startB: f32, size: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let toA = q - a;
  let lengthA = max(length(toA), 0.000001);
  let toB = q - b;
  let lengthB = max(length(toB), 0.000001);
  let gap = dot(b - a, a + b - 2.0 * q) / (lengthA + lengthB) + startB - startA;
  let u = toA / lengthA;
  let v = toB / lengthB;
  let across = cross(u, v);
  let apart = select(cross(across, u) - u * dot(across, across) / max(1.0 + dot(u, v), 0.5), v - u, dot(u, v) < 0.0);
  return gap / max(length(apart / size), 0.000001);
}

fn ease_cubic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * t * (t * (t * 6.0 - 15.0) + 10.0);
}

fn noise_cellHash3(cell: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return fract(textureLoad(noiseTex, vec3<i32>(floor(cell)) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((cell) / 64.0)) & vec3<i32>(63), 0).rgb);
}

fn noise_cornerGradient(cell: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return normalize(noise_cellHash3(cell, noiseTex) * 2.0 - 1.0);
}

fn noise_gradientLattice_gsl_anon(pi: vec3<f32>, pf: vec3<f32>, u: f32, v: f32, w: f32, acc: f32, o: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let wx = o.x * u + (1.0 - o.x) * (1.0 - u);
  let wy = o.y * v + (1.0 - o.y) * (1.0 - v);
  let wz = o.z * w + (1.0 - o.z) * (1.0 - w);
  let g = noise_cornerGradient(pi + o, noiseTex);
  let d = pf - o;
  return acc + wx * wy * wz * dot(g, d);
}

fn noise_gradientLattice(q: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let pi = floor(q);
  let pf = q - pi;
  let u = ease_cubic(pf.x, noiseTex);
  let v = ease_cubic(pf.y, noiseTex);
  let w = ease_cubic(pf.z, noiseTex);
  var gsl_fold = 0.0;
  for (var gsl_i = 0; gsl_i < 8; gsl_i = gsl_i + 1) {
    gsl_fold = noise_gradientLattice_gsl_anon(pi, pf, u, v, w, gsl_fold, gsl_foldK[gsl_i], noiseTex);
  }
  return gsl_fold;
}

fn ease_quadratic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * (3.0 - 2.0 * t);
}

fn texture_fadeDetail(size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 1.0 - texture_mask(gsl_pw, 0.375 * size, 0.125 * size, noiseTex);
}

fn texture_reliefRoughness(roughness: f32, slope: f32, size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return sqrt(roughness * roughness + (1.0 - texture_fadeDetail(size, noiseTex) * gsl_dispDrawn) * slope * slope);
}

fn material_make(rough: f32, metal: f32, index: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return mat4x4<f32>(vec4<f32>(rough, metal, index, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
}

fn material_withEmissive(m: mat4x4<f32>, rgb: vec3<f32>, strength: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return mat4x4<f32>(m[0], vec4<f32>(rgb.x, rgb.y, rgb.z, strength), m[2], m[3]);
}

fn color_srgbToLinearChannel(s: f32, noiseTex: texture_3d<f32>) -> f32 {
  return select(pow((s + 0.055) / 1.055, 2.4), s / 12.92, s <= 0.04045);
}

fn color_linearize(c: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return vec3<f32>(color_srgbToLinearChannel(c.x, noiseTex), color_srgbToLinearChannel(c.y, noiseTex), color_srgbToLinearChannel(c.z, noiseTex));
}

fn color_cbrt(x: f32, noiseTex: texture_3d<f32>) -> f32 {
  return sign(x) * pow(abs(x), 0.33333334);
}

fn color_rgb2oklab(c: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let lin = color_linearize(c, noiseTex);
  let l = dot(vec3<f32>(0.41222146, 0.53633255, 0.051445995), lin);
  let l_ = color_cbrt(l, noiseTex);
  let m = dot(vec3<f32>(0.2119035, 0.6806995, 0.10739696), lin);
  let m_ = color_cbrt(m, noiseTex);
  let s = dot(vec3<f32>(0.08830246, 0.28171885, 0.6299787), lin);
  let s_ = color_cbrt(s, noiseTex);
  let lms = vec3<f32>(l_, m_, s_);
  return vec3<f32>(dot(vec3<f32>(0.21045426, 0.7936178, -0.004072047), lms), dot(vec3<f32>(1.9779985, -2.4285922, 0.4505937), lms), dot(vec3<f32>(0.025904037, 0.78277177, -0.80867577), lms));
}

fn color_linearToSrgbChannel(l: f32, noiseTex: texture_3d<f32>) -> f32 {
  return select(1.055 * pow(l, 0.41666666) - 0.055, l * 12.92, l <= 0.0031308);
}

fn color_gammaEncode(c: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return vec3<f32>(color_linearToSrgbChannel(c.x, noiseTex), color_linearToSrgbChannel(c.y, noiseTex), color_linearToSrgbChannel(c.z, noiseTex));
}

fn color_oklab2rgb(c: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let l_ = c.x + 0.39633778 * c.y + 0.21580376 * c.z;
  let l = l_ * l_ * l_;
  let m_ = c.x - 0.105561346 * c.y - 0.06385417 * c.z;
  let m = m_ * m_ * m_;
  let s_ = c.x - 0.08948418 * c.y - 1.2914855 * c.z;
  let s = s_ * s_ * s_;
  let lms = vec3<f32>(l, m, s);
  let lin = vec3<f32>(dot(vec3<f32>(4.0767417, -3.3077116, 0.23096994), lms), dot(vec3<f32>(-1.268438, 2.6097574, -0.34131938), lms), dot(vec3<f32>(-0.0041960864, -0.7034186, 1.7076147), lms));
  return color_gammaEncode(lin, noiseTex);
}

fn color_mixOklab(a: vec3<f32>, b: vec3<f32>, t: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let la = color_rgb2oklab(a, noiseTex);
  let lb = color_rgb2oklab(b, noiseTex);
  return color_oklab2rgb(la + (lb - la) * t, noiseTex);
}

fn material_withDisplacement(m: mat4x4<f32>, height: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let volume = m[2];
  return mat4x4<f32>(m[0], m[1], vec4<f32>(volume.x, volume.y, height, volume.w), m[3]);
}

fn texture_coverageLine(gsl_var_distance: f32, width: f32, noiseTex: texture_3d<f32>) -> f32 {
  let half = max(0.5 * width, 0.5 * gsl_pw);
  return (1.0 - texture_mask(gsl_var_distance, half, 0.25 * half, noiseTex)) * 0.5 * width / half;
}

fn material_blend(a: mat4x4<f32>, b: mat4x4<f32>, t: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return mat4x4<f32>(a[0] + (b[0] - a[0]) * t, a[1] + (b[1] - a[1]) * t, a[2] + (b[2] - a[2]) * t, a[3] + (b[3] - a[3]) * t);
}

fn builtin_intersectSharp(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn builtin_intersectRound(fields: vec3<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return min(-radius, builtin_intersectSharp(fields, noiseTex)) + sqrt(sumOfSquares);
}

fn builtin_gcl_closure_intersectRound(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return builtin_intersectRound(fields, 0.0, noiseTex);
}

fn builtin_gcl_closure_boxCentered(p: vec3<f32>, center: vec3<f32>, size: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  let halfSize = size * 0.5;
  let d = abs(p - center) - halfSize;
  let solid = builtin_gcl_closure_intersectRound(vec3<f32>(d.x, d.y, d.z), noiseTex);
  return solid;
}

fn texture_mask(field: f32, edge: f32, width: f32, noiseTex: texture_3d<f32>) -> f32 {
  let t = clamp((field - edge) / width * 0.5 + 0.5, 0.0, 1.0);
  return ease_quadratic(t, noiseTex);
}

fn animated_incandescentStrength(celsius: f32, noiseTex: texture_3d<f32>) -> f32 {
  return exp(16.483099 - 23870.0 / (max(celsius, 0.0) + 273.15));
}

fn animated_incandescentColor(celsius: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return vec3<f32>(1.0, max(0.43 * (1.0 - exp(-(celsius - 560.0) / 280.0)), 0.0), 0.0);
}

fn animated_iceStage(stage: f32, values: array<f32, 7>, noiseTex: texture_3d<f32>) -> f32 {
  let s = clamp(stage, 0.0, 1.0);
  return select(select(select(select(select(select(values[0], values[0] + (values[1] - values[0]) * clamp((s - 0.0) / 0.2, 0.0, 1.0), s > 0.0), values[1] + (values[2] - values[1]) * clamp((s - 0.2) / 0.15, 0.0, 1.0), s > 0.2), values[2] + (values[3] - values[2]) * clamp((s - 0.35) / 0.15, 0.0, 1.0), s > 0.35), values[3] + (values[4] - values[3]) * clamp((s - 0.5) / 0.15, 0.0, 1.0), s > 0.5), values[4] + (values[5] - values[4]) * clamp((s - 0.65) / 0.15, 0.0, 1.0), s > 0.65), values[5] + (values[6] - values[5]) * clamp((s - 0.8) / 0.2, 0.0, 1.0), s > 0.8);
}

fn gcl_closure_gsl_anon(wind: vec3<f32>, x: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return wind * 0.15;
}

fn gcl_closure_driftingSeedXY_2(wind: vec3<f32>, time: f32, cell: vec3<f32>, h: vec3<f32>, size: vec3<f32>, life: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let born = vec3<f32>(cell.x + 0.25 + 0.5 * h.x, cell.y + 0.25 + 0.5 * h.y, 0.0);
  let flow = gcl_closure_gsl_anon(wind, born * size, noiseTex) / size * noise_driftingLife(time, h, life, noiseTex) * life;
  return born + vec3<f32>(flow.x, flow.y, 0.0);
}

fn gcl_closure_driftingCellsLatticeXY_gsl_anon(base: vec3<f32>, flat: vec3<f32>, size: vec3<f32>, life: f32, wind: vec3<f32>, time: f32, best: vec2<f32>, i: f32, noiseTex: texture_3d<f32>) -> vec2<f32> {
  let cell = noise_driftingCellXY(base, i, noiseTex);
  let h = noise_driftingHashXY(cell, noiseTex);
  let arrival = length(flat - gcl_closure_driftingSeedXY_2(wind, time, cell, h, size, life, noiseTex)) + noise_driftingStart(time, h, life, noiseTex);
  return select(best, vec2<f32>(arrival, i), arrival < best.x);
}

fn gcl_closure_driftingSeedXY(wind: vec3<f32>, time: f32, cell: vec3<f32>, h: vec3<f32>, size: vec3<f32>, life: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let born = vec3<f32>(cell.x + 0.25 + 0.5 * h.x, cell.y + 0.25 + 0.5 * h.y, 0.0);
  let flow = gcl_closure_gsl_anon(wind, born * size, noiseTex) / size * noise_driftingLife(time, h, life, noiseTex) * life;
  return born + vec3<f32>(flow.x, flow.y, 0.0);
}

fn gcl_closure_driftingSeedXY_3(wind: vec3<f32>, time: f32, cell: vec3<f32>, h: vec3<f32>, size: vec3<f32>, life: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let born = vec3<f32>(cell.x + 0.25 + 0.5 * h.x, cell.y + 0.25 + 0.5 * h.y, 0.0);
  let flow = gcl_closure_gsl_anon(wind, born * size, noiseTex) / size * noise_driftingLife(time, h, life, noiseTex) * life;
  return born + vec3<f32>(flow.x, flow.y, 0.0);
}

fn gcl_closure_driftingCellsLatticeXY_gsl_anon_2(first: vec3<f32>, flat: vec3<f32>, firstSeed: vec3<f32>, size: vec3<f32>, life: f32, wind: vec3<f32>, time: f32, firstStart: f32, firstAge: f32, nearest: vec2<f32>, i: f32, noiseTex: texture_3d<f32>) -> vec2<f32> {
  let cell = noise_driftingCellXY(first, i, noiseTex);
  let h = noise_driftingHashXY(cell, noiseTex);
  let toWall = noise_driftingWall(flat, firstSeed, gcl_closure_driftingSeedXY_3(wind, time, cell, h, size, life, noiseTex), firstStart, noise_driftingStart(time, h, life, noiseTex), size, noiseTex);
  return select(nearest, vec2<f32>(toWall, min(firstAge, noise_driftingLife(time, h, life, noiseTex) * life)), i != 4.0 && toWall < nearest.x);
}

fn gcl_closure_driftingCellsLatticeXY(wind: vec3<f32>, time: f32, q: vec3<f32>, size: vec3<f32>, life: f32, noiseTex: texture_3d<f32>) -> gsl_struct {
  let flat = vec3<f32>(q.x, q.y, 0.0);
  let base = floor(flat);
  var gsl_fold = vec2<f32>(1000.0, 0.0);
  for (var gsl_i = 0; gsl_i < 9; gsl_i = gsl_i + 1) {
    gsl_fold = gcl_closure_driftingCellsLatticeXY_gsl_anon(base, flat, size, life, wind, time, gsl_fold, f32(gsl_i), noiseTex);
  }
  let owner = gsl_fold;
  let first = noise_driftingCellXY(base, owner.y, noiseTex);
  let firstHash = noise_driftingHashXY(first, noiseTex);
  let firstSeed = gcl_closure_driftingSeedXY(wind, time, first, firstHash, size, life, noiseTex);
  let firstStart = noise_driftingStart(time, firstHash, life, noiseTex);
  let firstAge = noise_driftingLife(time, firstHash, life, noiseTex) * life;
  var gsl_fold_2 = vec2<f32>(1000.0, 0.0);
  for (var gsl_i_2 = 0; gsl_i_2 < 9; gsl_i_2 = gsl_i_2 + 1) {
    gsl_fold_2 = gcl_closure_driftingCellsLatticeXY_gsl_anon_2(first, flat, firstSeed, size, life, wind, time, firstStart, firstAge, gsl_fold_2, f32(gsl_i_2), noiseTex);
  }
  let wall = gsl_fold_2;
  let drift = (firstSeed - vec3<f32>(first.x + 0.25 + 0.5 * firstHash.x, first.y + 0.25 + 0.5 * firstHash.y, 0.0)) * size;
  var gsl_result : gsl_struct;
  gsl_result.v_1 = vec4<f32>(wall.x, wall.y, firstAge, fract(dot(firstHash, vec3<f32>(3.1, 7.7, 5.9))));
  gsl_result.v_2 = drift;
  return gsl_result;
}

fn gcl_closure_driftingSeedXY_5(wind: vec3<f32>, time: f32, cell: vec3<f32>, h: vec3<f32>, size: vec3<f32>, life: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let born = vec3<f32>(cell.x + 0.25 + 0.5 * h.x, cell.y + 0.25 + 0.5 * h.y, 0.0);
  let flow = wind * 0.02 / size * noise_driftingLife(time, h, life, noiseTex) * life;
  return born + vec3<f32>(flow.x, flow.y, 0.0);
}

fn gcl_closure_driftingCellsLatticeXY_2_gsl_anon(base: vec3<f32>, flat: vec3<f32>, size: vec3<f32>, life: f32, wind: vec3<f32>, time: f32, best: vec2<f32>, i: f32, noiseTex: texture_3d<f32>) -> vec2<f32> {
  let cell = noise_driftingCellXY(base, i, noiseTex);
  let h = noise_driftingHashXY(cell, noiseTex);
  let arrival = length(flat - gcl_closure_driftingSeedXY_5(wind, time, cell, h, size, life, noiseTex)) + noise_driftingStart(time, h, life, noiseTex);
  return select(best, vec2<f32>(arrival, i), arrival < best.x);
}

fn gcl_closure_driftingSeedXY_4(wind: vec3<f32>, time: f32, cell: vec3<f32>, h: vec3<f32>, size: vec3<f32>, life: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let born = vec3<f32>(cell.x + 0.25 + 0.5 * h.x, cell.y + 0.25 + 0.5 * h.y, 0.0);
  let flow = wind * 0.02 / size * noise_driftingLife(time, h, life, noiseTex) * life;
  return born + vec3<f32>(flow.x, flow.y, 0.0);
}

fn gcl_closure_driftingSeedXY_6(wind: vec3<f32>, time: f32, cell: vec3<f32>, h: vec3<f32>, size: vec3<f32>, life: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let born = vec3<f32>(cell.x + 0.25 + 0.5 * h.x, cell.y + 0.25 + 0.5 * h.y, 0.0);
  let flow = wind * 0.02 / size * noise_driftingLife(time, h, life, noiseTex) * life;
  return born + vec3<f32>(flow.x, flow.y, 0.0);
}

fn gcl_closure_driftingCellsLatticeXY_2_gsl_anon_2(first: vec3<f32>, flat: vec3<f32>, firstSeed: vec3<f32>, size: vec3<f32>, life: f32, wind: vec3<f32>, time: f32, firstStart: f32, firstAge: f32, nearest: vec2<f32>, i: f32, noiseTex: texture_3d<f32>) -> vec2<f32> {
  let cell = noise_driftingCellXY(first, i, noiseTex);
  let h = noise_driftingHashXY(cell, noiseTex);
  let toWall = noise_driftingWall(flat, firstSeed, gcl_closure_driftingSeedXY_6(wind, time, cell, h, size, life, noiseTex), firstStart, noise_driftingStart(time, h, life, noiseTex), size, noiseTex);
  return select(nearest, vec2<f32>(toWall, min(firstAge, noise_driftingLife(time, h, life, noiseTex) * life)), i != 4.0 && toWall < nearest.x);
}

fn gcl_closure_driftingCellsLatticeXY_2(wind: vec3<f32>, time: f32, q: vec3<f32>, size: vec3<f32>, life: f32, noiseTex: texture_3d<f32>) -> gsl_struct {
  let flat = vec3<f32>(q.x, q.y, 0.0);
  let base = floor(flat);
  var gsl_fold = vec2<f32>(1000.0, 0.0);
  for (var gsl_i = 0; gsl_i < 9; gsl_i = gsl_i + 1) {
    gsl_fold = gcl_closure_driftingCellsLatticeXY_2_gsl_anon(base, flat, size, life, wind, time, gsl_fold, f32(gsl_i), noiseTex);
  }
  let owner = gsl_fold;
  let first = noise_driftingCellXY(base, owner.y, noiseTex);
  let firstHash = noise_driftingHashXY(first, noiseTex);
  let firstSeed = gcl_closure_driftingSeedXY_4(wind, time, first, firstHash, size, life, noiseTex);
  let firstStart = noise_driftingStart(time, firstHash, life, noiseTex);
  let firstAge = noise_driftingLife(time, firstHash, life, noiseTex) * life;
  var gsl_fold_2 = vec2<f32>(1000.0, 0.0);
  for (var gsl_i_2 = 0; gsl_i_2 < 9; gsl_i_2 = gsl_i_2 + 1) {
    gsl_fold_2 = gcl_closure_driftingCellsLatticeXY_2_gsl_anon_2(first, flat, firstSeed, size, life, wind, time, firstStart, firstAge, gsl_fold_2, f32(gsl_i_2), noiseTex);
  }
  let wall = gsl_fold_2;
  let drift = (firstSeed - vec3<f32>(first.x + 0.25 + 0.5 * firstHash.x, first.y + 0.25 + 0.5 * firstHash.y, 0.0)) * size;
  var gsl_result : gsl_struct;
  gsl_result.v_1 = vec4<f32>(wall.x, wall.y, firstAge, fract(dot(firstHash, vec3<f32>(3.1, 7.7, 5.9))));
  gsl_result.v_2 = drift;
  return gsl_result;
}

fn gcl_closure_iceFloes(p: vec3<f32>, wind: vec3<f32>, time: f32, floeSize: f32, life: f32, stage: f32, openWater: f32, noiseTex: texture_3d<f32>) -> gsl_struct_2 {
  let roughness = animated_iceStage(stage, array<f32, 7>(0.6, 0.08, 0.12, 0.5, 0.5, 0.6, 0.9), noiseTex);
  let leadHalf = floeSize * (0.02 + 0.2 * openWater);
  let x = leadHalf / floeSize;
  let covered = exp(-2.9 * x - 2.5 * x * x);
  let pancake = 1.0 - clamp(abs(stage - 0.5) / 0.15, 0.0, 1.0);
  let gsl_result_2 = gcl_closure_driftingCellsLatticeXY_2(wind, time, p / floeSize, vec3<f32>(floeSize, floeSize, floeSize), life, noiseTex);
  let cells = gsl_result_2.v_1;
  let onFloe = 1.0 - texture_coverageLine(cells.x, 2.0 * leadHalf, noiseTex);
  let drift = gsl_result_2.v_2;
  let snow = noise_gradientLattice((p - drift) / (0.2 * floeSize) + vec3<f32>(19.0, 43.0, 7.0), noiseTex);
  let rimWidth = 0.06 * floeSize;
  let rim = pancake * exp(-((cells.x - leadHalf - rimWidth) / rimWidth) * (cells.x - leadHalf - rimWidth) / rimWidth);
  var gsl_simpleValue = material_withDisplacement(material_blend(mat4x4<f32>(vec4<f32>(0.06, 0.0, 1.33, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), material_make(roughness, 0.0, 1.31, noiseTex), covered, noiseTex), 0.03 * floeSize * covered * (1.0 + 0.285 * pancake), noiseTex);
  if (gsl_simpleDetail > 0.0) {
    let mat = material_blend(mat4x4<f32>(vec4<f32>(0.06, 0.0, 1.33, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), material_make(roughness, 0.0, 1.31, noiseTex), onFloe, noiseTex);
    let relief = onFloe * (0.03 * floeSize * (1.0 + 0.3 * snow * clamp(stage * 2.0 - 1.0, 0.0, 1.0)) + 0.03 * floeSize * rim);
    let gsl_simpleFull = material_withDisplacement(mat, relief, noiseTex);
    gsl_simpleValue = mat4x4<f32>(mix(gsl_simpleValue[0], gsl_simpleFull[0], gsl_simpleDetail), mix(gsl_simpleValue[1], gsl_simpleFull[1], gsl_simpleDetail), mix(gsl_simpleValue[2], gsl_simpleFull[2], gsl_simpleDetail), mix(gsl_simpleValue[3], gsl_simpleFull[3], gsl_simpleDetail));
  }
  let pack = gsl_simpleValue;
  let albedo = animated_iceStage(stage, array<f32, 7>(0.1, 0.15, 0.25, 0.4, 0.35, 0.6, 0.85), noiseTex);
  let flat = pow(clamp(albedo + (0.8 - albedo) * 0.285 * pancake, 0.0, 1.0), 0.45454547);
  var gsl_simpleValue_2 = color_mixOklab(vec3<f32>(0.02, 0.05, 0.08), color_mixOklab(vec3<f32>(flat * 0.9, flat * 0.96, flat), vec3<f32>(flat, flat, flat), clamp(stage * 2.0 - 1.0, 0.0, 1.0), noiseTex), covered, noiseTex);
  if (gsl_simpleDetail > 0.0) {
    let tone = pow(clamp(albedo + (0.8 - albedo) * rim + 0.04 * snow, 0.0, 1.0), 0.45454547);
    let ice = color_mixOklab(vec3<f32>(tone * 0.9, tone * 0.96, tone), vec3<f32>(tone, tone, tone), clamp(stage * 2.0 - 1.0, 0.0, 1.0), noiseTex);
    gsl_simpleValue_2 = mix(gsl_simpleValue_2, color_mixOklab(vec3<f32>(0.02, 0.05, 0.08), ice, onFloe, noiseTex), gsl_simpleDetail);
  }
  let floes = gsl_simpleValue_2;
  var gsl_result : gsl_struct_2;
  gsl_result.v_1 = pack;
  gsl_result.v_2 = floes;
  gsl_result.v_3 = 0.0;
  gsl_result.v_4 = 0.0;
  return gsl_result;
}

fn gsl_shapeParts(p: vec3<f32>, time: f32, gsl_param_12_13: f32, gsl_param_19_12: f32, noiseTex: texture_3d<f32>) -> array<f32, 2> {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-28.0, -12.0, -4.0);
  let gsl_partHi = vec3<f32>(-4.0, 12.0, 4.0);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = builtin_gcl_closure_boxCentered(p - vec3<f32>(-16.0, -0.0, 0.0), vec3<f32>(0.0, 0.0, 0.0) * 0.5, abs(vec3<f32>(24.0, 24.0, 8.0)), noiseTex) - 0.0 + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let lava = gsl_part;
  let windAngle = gsl_param_12_13;
  let wind = vec3<f32>(cos(windAngle * 3.1415927 / 180.0), sin(windAngle * 3.1415927 / 180.0), 0.0);
  let iceStage = gsl_param_19_12;
  let gsl_result_3 = gcl_closure_iceFloes(p, wind, time, 8.0, 120.0, iceStage, 0.3, noiseTex);
  let iceHeight = gsl_result_3.v_3;
  let ice = builtin_gcl_closure_boxCentered(p - vec3<f32>(16.0, -0.0, 0.0), vec3<f32>(0.0, 0.0, 0.0) * 0.5, abs(vec3<f32>(24.0, 24.0, 8.0)), noiseTex) - iceHeight;
  return array<f32, 2>(lava, ice);
}

fn sdShape(p: vec3<f32>, time: f32, gsl_param_12_13: f32, gsl_param_19_12: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  let gsl_d = gsl_shapeParts(p, time, gsl_param_12_13, gsl_param_19_12, noiseTex);
  let gsl_s1 = gsl_shape(gsl_d[0], 1u);
  let gsl_s2 = gsl_shape(gsl_d[1], 2u);
  return gsl_min(gsl_s1, gsl_s2);
}
fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32, gsl_param_12_13: f32, gsl_param_19_12: f32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  if (gsl_colorIdx == 1u) {
    var gsl_simpleValue = vec3<f32>(0.079512164, 0.072891675, 0.069725566);
    if (gsl_simpleDetail > 0.0) {
      let windAngle = gsl_param_12_13;
      let wind = vec3<f32>(cos(windAngle * 3.1415927 / 180.0), sin(windAngle * 3.1415927 / 180.0), 0.0);
      let gsl_result = gcl_closure_driftingCellsLatticeXY(wind, time, p / 8.0, vec3<f32>(8.0, 8.0, 8.0), 20.0, noiseTex);
      let cells = gsl_result.v_1;
      let drift = gsl_result.v_2;
      gsl_simpleValue = mix(gsl_simpleValue, color_mixOklab(vec3<f32>(0.05, 0.045, 0.045), vec3<f32>(0.14, 0.13, 0.12), texture_mask(cells.z / 20.0 + 0.3 * noise_gradientLattice((p - drift) / 1.2 + vec3<f32>(7.0, 31.0, 13.0), noiseTex), 0.6, 0.3, noiseTex), noiseTex), gsl_simpleDetail);
    }
    let lavaColor = gsl_simpleValue;
    return vec4<f32>(lavaColor, 1.0);
  }
  if (gsl_colorIdx == 2u) {
    let windAngle = gsl_param_12_13;
    let wind = vec3<f32>(cos(windAngle * 3.1415927 / 180.0), sin(windAngle * 3.1415927 / 180.0), 0.0);
    let iceStage = gsl_param_19_12;
    let gsl_result_3 = gcl_closure_iceFloes(p, wind, time, 8.0, 120.0, iceStage, 0.3, noiseTex);
    let iceColor = gsl_result_3.v_2;
    return vec4<f32>(iceColor, 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32, gsl_param_12_13: f32, gsl_param_19_12: f32, noiseTex: texture_3d<f32>) -> gsl_material {
  if (gsl_materialIdx == 1u) {
    var gsl_simpleValue_2 = mat4x4<f32>(vec4<f32>(0.494, 0.0, 1.58, 0.0), vec4<f32>(1.0, 0.026, 0.0, 0.146), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    if (gsl_simpleDetail > 0.0) {
      let windAngle = gsl_param_12_13;
      let wind = vec3<f32>(cos(windAngle * 3.1415927 / 180.0), sin(windAngle * 3.1415927 / 180.0), 0.0);
      let gsl_result = gcl_closure_driftingCellsLatticeXY(wind, time, p / 8.0, vec3<f32>(8.0, 8.0, 8.0), 20.0, noiseTex);
      let cells = gsl_result.v_1;
      let drift = gsl_result.v_2;
      let gsl_simpleFull = material_withDisplacement(material_withEmissive(material_make(sqrt(texture_reliefRoughness((0.25 + 0.3 * texture_mask(cells.z / 20.0 + 0.3 * noise_gradientLattice((p - drift) / 1.2 + vec3<f32>(7.0, 31.0, 13.0), noiseTex), 0.6, 0.3, noiseTex)) * (0.25 + 0.3 * texture_mask(cells.z / 20.0 + 0.3 * noise_gradientLattice((p - drift) / 1.2 + vec3<f32>(7.0, 31.0, 13.0), noiseTex), 0.6, 0.3, noiseTex)), 0.2, 1.2, noiseTex)), 0.0, 1.58, noiseTex), animated_incandescentColor(max(400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388) + (1050.0 + 125.0 * exp(-cells.y / 3.0) - (400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388))) * exp(-(cells.x / max(8.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw)) * cells.x / max(8.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw)), 400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388)), noiseTex), 20.0 * (animated_incandescentStrength(400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388), noiseTex) + (animated_incandescentStrength(max(400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388) + (1050.0 + 125.0 * exp(-cells.y / 3.0) - (400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388))) * exp(-(cells.x / max(8.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw)) * cells.x / max(8.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw)), 400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388)), noiseTex) - animated_incandescentStrength(400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388), noiseTex)) * 8.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)) / max(8.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw)), noiseTex), (-0.12 * exp(-(cells.x / max(8.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw)) * cells.x / max(8.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw)) * 8.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)) / max(8.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw) + 0.08 * noise_gradientLattice((p - drift) / 1.2 + vec3<f32>(7.0, 31.0, 13.0), noiseTex)) * (1.0 - texture_mask(gsl_pw, 0.45, 0.15, noiseTex)), noiseTex);
      gsl_simpleValue_2 = mat4x4<f32>(mix(gsl_simpleValue_2[0], gsl_simpleFull[0], gsl_simpleDetail), mix(gsl_simpleValue_2[1], gsl_simpleFull[1], gsl_simpleDetail), mix(gsl_simpleValue_2[2], gsl_simpleFull[2], gsl_simpleDetail), mix(gsl_simpleValue_2[3], gsl_simpleFull[3], gsl_simpleDetail));
    }
    let lavaMat = gsl_simpleValue_2;
    let gsl_m = lavaMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 2u) {
    let windAngle = gsl_param_12_13;
    let wind = vec3<f32>(cos(windAngle * 3.1415927 / 180.0), sin(windAngle * 3.1415927 / 180.0), 0.0);
    let iceStage = gsl_param_19_12;
    let gsl_result_3 = gcl_closure_iceFloes(p, wind, time, 8.0, 120.0, iceStage, 0.3, noiseTex);
    let iceMat = gsl_result_3.v_1;
    let gsl_m = iceMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
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

const gsl_foldK = array<vec3<f32>, 8>(vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 0.0, 0.0), vec3<f32>(0.0, 1.0, 0.0), vec3<f32>(1.0, 1.0, 0.0), vec3<f32>(0.0, 0.0, 1.0), vec3<f32>(1.0, 0.0, 1.0), vec3<f32>(0.0, 1.0, 1.0), vec3<f32>(1.0, 1.0, 1.0));

fn gsl_heights(p: vec3<f32>, time: f32, gsl_param_12_13: f32, gsl_param_19_12: f32, noiseTex: texture_3d<f32>) -> array<f32, 2> {
  let windAngle = gsl_param_12_13;
  let wind = vec3<f32>(cos(windAngle * 3.1415927 / 180.0), sin(windAngle * 3.1415927 / 180.0), 0.0);
  let gsl_result = gcl_closure_driftingCellsLatticeXY(wind, time, p / 8.0, vec3<f32>(8.0, 8.0, 8.0), 20.0, noiseTex);
  let cells = gsl_result.v_1;
  let drift = gsl_result.v_2;
  var gsl_simpleValue_3 = mat4x4<f32>(vec4<f32>(0.494, 0.0, 1.58, 0.0), vec4<f32>(1.0, 0.026, 0.0, 0.146), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleDetail > 0.0) {
    let gsl_simpleFull_2 = material_withDisplacement(material_withEmissive(material_make(sqrt(texture_reliefRoughness((0.25 + 0.3 * texture_mask(cells.z / 20.0 + 0.3 * noise_gradientLattice((p - drift) / 1.2 + vec3<f32>(7.0, 31.0, 13.0), noiseTex), 0.6, 0.3, noiseTex)) * (0.25 + 0.3 * texture_mask(cells.z / 20.0 + 0.3 * noise_gradientLattice((p - drift) / 1.2 + vec3<f32>(7.0, 31.0, 13.0), noiseTex), 0.6, 0.3, noiseTex)), 0.2, 1.2, noiseTex)), 0.0, 1.58, noiseTex), animated_incandescentColor(max(400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388) + (1050.0 + 125.0 * exp(-cells.y / 3.0) - (400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388))) * exp(-(cells.x / max(8.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw)) * cells.x / max(8.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw)), 400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388)), noiseTex), 20.0 * (animated_incandescentStrength(400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388), noiseTex) + (animated_incandescentStrength(max(400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388) + (1050.0 + 125.0 * exp(-cells.y / 3.0) - (400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388))) * exp(-(cells.x / max(8.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw)) * cells.x / max(8.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw)), 400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388)), noiseTex) - animated_incandescentStrength(400.0 + 775.0 / (1.0 + sqrt((cells.z + 0.25) / 0.1) - 1.5811388), noiseTex)) * 8.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)) / max(8.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw)), noiseTex), (-0.12 * exp(-(cells.x / max(8.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw)) * cells.x / max(8.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw)) * 8.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)) / max(8.0 * (0.025 + 0.05 * exp(-cells.y / 3.0)), 0.5 * gsl_pw) + 0.08 * noise_gradientLattice((p - drift) / 1.2 + vec3<f32>(7.0, 31.0, 13.0), noiseTex)) * (1.0 - texture_mask(gsl_pw, 0.45, 0.15, noiseTex)), noiseTex);
    gsl_simpleValue_3 = mat4x4<f32>(mix(gsl_simpleValue_3[0], gsl_simpleFull_2[0], gsl_simpleDetail), mix(gsl_simpleValue_3[1], gsl_simpleFull_2[1], gsl_simpleDetail), mix(gsl_simpleValue_3[2], gsl_simpleFull_2[2], gsl_simpleDetail), mix(gsl_simpleValue_3[3], gsl_simpleFull_2[3], gsl_simpleDetail));
  }
  let lavaMat = gsl_simpleValue_3;
  let iceStage = gsl_param_19_12;
  let gsl_result_3 = gcl_closure_iceFloes(p, wind, time, 8.0, 120.0, iceStage, 0.3, noiseTex);
  let iceMat = gsl_result_3.v_1;
  return array<f32, 2>(lavaMat[2].z, iceMat[2].z);
}

fn gsl_sdShapeDisplaced(p: vec3<f32>, time: f32, gsl_dispK: f32, gsl_param_12_13: f32, gsl_param_19_12: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  let gsl_d = gsl_shapeParts(p, time, gsl_param_12_13, gsl_param_19_12, noiseTex);
  let gsl_h = gsl_heights(p, time, gsl_param_12_13, gsl_param_19_12, noiseTex);
  let gsl_s1 = gsl_shape(gsl_d[0] - gsl_h[0] * gsl_dispK, 1u);
  let gsl_s2 = gsl_shape(gsl_d[1] - gsl_h[1] * gsl_dispK, 2u);
  return gsl_min(gsl_s1, gsl_s2);
}

struct gsl_struct {
  v_1 : vec4<f32>,
  v_2 : vec3<f32>,
}
struct gsl_struct_2 {
  v_1 : mat4x4<f32>,
  v_2 : vec3<f32>,
  v_3 : f32,
  v_4 : f32,
}

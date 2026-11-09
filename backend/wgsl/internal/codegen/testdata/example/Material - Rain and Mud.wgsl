fn gsl_floored(a: f32, n: f32) -> f32 {
  return a - n * floor(a / n);
}

fn ease_quadratic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * (3.0 - 2.0 * t);
}

fn texture_mask(field: f32, edge: f32, width: f32, noiseTex: texture_3d<f32>) -> f32 {
  let t = clamp((field - edge) / width * 0.5 + 0.5, 0.0, 1.0);
  return ease_quadratic(t, noiseTex);
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

fn material_withDisplacement(m: mat4x4<f32>, height: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  let volume = m[2];
  return mat4x4<f32>(m[0], m[1], vec4<f32>(volume.x, volume.y, height, volume.w), m[3]);
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

fn material_blend(a: mat4x4<f32>, b: mat4x4<f32>, t: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return mat4x4<f32>(a[0] + (b[0] - a[0]) * t, a[1] + (b[1] - a[1]) * t, a[2] + (b[2] - a[2]) * t, a[3] + (b[3] - a[3]) * t);
}

fn builtin_intersectSharp(fields: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return max(max(fields.x, fields.y), fields.z);
}

fn builtin_intersectRound(fields: vec3<f32>, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  let sharp = builtin_intersectSharp(fields, noiseTex);
  let sumOfSquares = 0.0 + max(fields.x + radius, 0.0) * max(fields.x + radius, 0.0) + max(fields.y + radius, 0.0) * max(fields.y + radius, 0.0) + max(fields.z + radius, 0.0) * max(fields.z + radius, 0.0);
  return sharp - max(sharp + radius, 0.0) + sqrt(max(sumOfSquares, 0.000000000000000000000000000001));
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

fn animated_rainRing(r: f32, age: f32, dropSize: f32, noiseTex: texture_3d<f32>) -> f32 {
  let a = max(age, 0.001);
  let k = r / (1.5 * a) * r / (1.5 * a) / 72800.0;
  let amplitude = 0.1 * dropSize * dropSize / (dropSize + r) * exp(-2.0 * k * k * a);
  let calm = texture_mask(r / a, 177.0, 30.0, noiseTex);
  return amplitude * calm * cos(r * r * r / (491400.0 * a * a));
}

fn animated_rain_gsl_anon(p: vec3<f32>, base: vec3<f32>, time: f32, period: f32, cellSize: f32, dropSize: f32, acc: f32, i: f32, noiseTex: texture_3d<f32>) -> f32 {
  let cell = base + vec3<f32>(gsl_floored(i, 5.0) - 2.0, floor(i / 5.0) - 2.0, 0.0);
  let h = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(cell.x + 71.0, cell.y + 13.0, 0.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(cell.x + 71.0, cell.y + 13.0, 0.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let cycles = time / period + h.x;
  let drop = floor(cycles);
  let spot = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(cell.x + 0.37 * drop + 5.0, cell.y + 0.11 * drop + 29.0, cell.y + 17.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(cell.x + 0.37 * drop + 5.0, cell.y + 0.11 * drop + 29.0, cell.y + 17.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let x = p.x - (cell.x + spot.x) * cellSize;
  let y = p.y - (cell.y + spot.y) * cellSize;
  let r = sqrt(x * x + y * y);
  let fade = 1.0 - texture_mask(r, 48.0, 9.0, noiseTex);
  return acc + fade * animated_rainRing(r, fract(cycles) * period, dropSize, noiseTex);
}

fn animated_rain(p: vec3<f32>, time: f32, intensity: f32, dropSize: f32, noiseTex: texture_3d<f32>) -> gsl_struct {
  let base = floor(vec3<f32>(p.x / 30.0, p.y / 30.0, 0.0));
  let period = 1.0 / (0.18 * intensity);
  let cellSize = 30.0;
  var gsl_fold = 0.0;
  for (var gsl_i = 0; gsl_i < 25; gsl_i = gsl_i + 1) {
    gsl_fold = animated_rain_gsl_anon(p, base, time, period, cellSize, dropSize, gsl_fold, f32(gsl_i), noiseTex);
  }
  var gsl_simpleValue = mat4x4<f32>(vec4<f32>(0.18, 0.0, 1.33, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
  if (gsl_simpleDetail > 0.0) {
    let relief = gsl_fold;
    let ringSize = 3.0 * dropSize;
    let water = material_make(sqrt(texture_reliefRoughness(0.0004, 30.0 * abs(relief) / ringSize, ringSize, noiseTex)), 0.0, 1.33, noiseTex);
    let gsl_simpleFull = material_withDisplacement(water, relief * texture_fadeDetail(ringSize, noiseTex), noiseTex);
    gsl_simpleValue = mat4x4<f32>(mix(gsl_simpleValue[0], gsl_simpleFull[0], gsl_simpleDetail), mix(gsl_simpleValue[1], gsl_simpleFull[1], gsl_simpleDetail), mix(gsl_simpleValue[2], gsl_simpleFull[2], gsl_simpleDetail), mix(gsl_simpleValue[3], gsl_simpleFull[3], gsl_simpleDetail));
  }
  let pool = gsl_simpleValue;
  var gsl_result : gsl_struct;
  gsl_result.v_1 = pool;
  gsl_result.v_2 = vec3<f32>(0.05, 0.11, 0.12);
  gsl_result.v_3 = 0.0;
  gsl_result.v_4 = 0.0;
  return gsl_result;
}

fn animated_mudBubble(r: f32, phase: f32, cycle: f32, size: f32, thinness: f32, noiseTex: texture_3d<f32>) -> f32 {
  let radius = 0.5 * size;
  let crater = -0.3 * radius * max(1.0 - r / radius * r / radius, 0.0);
  let rim = 0.25 * radius * exp(-((r - radius) / (0.25 * radius)) * (r - radius) / (0.25 * radius));
  let age = max(phase - 0.8, 0.0) * cycle;
  let relax = exp(-age / (cycle * (0.02 + 0.6 * (1.0 - thinness))));
  let front = texture_mask(radius + 150.0 * age - r, 0.0, 0.5, noiseTex);
  let rings = thinness * 0.1 * radius * cos(6.2831855 * (r - 150.0 * age) / (0.6 * radius)) * exp(-age / 0.3) * front * texture_mask(r, radius, 0.2 * radius, noiseTex);
  let swell = clamp(phase / 0.8, 0.0, 1.0);
  let grown = radius * pow(swell, 0.33);
  let inside = max(1.0 - r / max(grown, 0.001) * r / max(grown, 0.001), 0.0);
  let dome = 0.7 * grown * swell * inside * inside;
  return select((crater + rim) * relax + rings, dome, phase < 0.8);
}

fn animated_mudPot_gsl_anon(p: vec3<f32>, base: vec3<f32>, cellSize: f32, period: f32, time: f32, bubbleSize: f32, thinness: f32, acc: vec2<f32>, i: f32, noiseTex: texture_3d<f32>) -> vec2<f32> {
  let cell = base + vec3<f32>(gsl_floored(i, 3.0) - 1.0, floor(i / 3.0) - 1.0, 0.0);
  let h = fract(textureLoad(noiseTex, vec3<i32>(floor(vec3<f32>(cell.x + 3.0, cell.y + 97.0, 0.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((vec3<f32>(cell.x + 3.0, cell.y + 97.0, 0.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
  let x = p.x - (cell.x + 0.33 + 0.33 * h.x) * cellSize;
  let y = p.y - (cell.y + 0.33 + 0.33 * h.y) * cellSize;
  let r = sqrt(x * x + y * y);
  let cycle = period * (0.7 + 0.6 * h.z);
  let phase = fract(time / cycle + 7.3 * h.x);
  return vec2<f32>(acc.x + animated_mudBubble(r, phase, cycle, bubbleSize * (0.6 + 0.8 * h.y), thinness, noiseTex), min(acc.y, r));
}

fn animated_mudPot(p: vec3<f32>, time: f32, bubbleSize: f32, period: f32, thinness: f32, noiseTex: texture_3d<f32>) -> gsl_struct {
  let cellSize = 3.0 * bubbleSize;
  let base = floor(vec3<f32>(p.x / cellSize, p.y / cellSize, 0.0));
  var gsl_fold = vec2<f32>(0.0, 1000.0);
  for (var gsl_i = 0; gsl_i < 9; gsl_i = gsl_i + 1) {
    gsl_fold = animated_mudPot_gsl_anon(p, base, cellSize, period, time, bubbleSize, thinness, gsl_fold, f32(gsl_i), noiseTex);
  }
  let found = gsl_fold;
  let dry = texture_mask(found.y / cellSize, 0.55, 0.15, noiseTex);
  var gsl_simpleValue = material_withDisplacement(mat4x4<f32>(vec4<f32>(0.25855, 0.0, 1.4167, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), 0.003 * bubbleSize, noiseTex);
  if (gsl_simpleDetail > 0.0) {
    let gsl_simpleFull = material_withDisplacement(material_blend(mat4x4<f32>(vec4<f32>(0.15, 0.0, 1.4, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), mat4x4<f32>(vec4<f32>(0.8, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), dry, noiseTex), found.x, noiseTex);
    gsl_simpleValue = mat4x4<f32>(mix(gsl_simpleValue[0], gsl_simpleFull[0], gsl_simpleDetail), mix(gsl_simpleValue[1], gsl_simpleFull[1], gsl_simpleDetail), mix(gsl_simpleValue[2], gsl_simpleFull[2], gsl_simpleDetail), mix(gsl_simpleValue[3], gsl_simpleFull[3], gsl_simpleDetail));
  }
  let pot = gsl_simpleValue;
  var gsl_simpleValue_2 = vec3<f32>(0.34137827, 0.33132967, 0.309709);
  if (gsl_simpleDetail > 0.0) {
    gsl_simpleValue_2 = mix(gsl_simpleValue_2, color_mixOklab(vec3<f32>(0.3, 0.29, 0.27), vec3<f32>(0.56, 0.55, 0.52), dry, noiseTex), gsl_simpleDetail);
  }
  let col = gsl_simpleValue_2;
  var gsl_result : gsl_struct;
  gsl_result.v_1 = pot;
  gsl_result.v_2 = col;
  gsl_result.v_3 = 0.0;
  gsl_result.v_4 = 0.0;
  return gsl_result;
}

fn gsl_shapeParts(p: vec3<f32>, time: f32, gsl_param_10_12: f32, gsl_param_16_12: f32, noiseTex: texture_3d<f32>) -> array<f32, 2> {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let rainfall = gsl_param_10_12;
  let gsl_result = animated_rain(p, time, rainfall, 2.0, noiseTex);
  let rainHeight = gsl_result.v_3;
  let gsl_partLo = vec3<f32>(-28.0 - max(rainHeight, 0.0), -12.0 - max(rainHeight, 0.0), -4.0 - max(rainHeight, 0.0));
  let gsl_partHi = vec3<f32>(-4.0 + max(rainHeight, 0.0), 12.0 + max(rainHeight, 0.0), 4.0 + max(rainHeight, 0.0));
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = builtin_gcl_closure_boxCentered(p - vec3<f32>(-16.0, -0.0, 0.0), vec3<f32>(0.0, 0.0, 0.0) * 0.5, abs(vec3<f32>(24.0, 24.0, 8.0)), noiseTex) - rainHeight + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let rain = gsl_part;
  let thinness = gsl_param_16_12;
  let gsl_result_2 = animated_mudPot(p, time, 6.0, 2.0, thinness, noiseTex);
  let mudHeight = gsl_result_2.v_3;
  let gsl_partLo_2 = vec3<f32>(4.0 - max(mudHeight, 0.0), -12.0 - max(mudHeight, 0.0), -4.0 - max(mudHeight, 0.0));
  let gsl_partHi_2 = vec3<f32>(28.0 + max(mudHeight, 0.0), 12.0 + max(mudHeight, 0.0), 4.0 + max(mudHeight, 0.0));
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = builtin_gcl_closure_boxCentered(p - vec3<f32>(16.0, -0.0, 0.0), vec3<f32>(0.0, 0.0, 0.0) * 0.5, abs(vec3<f32>(24.0, 24.0, 8.0)), noiseTex) - mudHeight + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let mud = gsl_part_2;
  return array<f32, 2>(rain, mud);
}

fn sdShape(p: vec3<f32>, time: f32, gsl_param_10_12: f32, gsl_param_16_12: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  let gsl_d = gsl_shapeParts(p, time, gsl_param_10_12, gsl_param_16_12, noiseTex);
  let gsl_s1 = gsl_shape(gsl_d[0], 1u);
  let gsl_s2 = gsl_shape(gsl_d[1], 2u);
  return gsl_min(gsl_s1, gsl_s2);
}
fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32, gsl_param_10_12: f32, gsl_param_16_12: f32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  if (gsl_colorIdx == 1u) {
    let rainfall = gsl_param_10_12;
    let gsl_result = animated_rain(p, time, rainfall, 2.0, noiseTex);
    let rainColor = gsl_result.v_2;
    return vec4<f32>(rainColor, 1.0);
  }
  if (gsl_colorIdx == 2u) {
    let thinness = gsl_param_16_12;
    let gsl_result_2 = animated_mudPot(p, time, 6.0, 2.0, thinness, noiseTex);
    let mudColor = gsl_result_2.v_2;
    return vec4<f32>(mudColor, 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32, gsl_param_10_12: f32, gsl_param_16_12: f32, noiseTex: texture_3d<f32>) -> gsl_material {
  if (gsl_materialIdx == 1u) {
    let rainfall = gsl_param_10_12;
    let gsl_result = animated_rain(p, time, rainfall, 2.0, noiseTex);
    let rainMat = gsl_result.v_1;
    let gsl_m = rainMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 2u) {
    let thinness = gsl_param_16_12;
    let gsl_result_2 = animated_mudPot(p, time, 6.0, 2.0, thinness, noiseTex);
    let mudMat = gsl_result_2.v_1;
    let gsl_m = mudMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
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

fn gsl_heights(p: vec3<f32>, time: f32, gsl_param_10_12: f32, gsl_param_16_12: f32, noiseTex: texture_3d<f32>) -> array<f32, 2> {
  let rainfall = gsl_param_10_12;
  let gsl_result = animated_rain(p, time, rainfall, 2.0, noiseTex);
  let rainMat = gsl_result.v_1;
  let thinness = gsl_param_16_12;
  let gsl_result_2 = animated_mudPot(p, time, 6.0, 2.0, thinness, noiseTex);
  let mudMat = gsl_result_2.v_1;
  return array<f32, 2>(rainMat[2].z, mudMat[2].z);
}

fn gsl_sdShapeDisplaced(p: vec3<f32>, time: f32, gsl_dispK: f32, gsl_param_10_12: f32, gsl_param_16_12: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  let gsl_d = gsl_shapeParts(p, time, gsl_param_10_12, gsl_param_16_12, noiseTex);
  let gsl_h = gsl_heights(p, time, gsl_param_10_12, gsl_param_16_12, noiseTex);
  let gsl_s1 = gsl_shape(gsl_d[0] - gsl_h[0] * gsl_dispK, 1u);
  let gsl_s2 = gsl_shape(gsl_d[1] - gsl_h[1] * gsl_dispK, 2u);
  return gsl_min(gsl_s1, gsl_s2);
}

struct gsl_struct {
  v_1 : mat4x4<f32>,
  v_2 : vec3<f32>,
  v_3 : f32,
  v_4 : f32,
}

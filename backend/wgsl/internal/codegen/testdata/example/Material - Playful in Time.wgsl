fn gsl_floored_vec2(a: vec2<f32>, n: vec2<f32>) -> vec2<f32> {
  return a - n * floor(a / n);
}

fn ease_quadratic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * (3.0 - 2.0 * t);
}

fn texture_mask(field: f32, edge: f32, width: f32, noiseTex: texture_3d<f32>) -> f32 {
  let t = clamp((field - edge) / width * 0.5 + 0.5, 0.0, 1.0);
  return ease_quadratic(t, noiseTex);
}

fn texture_coverageDisc(gsl_var_distance: f32, radius: f32, noiseTex: texture_3d<f32>) -> f32 {
  let drawn = max(radius, 0.5 * gsl_pw);
  let soft = max(0.5 * gsl_pw, 0.01 * radius);
  return (1.0 - texture_mask(gsl_var_distance, drawn, soft, noiseTex)) * radius / drawn * radius / drawn;
}

fn texture_fadeDetail(size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 1.0 - texture_mask(gsl_pw, 0.375 * size, 0.125 * size, noiseTex);
}

fn texture_fadeTo(gsl_var_value: f32, mean: f32, size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return mean + (gsl_var_value - mean) * texture_fadeDetail(size, noiseTex);
}

fn map_repeatCentered(x: vec2<f32>, period: vec2<f32>, noiseTex: texture_3d<f32>) -> vec2<f32> {
  let half = period * 0.5;
  return gsl_floored_vec2((gsl_floored_vec2((x + half), period) + period), period) - half;
}

fn gsl_remap(p: vec3<f32>, material_polishedMetal: mat4x4<f32>, color_white: vec3<f32>, time: f32, color_black: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  var gsl_simpleValue = vec3<f32>(0.6604224, 0.6490635, 0.6588749);
  if (gsl_simpleDetail > 0.0) {
    gsl_simpleValue = mix(gsl_simpleValue, color_cyclicRainbow(fract(p.z / 10.0 - time / 5.0), noiseTex), gsl_simpleDetail);
  }
  let rainbow = gsl_simpleValue;
  let gsl_tmp_61 = color_rgb2oklab(rainbow, noiseTex);
  let gsl_tmp_57 = map_repeatCentered(p.xy - vec2<f32>(5.196152, 15.0), vec2<f32>(10.392304, 18.0), noiseTex);
  let gsl_tmp_56 = map_repeatCentered(p.xy - vec2<f32>(5.196152, 9.0), vec2<f32>(10.392304, 18.0), noiseTex);
  let gsl_tmp_59 = select(gsl_tmp_57, gsl_tmp_56, dot(gsl_tmp_56, gsl_tmp_56) < dot(gsl_tmp_57, gsl_tmp_57));
  let gsl_tmp_55 = map_repeatCentered(p.xy - vec2<f32>(0.0, 6.0), vec2<f32>(10.392304, 18.0), noiseTex);
  let gsl_tmp_54 = map_repeatCentered(p.xy, vec2<f32>(10.392304, 18.0), noiseTex);
  let gsl_tmp_58 = select(gsl_tmp_55, gsl_tmp_54, dot(gsl_tmp_54, gsl_tmp_54) < dot(gsl_tmp_55, gsl_tmp_55));
  let gsl_tmp_60 = select(gsl_tmp_59, gsl_tmp_58, dot(gsl_tmp_58, gsl_tmp_58) < dot(gsl_tmp_59, gsl_tmp_59));
  let gsl_tmp_62 = gsl_tmp_61 + (vec3<f32>(1.0, 0.00000000008095291, 0.000000037273907) - gsl_tmp_61) * texture_fadeTo(texture_coverageDisc(length(vec3<f32>(gsl_tmp_60.x, gsl_tmp_60.y, p.z).xy), 2.0, noiseTex), 0.26871103, 6.0, noiseTex);
  let gsl_tmp_63 = gsl_tmp_62.x + 0.39633778 * gsl_tmp_62.y + 0.21580376 * gsl_tmp_62.z;
  let gsl_tmp_64 = gsl_tmp_62.x - 0.105561346 * gsl_tmp_62.y - 0.06385417 * gsl_tmp_62.z;
  let gsl_tmp_65 = gsl_tmp_62.x - 0.08948418 * gsl_tmp_62.y - 1.2914855 * gsl_tmp_62.z;
  let gsl_tmp_66 = vec3<f32>(gsl_tmp_63 * gsl_tmp_63 * gsl_tmp_63, gsl_tmp_64 * gsl_tmp_64 * gsl_tmp_64, gsl_tmp_65 * gsl_tmp_65 * gsl_tmp_65);
  let gsl_tmp_73 = vec3<f32>(dot(vec3<f32>(4.0767417, -3.3077116, 0.23096994), gsl_tmp_66), dot(vec3<f32>(-1.268438, 2.6097574, -0.34131938), gsl_tmp_66), dot(vec3<f32>(-0.0041960864, -0.7034186, 1.7076147), gsl_tmp_66));
  return vec3<f32>(color_linearToSrgbChannel(gsl_tmp_73.x, noiseTex), color_linearToSrgbChannel(gsl_tmp_73.y, noiseTex), color_linearToSrgbChannel(gsl_tmp_73.z, noiseTex));
}

fn gsl_remap_2(p: vec3<f32>, material_polishedMetal: mat4x4<f32>, color_white: vec3<f32>, time: f32, color_black: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  var gsl_simpleValue_2 = vec3<f32>(0.6604224, 0.6490635, 0.6588749);
  if (gsl_simpleDetail > 0.0) {
    gsl_simpleValue_2 = mix(gsl_simpleValue_2, color_cyclicRainbow(fract(p.z / 10.0 - time / 5.0), noiseTex), gsl_simpleDetail);
  }
  let rainbow = gsl_simpleValue_2;
  var gsl_simpleValue_3 = color_mixOklab(rainbow, color_white, 0.26871103, noiseTex);
  if (gsl_simpleDetail > 0.0) {
    gsl_simpleValue_3 = mix(gsl_simpleValue_3, gsl_remap(vec3<f32>(dot(p, vec3<f32>(1.0, 0.0, 0.0)), dot(p, vec3<f32>(0.0, 1.0, 0.0)), dot(p, vec3<f32>(0.0, 0.0, 1.0))), material_polishedMetal, color_white, time, color_black, noiseTex), gsl_simpleDetail);
  }
  return gsl_simpleValue_3;
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

fn color_sampleMap(palette: array<vec3<f32>, 7>, t: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let tt = clamp(t, 0.0, 1.0);
  let weighted = array<vec3<f32>, 7>(palette[0] * max(0.0, 1.0 - abs(tt * 6.0 - 0.0)), palette[1] * max(0.0, 1.0 - abs(tt * 6.0 - 1.0)), palette[2] * max(0.0, 1.0 - abs(tt * 6.0 - 2.0)), palette[3] * max(0.0, 1.0 - abs(tt * 6.0 - 3.0)), palette[4] * max(0.0, 1.0 - abs(tt * 6.0 - 4.0)), palette[5] * max(0.0, 1.0 - abs(tt * 6.0 - 5.0)), palette[6] * max(0.0, 1.0 - abs(tt * 6.0 - 6.0)));
  return weighted[0] + weighted[1] + weighted[2] + weighted[3] + weighted[4] + weighted[5] + weighted[6];
}

fn color_cyclicRainbow(t: f32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return color_sampleMap(array<vec3<f32>, 7>(vec3<f32>(1.0, 0.0, 0.0), vec3<f32>(1.0, 1.0, 0.0), vec3<f32>(0.0, 1.0, 0.0), vec3<f32>(0.0, 1.0, 1.0), vec3<f32>(0.0, 0.0, 1.0), vec3<f32>(1.0, 0.0, 1.0), vec3<f32>(1.0, 0.0, 0.0)), t, noiseTex);
}

fn material_blend(a: mat4x4<f32>, b: mat4x4<f32>, t: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return mat4x4<f32>(a[0] + (b[0] - a[0]) * t, a[1] + (b[1] - a[1]) * t, a[2] + (b[2] - a[2]) * t, a[3] + (b[3] - a[3]) * t);
}

fn material_withEmissive(m: mat4x4<f32>, rgb: vec3<f32>, strength: f32, noiseTex: texture_3d<f32>) -> mat4x4<f32> {
  return mat4x4<f32>(m[0], vec4<f32>(rgb.x, rgb.y, rgb.z, strength), m[2], m[3]);
}

fn playful_glitterShare(density: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 1.0 - exp(-4.18879 * density * density * density);
}

fn playful_glitterFlakes(f1: f32, flakeCellSize: f32, density: f32, noiseTex: texture_3d<f32>) -> f32 {
  return texture_fadeTo(texture_coverageDisc(f1, density * flakeCellSize, noiseTex), playful_glitterShare(density, noiseTex), flakeCellSize, noiseTex);
}

fn gsl_ball_remap(p: vec3<f32>, noiseTex: texture_3d<f32>) -> f32 {
  return length(p) - 12.0;
}

fn sdShape(p: vec3<f32>, time: f32, noiseTex: texture_3d<f32>) -> gsl_shape {
  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);
  let gsl_partLo = vec3<f32>(-28.0, -12.0, -12.0);
  let gsl_partHi = vec3<f32>(-4.0, 12.0, 12.0);
  let gsl_partBox = gsl_boxDistance(p, gsl_partLo, gsl_partHi);
  var gsl_part = gsl_partBox;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox <= gsl_vanishView.w) {
    let gsl_partVanish = gsl_vanish(gsl_partLo, gsl_partHi);
    gsl_part = 1e10;
    if (gsl_partVanish.x > 0.0) { gsl_part = gsl_ball_remap(p - vec3<f32>(-16.0, -0.0, 0.0), noiseTex) + gsl_partVanish.y; }
    if (abs(gsl_part) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part), gsl_partVanish.zw); }
  }
  let glitter = gsl_part;
  let gsl_partLo_2 = vec3<f32>(4.0, -12.0, -12.0);
  let gsl_partHi_2 = vec3<f32>(28.0, 12.0, 12.0);
  let gsl_partBox_2 = gsl_boxDistance(p, gsl_partLo_2, gsl_partHi_2);
  var gsl_part_2 = gsl_partBox_2;
  if (gsl_vanishView.w <= 0.0 || gsl_partBox_2 <= gsl_vanishView.w) {
    let gsl_partVanish_2 = gsl_vanish(gsl_partLo_2, gsl_partHi_2);
    gsl_part_2 = 1e10;
    if (gsl_partVanish_2.x > 0.0) { gsl_part_2 = gsl_ball_remap(p - vec3<f32>(16.0, -0.0, 0.0), noiseTex) + gsl_partVanish_2.y; }
    if (abs(gsl_part_2) < gsl_partNear.x) { gsl_partNear = vec3<f32>(abs(gsl_part_2), gsl_partVanish_2.zw); }
  }
  let dots = gsl_part_2;
  let gsl_s1 = gsl_shape(glitter, 1u);
  let gsl_s2 = gsl_shape(dots, 2u);
  return gsl_min(gsl_s1, gsl_s2);
}

fn gsl_colorField(p: vec3<f32>, time: f32, gsl_colorIdx: u32, noiseTex: texture_3d<f32>) -> vec4<f32> {
  if (gsl_colorIdx == 1u) {
    var gsl_simpleValue = vec3<f32>(0.0032853691, 0.003285369, 0.003285368);
    if (gsl_simpleDetail > 0.0) {
      let color_black = vec3<f32>(0.0, 0.0, 0.0);
      let color_white = vec3<f32>(1.0, 1.0, 1.0);
      let gsl_tmp_1 = p / 1.0;
      let gsl_tmp_2 = playful_glitterFlakes(gsl_cellularScan(gsl_tmp_1, 1, noiseTex).f1 * 1.0, 1.0, 0.25, noiseTex);
      gsl_simpleValue = mix(gsl_simpleValue, color_mixOklab(color_black, color_white, gsl_tmp_2, noiseTex), gsl_simpleDetail);
    }
    let glitterColor = gsl_simpleValue;
    return vec4<f32>(glitterColor, 1.0);
  }
  if (gsl_colorIdx == 2u) {
    var gsl_simpleValue_2 = vec3<f32>(0.6604224, 0.6490635, 0.6588749);
    if (gsl_simpleDetail > 0.0) {
      gsl_simpleValue_2 = mix(gsl_simpleValue_2, color_cyclicRainbow(fract(p.z / 10.0 - time / 5.0), noiseTex), gsl_simpleDetail);
    }
    let rainbow = gsl_simpleValue_2;
    let color_white = vec3<f32>(1.0, 1.0, 1.0);
    var gsl_simpleValue_3 = color_mixOklab(rainbow, color_white, 0.26871103, noiseTex);
    if (gsl_simpleDetail > 0.0) {
      let material_polishedMetal = mat4x4<f32>(vec4<f32>(0.15, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
      let color_black = vec3<f32>(0.0, 0.0, 0.0);
      gsl_simpleValue_3 = mix(gsl_simpleValue_3, gsl_remap_2(p - vec3<f32>(0.0, 0.0, 0.0) - vec3<f32>(2.0, 0.0, 0.0) * time, material_polishedMetal, color_white, time, color_black, noiseTex), gsl_simpleDetail);
    }
    return vec4<f32>(gsl_simpleValue_3, 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32, noiseTex: texture_3d<f32>) -> gsl_material {
  if (gsl_materialIdx == 1u) {
    let material_polishedMetal = mat4x4<f32>(vec4<f32>(0.15, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    var gsl_simpleValue_4 = mat4x4<f32>(material_polishedMetal[0] + (vec4<f32>(0.08, 1.0, 1.5, 0.0) - material_polishedMetal[0]) * 0.06335398, vec4<f32>(1.0, 1.0, 1.0, 0.0012607756), material_polishedMetal[2] + (vec4<f32>(0.0, 0.0, 0.0, 0.0) - material_polishedMetal[2]) * 0.06335398, material_polishedMetal[3] + (vec4<f32>(0.0, 0.0, 0.0, 0.0) - material_polishedMetal[3]) * 0.06335398);
    if (gsl_simpleDetail > 0.0) {
      let gsl_tmp_1 = p / 1.0;
      let gsl_tmp_2 = playful_glitterFlakes(gsl_cellularScan(gsl_tmp_1, 1, noiseTex).f1 * 1.0, 1.0, 0.25, noiseTex);
      let color_white = vec3<f32>(1.0, 1.0, 1.0);
      let gsl_simpleFull = material_withEmissive(material_blend(material_polishedMetal, mat4x4<f32>(vec4<f32>(0.08, 1.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)), gsl_tmp_2, noiseTex), color_white, 8.0 * texture_fadeTo(pow(max(dot(normalize(fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_cellularScan(gsl_tmp_1, 1, noiseTex).nearCell + vec3<f32>(17.0, 5.0, 3.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((gsl_cellularScan(gsl_tmp_1, 1, noiseTex).nearCell + vec3<f32>(17.0, 5.0, 3.0)) / 64.0)) & vec3<i32>(63), 0).rgb) * 2.0 - 1.0), normalize(vec3<f32>(sin(0.7 * time), cos(0.49 * time), sin(0.91 * time + 1.0)))), 0.0), 200.0), 0.0024875621, 1.0, noiseTex) * gsl_tmp_2, noiseTex);
      gsl_simpleValue_4 = mat4x4<f32>(mix(gsl_simpleValue_4[0], gsl_simpleFull[0], gsl_simpleDetail), mix(gsl_simpleValue_4[1], gsl_simpleFull[1], gsl_simpleDetail), mix(gsl_simpleValue_4[2], gsl_simpleFull[2], gsl_simpleDetail), mix(gsl_simpleValue_4[3], gsl_simpleFull[3], gsl_simpleDetail));
    }
    let glitterMat = gsl_simpleValue_4;
    let gsl_m = glitterMat; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  if (gsl_materialIdx == 2u) {
    let material_plastic = mat4x4<f32>(vec4<f32>(0.35, 0.0, 1.46, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0));
    let gsl_m = material_plastic; return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);
  }
  return gsl_material(vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), 0.0);
}
struct gsl_CellularScan {
  f1: f32,
  f2: f32,
  feature: vec3<f32>,
  farFeature: vec3<f32>,
  nearCell: vec3<f32>,
  farCell: vec3<f32>,
}

fn gsl_cellularHash(cell: vec3<f32>, oct: i32, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let fine = fract(textureLoad(noiseTex, vec3<i32>(floor(cell)) & vec3<i32>(63), 0).rgb);
  if (oct <= 1) { return fine; }
  // fract(fract(a) + b) == fract(a + b): the inner fract only removes an
  // integer, so folding the supercell on top of the already-fracted fine
  // lookup is the same value noiseHash3's two-lookup form produces.
  return fract(fine + textureLoad(noiseTex, vec3<i32>(floor(cell / 64.0)) & vec3<i32>(63), 0).rgb);
}

fn gsl_cellularScan(q: vec3<f32>, oct: i32, noiseTex: texture_3d<f32>) -> gsl_CellularScan {
  let pi = floor(q);
  let pf = q - pi;
  // Seeded with an unreachable distance, squared, because the scan compares
  // squared distances and roots the two winners at the end.
  var best = 64.000000;
  var second = 64.000000;
  var bestCell = pi;
  var bestH = vec3<f32>(0.0);
  var secondCell = pi;
  var secondH = vec3<f32>(0.0);
  for (var i = 0; i < 27; i = i + 1) {
    // The same offset order cellularOffsets() lists: x fastest, then y, then z.
    let o = vec3<f32>(f32(i % 3) - 1.0, f32((i / 3) % 3) - 1.0, f32(i / 9) - 1.0);
    let h = gsl_cellularHash(pi + o, oct, noiseTex);
    let v = pf - (o + h);
    let d2 = dot(v, v);
    if (d2 < best) {
      second = best;
      secondCell = bestCell;
      secondH = bestH;
      best = d2;
      bestCell = pi + o;
      bestH = h;
    } else if (d2 < second) {
      second = d2;
      secondCell = pi + o;
      secondH = h;
    }
  }
  var out: gsl_CellularScan;
  out.f1 = sqrt(best);
  out.f2 = sqrt(second);
  // Both closest points, so both gradients are available without a second hash.
  out.feature = bestCell + bestH;
  out.farFeature = secondCell + secondH;
  out.nearCell = bestCell;
  out.farCell = secondCell;
  return out;
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

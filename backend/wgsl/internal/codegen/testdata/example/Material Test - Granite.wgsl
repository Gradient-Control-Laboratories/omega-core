fn ease_cubic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * t * (t * (t * 6.0 - 15.0) + 10.0);
}

fn texture_mask(field: f32, edge: f32, width: f32, noiseTex: texture_3d<f32>) -> f32 {
  let t = clamp((field - edge) / width * 0.5 + 0.5, 0.0, 1.0);
  return ease_quadratic(t, noiseTex);
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

fn texture_fadeDetail(size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return 1.0 - texture_mask(gsl_pw, 0.375 * size, 0.125 * size, noiseTex);
}

fn texture_fadeTo(gsl_var_value: f32, mean: f32, size: f32, noiseTex: texture_3d<f32>) -> f32 {
  return mean + (gsl_var_value - mean) * texture_fadeDetail(size, noiseTex);
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

fn ease_quadratic(t: f32, noiseTex: texture_3d<f32>) -> f32 {
  return t * t * (3.0 - 2.0 * t);
}

fn noise_cellHash3(cell: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return fract(textureLoad(noiseTex, vec3<i32>(floor(cell)) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((cell) / 64.0)) & vec3<i32>(63), 0).rgb);
}

fn noise_cornerGradient(cell: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  return normalize(noise_cellHash3(cell, noiseTex) * 2.0 - 1.0);
}

fn noise_gradientLatticeVector_gsl_anon(pi: vec3<f32>, pf: vec3<f32>, e: vec3<f32>, acc: vec3<f32>, o: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let weight = (o.x * e.x + (1.0 - o.x) * (1.0 - e.x)) * (o.y * e.y + (1.0 - o.y) * (1.0 - e.y)) * (o.z * e.z + (1.0 - o.z) * (1.0 - e.z));
  let g = noise_cornerGradient(pi + o, noiseTex);
  let d = pf - o;
  return acc + weight * vec3<f32>(dot(g, d), dot(vec3<f32>(g.y, g.z, g.x), d), dot(vec3<f32>(g.z, g.x, g.y), d));
}

fn noise_gradientLatticeVector(q: vec3<f32>, noiseTex: texture_3d<f32>) -> vec3<f32> {
  let pi = floor(q);
  let pf = q - pi;
  let e = vec3<f32>(ease_cubic(pf.x, noiseTex), ease_cubic(pf.y, noiseTex), ease_cubic(pf.z, noiseTex));
  var gsl_fold = vec3<f32>(0.0, 0.0, 0.0);
  for (var gsl_i = 0; gsl_i < 8; gsl_i = gsl_i + 1) {
    gsl_fold = noise_gradientLatticeVector_gsl_anon(pi, pf, e, gsl_fold, gsl_foldK[gsl_i], noiseTex);
  }
  return gsl_fold;
}

fn stone_grainOf(c: vec3<f32>, size: f32, s: f32, zone: f32, zoneWeight: f32, clump: f32, seed: vec3<f32>, scatter: f32, level: f32, noiseTex: texture_3d<f32>) -> f32 {
  let at = (c + fract(textureLoad(noiseTex, vec3<i32>(floor(c)) & vec3<i32>(63), 0).rgb)) * size;
  let cluster = zoneWeight * noise_gradientLattice(at / (zone * s) + seed, noiseTex) + 0.6 * noise_gradientLattice(at / (clump * s) + seed * 1.7, noiseTex);
  return select(0.0, 1.0, cluster - scatter * fract(textureLoad(noiseTex, vec3<i32>(floor(c + seed)) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((c + seed) / 64.0)) & vec3<i32>(63), 0).rgb).x > level);
}

fn stone_quartzOf(c: vec3<f32>, g: f32, s: f32, noiseTex: texture_3d<f32>) -> f32 {
  return stone_grainOf(c, g, s, 4.0, 0.75, 1.2, vec3<f32>(7.3, 3.1, 11.7), 0.8, -0.31, noiseTex);
}

fn stone_maficOf(c: vec3<f32>, t: f32, s: f32, nearQuartz: f32, noiseTex: texture_3d<f32>) -> f32 {
  return stone_grainOf(c, t, s, 4.0, 0.25, 0.6, vec3<f32>(13.1, 7.7, 2.3), 0.25, 0.17 - 0.15 * nearQuartz, noiseTex);
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
  if (gsl_colorIdx == 1u) {
    var gsl_detailValue = vec3<f32>(0.55051714, 0.5512089, 0.5521506);
    if (max(gsl_pw, gsl_detailPw) < 1.2) { let gsl_detailOuterPw = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
      let gsl_tmp_1 = p + (noise_gradientLatticeVector(p / 5.2, noiseTex) + noise_gradientLatticeVector(p * 2.03 / 5.2 + vec3<f32>(17.3, 5.1, 9.7), noiseTex) * 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.975) / 0.325 * 0.5 + 0.5, 0.0, 1.0), noiseTex))) * 1.056;
      let gsl_tmp_2 = gsl_tmp_1 / 4.0;
      let gsl_tmp_5 = gsl_cellularScan(gsl_tmp_2, 1, noiseTex).nearCell;
      let gsl_tmp_25 = fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_5 + vec3<f32>(101.0, 0.0, 0.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((gsl_tmp_5 + vec3<f32>(101.0, 0.0, 0.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
      let gsl_tmp_29 = noise_gradientLattice(p / 1.2 + vec3<f32>(1.7, 9.1, 4.4), noiseTex) + 0.15 * (1.0 - texture_mask(gsl_pw, 0.18, 0.06, noiseTex)) * noise_gradientLattice(p / 0.48 + vec3<f32>(8.2, 2.6, 7.3), noiseTex) + 0.4 * noise_gradientLattice(p / 3.6 + vec3<f32>(4.2, 6.6, 1.3), noiseTex);
      let gsl_tmp_27 = normalize(gsl_tmp_25 - 0.5 + vec3<f32>(0.0001, 0.0, 0.0));
      let gsl_tmp_28 = dot(p, gsl_tmp_27);
      let gsl_tmp_4 = gsl_cellularScan(gsl_tmp_2, 1, noiseTex).f2;
      let gsl_tmp_3 = gsl_cellularScan(gsl_tmp_2, 1, noiseTex).f1;
      let gsl_tmp_6 = gsl_cellularScan(gsl_tmp_2, 1, noiseTex).farCell;
      let gsl_tmp_31 = 1.0 - texture_mask((gsl_tmp_4 * gsl_tmp_4 - gsl_tmp_3 * gsl_tmp_3) / (2.0 * max(length(gsl_tmp_6 + fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_6)) & vec3<i32>(63), 0).rgb) - gsl_tmp_5 - fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_5)) & vec3<i32>(63), 0).rgb)), 0.001)) * 4.0, 0.24, 0.2, noiseTex);
      let gsl_tmp_30 = texture_fadeTo(abs(noise_gradientLattice(p / 0.6 + vec3<f32>(6.1, 4.9, 2.2), noiseTex)), 0.157, 0.6, noiseTex) + 0.5 * texture_fadeTo(abs(noise_gradientLattice(p / 0.24 + vec3<f32>(1.1, 3.5, 8.8), noiseTex)), 0.157, 0.24, noiseTex);
      let gsl_tmp_32 = vec3<f32>(0.5981807, 0.000000000048424487, 0.000000022296533) + vec3<f32>((gsl_tmp_25.y - 0.5) * 0.1 + 0.1 * gsl_tmp_29 + 0.015 * noise_gradientLattice((p - gsl_tmp_27 * gsl_tmp_28) / 0.32 + gsl_tmp_27 * gsl_tmp_28 / 3.2, noiseTex) * (1.0 - texture_mask(gsl_pw, 0.12, 0.04, noiseTex)) + 0.04 * gsl_tmp_31 + 0.12 * (gsl_tmp_30 - 0.235) * (0.5 + gsl_tmp_31), 0.0, 0.0) - vec3<f32>(0.0, 0.000000000048424487, 0.000000022296533) * 0.3 * gsl_tmp_31;
      let gsl_tmp_7 = (gsl_tmp_1 + p + (noise_gradientLatticeVector(p / 2.0, noiseTex) + noise_gradientLatticeVector(p * 2.03 / 2.0 + vec3<f32>(17.3, 5.1, 9.7), noiseTex) * 0.5 * (1.0 - ease_quadratic(clamp((gsl_pw - 0.375) / 0.125 * 0.5 + 0.5, 0.0, 1.0), noiseTex))) * 0.384 - p + vec3<f32>(148.4, 45.2, 23.6)) / 1.8;
      let gsl_tmp_10 = gsl_cellularScan(gsl_tmp_7, 1, noiseTex).nearCell;
      let gsl_tmp_13 = fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_10 + vec3<f32>(7.0, 29.0, 0.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((gsl_tmp_10 + vec3<f32>(7.0, 29.0, 0.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
      let gsl_tmp_15 = stone_quartzOf(gsl_tmp_10, 1.8, 4.0, noiseTex);
      let gsl_tmp_24 = select(0.0, 1.0 - gsl_tmp_15, gsl_tmp_13.x < 0.15);
      let gsl_tmp_34 = gsl_tmp_32 + (vec3<f32>(0.5981807, 0.000000000048424487, 0.000000022296533) + vec3<f32>(0.06 + (gsl_tmp_13.y - 0.5) * 0.06 + 0.04 * gsl_tmp_29 + 0.12 * (gsl_tmp_30 - 0.235), -0.000000000024212243, -0.000000011148266) - gsl_tmp_32) * gsl_tmp_24;
      let gsl_tmp_8 = gsl_cellularScan(gsl_tmp_7, 1, noiseTex).f1;
      let gsl_tmp_33 = 1.0 - texture_mask(gsl_tmp_8, 0.3, 0.25, noiseTex);
      let gsl_tmp_9 = gsl_cellularScan(gsl_tmp_7, 1, noiseTex).f2;
      let gsl_tmp_11 = gsl_cellularScan(gsl_tmp_7, 1, noiseTex).farCell;
      let gsl_tmp_12 = (gsl_tmp_9 * gsl_tmp_9 - gsl_tmp_8 * gsl_tmp_8) / (2.0 * max(length(gsl_tmp_11 + fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_11)) & vec3<i32>(63), 0).rgb) - gsl_tmp_10 - fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_10)) & vec3<i32>(63), 0).rgb)), 0.001));
      let gsl_tmp_14 = 2.0 * texture_mask(gsl_tmp_12 * 1.8, 0.0, 0.2, noiseTex) - 1.0;
      let gsl_tmp_22 = gsl_tmp_15 * (0.5 + 0.5 * gsl_tmp_14) + stone_quartzOf(gsl_tmp_11, 1.8, 4.0, noiseTex) * (0.5 - 0.5 * gsl_tmp_14);
      let gsl_tmp_35 = gsl_tmp_34 + (vec3<f32>(0.775, 0.00000000006476233, 0.000000029819127) + vec3<f32>((gsl_tmp_13.z - 0.5) * 0.12 - 0.1 * gsl_tmp_33 + 0.05 * dot(gsl_tmp_7 - gsl_tmp_10 - fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_10)) & vec3<i32>(63), 0).rgb), normalize(gsl_tmp_13 - 0.5 + vec3<f32>(0.0, 0.0001, 0.0))) + 0.01 * (1.0 - texture_mask(gsl_tmp_12 * 1.8, 0.08, 0.06, noiseTex)) * (1.0 - texture_mask(gsl_pw, 0.075, 0.025, noiseTex)) + 0.02 * gsl_tmp_29 + 0.06 * (gsl_tmp_30 - 0.235), -0.003 * gsl_tmp_33, -0.01 * gsl_tmp_33 + (gsl_tmp_13.y - 0.5) * 0.02) - gsl_tmp_34) * gsl_tmp_22 + vec3<f32>(0.15 * noise_gradientLattice(p / 12.0 + vec3<f32>(2.9, 8.3, 6.1), noiseTex) + 0.08 * noise_gradientLattice(p / 5.6 + vec3<f32>(9.4, 1.2, 3.8), noiseTex), 0.0, 0.0);
      let gsl_tmp_16 = (gsl_tmp_1 + vec3<f32>(21.2, 246.8, 68.4)) / 0.6;
      let gsl_tmp_19 = gsl_cellularScan(gsl_tmp_16, 1, noiseTex).nearCell;
      let gsl_tmp_26 = fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_19 + vec3<f32>(3.0, 0.0, 41.0))) & vec3<i32>(63), 0).rgb + textureLoad(noiseTex, vec3<i32>(floor((gsl_tmp_19 + vec3<f32>(3.0, 0.0, 41.0)) / 64.0)) & vec3<i32>(63), 0).rgb);
      let gsl_tmp_18 = gsl_cellularScan(gsl_tmp_16, 1, noiseTex).f2;
      let gsl_tmp_17 = gsl_cellularScan(gsl_tmp_16, 1, noiseTex).f1;
      let gsl_tmp_20 = gsl_cellularScan(gsl_tmp_16, 1, noiseTex).farCell;
      let gsl_tmp_21 = 2.0 * texture_mask((gsl_tmp_18 * gsl_tmp_18 - gsl_tmp_17 * gsl_tmp_17) / (2.0 * max(length(gsl_tmp_20 + fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_20)) & vec3<i32>(63), 0).rgb) - gsl_tmp_19 - fract(textureLoad(noiseTex, vec3<i32>(floor(gsl_tmp_19)) & vec3<i32>(63), 0).rgb)), 0.001)) * 0.6, 0.0, 0.2, noiseTex) - 1.0;
      let gsl_tmp_23 = stone_maficOf(gsl_tmp_19, 0.6, 4.0, gsl_tmp_15, noiseTex) * (0.5 + 0.5 * gsl_tmp_21) + stone_maficOf(gsl_tmp_20, 0.6, 4.0, gsl_tmp_15, noiseTex) * (0.5 - 0.5 * gsl_tmp_21);
      gsl_detailValue = color_oklab2rgb(vec3<f32>(0.6418677, -0.000099999954, -0.00039997892) + (gsl_tmp_35 + (vec3<f32>(0.30524173, 0.0, 0.0) + vec3<f32>((gsl_tmp_26.y - 0.5) * 0.06 + select(0.0, 0.12, gsl_tmp_26.z > 0.96), 0.0, 0.0) + select(vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(0.03, 0.01, 0.02), gsl_tmp_26.z < 0.08) - gsl_tmp_35) * gsl_tmp_23 - vec3<f32>(0.6418677, -0.000099999954, -0.00039997892)) * (1.0 - texture_mask(gsl_pw, 0.9, 0.3, noiseTex)), noiseTex);
      gsl_pw = gsl_detailOuterPw;
    }
    var gsl_simpleValue = vec3<f32>(0.55051714, 0.5512089, 0.5521506);
    if (gsl_simpleDetail > 0.0) {
      gsl_simpleValue = mix(gsl_simpleValue, gsl_detailValue, gsl_simpleDetail);
    }
    let graniteColor = gsl_simpleValue;
    return vec4<f32>(graniteColor, 1.0);
  }
  return vec4<f32>(0.0, 0.0, 0.0, 0.0);
}
fn gsl_materialField(p: vec3<f32>, time: f32, gsl_materialIdx: u32, noiseTex: texture_3d<f32>) -> gsl_material {
  if (gsl_materialIdx == 1u) { let gsl_m = mat4x4<f32>(vec4<f32>(0.6, 0.0, 1.55, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0)); return gsl_material(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0); }
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

var<private> gsl_detailPw: f32 = 0.0;

const gsl_foldK = array<vec3<f32>, 8>(vec3<f32>(0.0, 0.0, 0.0), vec3<f32>(1.0, 0.0, 0.0), vec3<f32>(0.0, 1.0, 0.0), vec3<f32>(1.0, 1.0, 0.0), vec3<f32>(0.0, 0.0, 1.0), vec3<f32>(1.0, 0.0, 1.0), vec3<f32>(0.0, 1.0, 1.0), vec3<f32>(1.0, 1.0, 1.0));
